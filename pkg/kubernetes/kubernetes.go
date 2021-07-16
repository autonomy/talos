// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/go-retry/retry"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/connrotation"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
)

const (
	// DrainTimeout is maximum time to wait for the node to be drained.
	DrainTimeout = 5 * time.Minute
)

// Client represents a set of helper methods for interacting with the
// Kubernetes API.
type Client struct {
	*kubernetes.Clientset

	dialer *connrotation.Dialer
}

func newDialer() *connrotation.Dialer {
	return connrotation.NewDialer((&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext)
}

// NewClientFromKubeletKubeconfig initializes and returns a Client.
func NewClientFromKubeletKubeconfig() (client *Client, err error) {
	var config *restclient.Config

	config, err = clientcmd.BuildConfigFromFlags("", constants.KubeletKubeconfig)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	dialer := newDialer()
	config.Dial = dialer.DialContext

	var clientset *kubernetes.Clientset

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset: clientset,
		dialer:    dialer,
	}, nil
}

// NewForConfig initializes and returns a client using the provided config.
func NewForConfig(config *restclient.Config) (client *Client, err error) {
	var clientset *kubernetes.Clientset

	if config.Dial != nil {
		return nil, fmt.Errorf("dialer is already set")
	}

	dialer := newDialer()
	config.Dial = dialer.DialContext

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset: clientset,
		dialer:    dialer,
	}, nil
}

// NewClientFromPKI initializes and returns a Client.
//
//nolint:interfacer
func NewClientFromPKI(ca, crt, key []byte, endpoint *url.URL) (client *Client, err error) {
	tlsClientConfig := restclient.TLSClientConfig{
		CAData:   ca,
		CertData: crt,
		KeyData:  key,
	}

	config := &restclient.Config{
		Host:            endpoint.String(),
		TLSClientConfig: tlsClientConfig,
		Timeout:         30 * time.Second,
	}

	dialer := newDialer()
	config.Dial = dialer.DialContext

	var clientset *kubernetes.Clientset

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset: clientset,
		dialer:    dialer,
	}, nil
}

// NewTemporaryClientFromPKI initializes a Kubernetes client using a certificate
// with a TTL of 10 minutes.
func NewTemporaryClientFromPKI(ca *x509.PEMEncodedCertificateAndKey, endpoint *url.URL) (client *Client, err error) {
	opts := []x509.Option{
		x509.CommonName(constants.KubernetesAdminCertCommonName),
		x509.Organization(constants.KubernetesAdminCertOrganization),
		x509.NotBefore(time.Now().Add(-time.Minute)), // allow for a minute for the time to be not in sync across nodes
		x509.NotAfter(time.Now().Add(10 * time.Minute)),
	}

	k8sCA, err := x509.NewCertificateAuthorityFromCertificateAndKey(ca)
	if err != nil {
		return nil, fmt.Errorf("failed decoding Kubernetes CA: %w", err)
	}

	keyPair, err := x509.NewKeyPair(k8sCA, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed generating temporary cert: %w", err)
	}

	h, err := NewClientFromPKI(ca.Crt, keyPair.CrtPEM, keyPair.KeyPEM, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return h, nil
}

// Close all connections.
func (h *Client) Close() error {
	h.dialer.CloseAll()

	return nil
}

// MasterIPs returns a list of control plane endpoints (IP addresses).
func (h *Client) MasterIPs(ctx context.Context) (addrs []string, err error) {
	endpoints, err := h.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	addrs = []string{}

	for _, endpoint := range endpoints.Subsets {
		for _, addr := range endpoint.Addresses {
			addrs = append(addrs, addr.IP)
		}
	}

	return addrs, nil
}

// NodeIPs returns list of node IP addresses by machine type.
//
//nolint:gocyclo
func (h *Client) NodeIPs(ctx context.Context, machineType machine.Type) (addrs []string, err error) {
	resp, err := h.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	addrs = []string{}

	for _, node := range resp.Items {
		_, labelMaster := node.Labels[constants.LabelNodeRoleMaster]
		_, labelControlPlane := node.Labels[constants.LabelNodeRoleControlPlane]

		var skip bool

		switch machineType {
		case machine.TypeInit, machine.TypeControlPlane:
			skip = !(labelMaster || labelControlPlane)
		case machine.TypeWorker:
			skip = labelMaster || labelControlPlane
		case machine.TypeUnknown:
			fallthrough
		default:
			panic(fmt.Sprintf("unexpected machine type %v", machineType))
		}

		if skip {
			continue
		}

		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == corev1.NodeInternalIP {
				addrs = append(addrs, nodeAddress.Address)

				break
			}
		}
	}

	return addrs, nil
}

// LabelNodeAsMaster labels a node with the required master label and NoSchedule taint.
//
//nolint:gocyclo
func (h *Client) LabelNodeAsMaster(ctx context.Context, name string, taintNoSchedule bool) (err error) {
	n, err := h.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// The node may appear to have no labels at first, so we check for the
	// existence of a well known label to ensure the patch will be successful.
	if _, found := n.ObjectMeta.Labels[corev1.LabelHostname]; !found {
		return errors.New("could not find hostname label")
	}

	oldData, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal unmodified node %q into JSON: %w", n.Name, err)
	}

	n.Labels[constants.LabelNodeRoleMaster] = ""
	n.Labels[constants.LabelNodeRoleControlPlane] = ""

	taintIndex := -1

	// TODO: with K8s 1.21, add new taint LabelNodeRoleControlPlane

	for i, taint := range n.Spec.Taints {
		if taint.Key == constants.LabelNodeRoleMaster {
			taintIndex = i

			break
		}
	}

	if taintIndex == -1 && taintNoSchedule {
		n.Spec.Taints = append(n.Spec.Taints, corev1.Taint{
			Key:    constants.LabelNodeRoleMaster,
			Effect: corev1.TaintEffectNoSchedule,
		})
	} else if taintIndex != -1 && !taintNoSchedule {
		n.Spec.Taints = append(n.Spec.Taints[:taintIndex], n.Spec.Taints[taintIndex+1:]...)
	}

	newData, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal modified node %q into JSON: %w", n.Name, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Node{})
	if err != nil {
		return fmt.Errorf("failed to create two way merge patch: %w", err)
	}

	if _, err := h.CoreV1().Nodes().Patch(ctx, n.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			return fmt.Errorf("unable to update node metadata due to conflict: %w", err)
		}

		return fmt.Errorf("error patching node %q: %w", n.Name, err)
	}

	return nil
}

// WaitUntilReady waits for a node to be ready.
func (h *Client) WaitUntilReady(ctx context.Context, name string) error {
	return retry.Exponential(10*time.Minute, retry.WithUnits(250*time.Millisecond), retry.WithJitter(50*time.Millisecond), retry.WithErrorLogging(true)).RetryWithContext(ctx,
		func(ctx context.Context) error {
			attemptCtx, attemptCtxCancel := context.WithTimeout(ctx, 30*time.Second)
			defer attemptCtxCancel()

			node, err := h.CoreV1().Nodes().Get(attemptCtx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return retry.ExpectedError(err)
				}

				if apierrors.ReasonForError(err) == metav1.StatusReasonUnknown || IsRetryableError(err) {
					// non-API error, e.g. networking error
					return retry.ExpectedError(err)
				}

				return err
			}

			for _, cond := range node.Status.Conditions {
				if cond.Type == corev1.NodeReady {
					if cond.Status != corev1.ConditionTrue {
						return retry.ExpectedError(fmt.Errorf("node not ready"))
					}
				}
			}

			return nil
		})
}

// CordonAndDrain cordons and drains a node in one call.
func (h *Client) CordonAndDrain(ctx context.Context, node string) (err error) {
	if err = h.Cordon(ctx, node); err != nil {
		return err
	}

	return h.Drain(ctx, node)
}

// Cordon marks a node as unschedulable.
func (h *Client) Cordon(ctx context.Context, name string) error {
	err := retry.Exponential(30*time.Second, retry.WithUnits(250*time.Millisecond), retry.WithJitter(50*time.Millisecond)).RetryWithContext(ctx, func(ctx context.Context) error {
		node, err := h.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if IsRetryableError(err) {
				return retry.ExpectedError(err)
			}

			return err
		}

		if node.Spec.Unschedulable {
			return nil
		}

		node.Annotations[constants.AnnotationCordonedKey] = constants.AnnotationCordonedValue
		node.Spec.Unschedulable = true

		if _, err := h.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
			return retry.ExpectedError(err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to cordon node %s: %w", name, err)
	}

	return nil
}

// Uncordon marks a node as schedulable.
//
// If force is set, node will be uncordoned even if cordoned not by Talos.
func (h *Client) Uncordon(ctx context.Context, name string, force bool) error {
	err := retry.Exponential(30*time.Second, retry.WithUnits(250*time.Millisecond), retry.WithJitter(50*time.Millisecond)).RetryWithContext(ctx, func(ctx context.Context) error {
		attemptCtx, attemptCtxCancel := context.WithTimeout(ctx, 10*time.Second)
		defer attemptCtxCancel()

		node, err := h.CoreV1().Nodes().Get(attemptCtx, name, metav1.GetOptions{})
		if err != nil {
			if IsRetryableError(err) {
				return retry.ExpectedError(err)
			}

			return err
		}

		if !force && node.Annotations[constants.AnnotationCordonedKey] != constants.AnnotationCordonedValue {
			// not cordoned by Talos, skip it
			return nil
		}

		if node.Spec.Unschedulable {
			node.Spec.Unschedulable = false
			delete(node.Annotations, constants.AnnotationCordonedKey)

			if _, err := h.CoreV1().Nodes().Update(attemptCtx, node, metav1.UpdateOptions{}); err != nil {
				return retry.ExpectedError(err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to uncordon node %s: %w", name, err)
	}

	return nil
}

// Drain evicts all pods on a given node.
func (h *Client) Drain(ctx context.Context, node string) error {
	ctx, cancel := context.WithTimeout(ctx, DrainTimeout)
	defer cancel()

	opts := metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	}

	pods, err := h.CoreV1().Pods(metav1.NamespaceAll).List(ctx, opts)
	if err != nil {
		return fmt.Errorf("cannot get pods for node %s: %w", node, err)
	}

	var eg errgroup.Group

	// Evict each pod.

	for _, pod := range pods.Items {
		p := pod

		eg.Go(func() error {
			if _, ok := p.ObjectMeta.Annotations[corev1.MirrorPodAnnotationKey]; ok {
				log.Printf("skipping mirror pod %s/%s\n", p.GetNamespace(), p.GetName())

				return nil
			}

			controllerRef := metav1.GetControllerOf(&p)

			if controllerRef == nil {
				log.Printf("skipping unmanaged pod %s/%s\n", p.GetNamespace(), p.GetName())

				return nil
			}

			if controllerRef.Kind == appsv1.SchemeGroupVersion.WithKind("DaemonSet").Kind {
				log.Printf("skipping DaemonSet pod %s/%s\n", p.GetNamespace(), p.GetName())

				return nil
			}

			if !p.DeletionTimestamp.IsZero() {
				log.Printf("skipping deleted pod %s/%s\n", p.GetNamespace(), p.GetName())
			}

			if err := h.evict(ctx, p, int64(60)); err != nil {
				log.Printf("WARNING: failed to evict pod: %v", err)
			}

			return nil
		})
	}

	return eg.Wait()
}

func (h *Client) evict(ctx context.Context, p corev1.Pod, gracePeriod int64) error {
	for {
		pol := &policy.Eviction{
			ObjectMeta:    metav1.ObjectMeta{Namespace: p.GetNamespace(), Name: p.GetName()},
			DeleteOptions: &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod},
		}
		err := h.CoreV1().Pods(p.GetNamespace()).Evict(ctx, pol)

		switch {
		case apierrors.IsTooManyRequests(err):
			time.Sleep(5 * time.Second)
		case apierrors.IsNotFound(err):
			return nil
		case err != nil:
			return fmt.Errorf("failed to evict pod %s/%s: %w", p.GetNamespace(), p.GetName(), err)
		default:
			if err = h.waitForPodDeleted(ctx, &p); err != nil {
				return fmt.Errorf("failed waiting on pod %s/%s to be deleted: %w", p.GetNamespace(), p.GetName(), err)
			}

			return nil
		}
	}
}

func (h *Client) waitForPodDeleted(ctx context.Context, p *corev1.Pod) error {
	return retry.Constant(time.Minute, retry.WithUnits(3*time.Second)).RetryWithContext(ctx, func(ctx context.Context) error {
		pod, err := h.CoreV1().Pods(p.GetNamespace()).Get(ctx, p.GetName(), metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return nil
		case err != nil:
			if IsRetryableError(err) {
				return retry.ExpectedError(err)
			}

			return fmt.Errorf("failed to get pod %s/%s: %w", p.GetNamespace(), p.GetName(), err)
		}

		if pod.GetUID() != p.GetUID() {
			return nil
		}

		return retry.ExpectedError(errors.New("pod is still running on the node"))
	})
}
