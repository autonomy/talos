// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package secrets

import (
	"bytes"
	"context"
	stdlibx509 "crypto/x509"
	"fmt"
	"net/url"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/talos-systems/crypto/x509"
	"go.uber.org/zap"

	"github.com/talos-systems/talos/internal/pkg/kubeconfig"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/talos-systems/talos/pkg/resources/network"
	"github.com/talos-systems/talos/pkg/resources/secrets"
	timeresource "github.com/talos-systems/talos/pkg/resources/time"
	"github.com/talos-systems/talos/pkg/resources/v1alpha1"
)

// KubernetesCertificateValidityDuration is the validity duration for the certificates created with this controller.
//
// Controller automatically refreshes certs at 50% of CertificateValidityDuration.
const KubernetesCertificateValidityDuration = constants.KubernetesDefaultCertificateValidityDuration

// KubernetesController manages secrets.Kubernetes based on configuration.
type KubernetesController struct{}

// Name implements controller.Controller interface.
func (ctrl *KubernetesController) Name() string {
	return "secrets.KubernetesController"
}

// Inputs implements controller.Controller interface.
func (ctrl *KubernetesController) Inputs() []controller.Input {
	return []controller.Input{
		{
			Namespace: network.NamespaceName,
			Type:      network.StatusType,
			ID:        pointer.ToString(network.StatusID),
			Kind:      controller.InputWeak,
		},
	}
}

// Outputs implements controller.Controller interface.
func (ctrl *KubernetesController) Outputs() []controller.Output {
	return []controller.Output{
		{
			Type: secrets.KubernetesType,
			Kind: controller.OutputExclusive,
		},
	}
}

// Run implements controller.Controller interface.
//
//nolint:gocyclo,cyclop
func (ctrl *KubernetesController) Run(ctx context.Context, r controller.Runtime, logger *zap.Logger) error {
	// wait for the network to be ready first, then switch to regular inputs
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		}
		// wait for network to be ready as it might change IPs/hostname
		networkResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.StatusType, network.StatusID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		networkStatus := networkResource.(*network.Status).TypedSpec()

		if networkStatus.AddressReady && networkStatus.HostnameReady {
			break
		}
	}

	// switch to regular inputs once the network is ready
	if err := r.UpdateInputs([]controller.Input{
		{
			Namespace: secrets.NamespaceName,
			Type:      secrets.RootType,
			ID:        pointer.ToString(secrets.RootKubernetesID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: v1alpha1.NamespaceName,
			Type:      timeresource.StatusType,
			ID:        pointer.ToString(timeresource.StatusID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: network.NamespaceName,
			Type:      network.HostnameStatusType,
			ID:        pointer.ToString(network.HostnameID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: network.NamespaceName,
			Type:      network.NodeAddressType,
			ID:        pointer.ToString(network.NodeAddressAccumulativeID),
			Kind:      controller.InputWeak,
		},
	}); err != nil {
		return fmt.Errorf("error updating inputs: %w", err)
	}

	r.QueueReconcile()

	rateLimitedEventCh := RateLimitEvents(ctx, r.EventCh(), time.Minute)

	refreshTicker := time.NewTicker(KubernetesCertificateValidityDuration / 2)
	defer refreshTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-rateLimitedEventCh:
		case <-refreshTicker.C:
		}

		k8sRootRes, err := r.Get(ctx, resource.NewMetadata(secrets.NamespaceName, secrets.RootType, secrets.RootKubernetesID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				if err = ctrl.teardownAll(ctx, r); err != nil {
					return fmt.Errorf("error destroying resources: %w", err)
				}

				continue
			}

			return fmt.Errorf("error getting root k8s secrets: %w", err)
		}

		k8sRoot := k8sRootRes.(*secrets.Root).KubernetesSpec()

		// wait for time sync as certs depend on current time
		timeSyncResource, err := r.Get(ctx, resource.NewMetadata(v1alpha1.NamespaceName, timeresource.StatusType, timeresource.StatusID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		if !timeSyncResource.(*timeresource.Status).Status().Synced {
			continue
		}

		hostnameResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.HostnameStatusType, network.HostnameID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		hostnameStatus := hostnameResource.(*network.HostnameStatus).TypedSpec()

		addressesResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.NodeAddressType, network.NodeAddressAccumulativeID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		nodeAddresses := addressesResource.(*network.NodeAddress).TypedSpec()

		if err = r.Modify(ctx, secrets.NewKubernetes(), func(r resource.Resource) error {
			return ctrl.updateSecrets(k8sRoot, r.(*secrets.Kubernetes).Certs(), hostnameStatus, nodeAddresses)
		}); err != nil {
			return err
		}
	}
}

func (ctrl *KubernetesController) updateSecrets(k8sRoot *secrets.RootKubernetesSpec, k8sSecrets *secrets.KubernetesCertsSpec,
	hostnameStatus *network.HostnameStatusSpec, nodeAddresses *network.NodeAddressSpec) error {
	var altNames AltNames

	altNames.Append(k8sRoot.Endpoint.Hostname())
	altNames.Append(k8sRoot.CertSANs...)

	altNames.AppendDNSNames(
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc."+k8sRoot.DNSDomain,
		"localhost",
	)

	altNames.Append(
		hostnameStatus.Hostname,
		hostnameStatus.FQDN(),
	)

	altNames.AppendIPs(k8sRoot.APIServerIPs...)

	for _, addr := range nodeAddresses.Addresses {
		altNames.AppendIPs(addr.IPAddr().IP)
	}

	ca, err := x509.NewCertificateAuthorityFromCertificateAndKey(k8sRoot.CA)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	apiServer, err := x509.NewKeyPair(ca,
		x509.IPAddresses(altNames.IPs),
		x509.DNSNames(altNames.DNSNames),
		x509.CommonName("kube-apiserver"),
		x509.Organization("kube-master"),
		x509.NotAfter(time.Now().Add(KubernetesCertificateValidityDuration)),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageServerAuth,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate api-server cert: %w", err)
	}

	k8sSecrets.APIServer = x509.NewCertificateAndKeyFromKeyPair(apiServer)

	apiServerKubeletClient, err := x509.NewKeyPair(ca,
		x509.CommonName(constants.KubernetesAPIServerKubeletClientCommonName),
		x509.Organization(constants.KubernetesAdminCertOrganization),
		x509.NotAfter(time.Now().Add(KubernetesCertificateValidityDuration)),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageClientAuth,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate api-server cert: %w", err)
	}

	k8sSecrets.APIServerKubeletClient = x509.NewCertificateAndKeyFromKeyPair(apiServerKubeletClient)

	aggregatorCA, err := x509.NewCertificateAuthorityFromCertificateAndKey(k8sRoot.AggregatorCA)
	if err != nil {
		return fmt.Errorf("failed to parse aggregator CA: %w", err)
	}

	frontProxy, err := x509.NewKeyPair(aggregatorCA,
		x509.CommonName("front-proxy-client"),
		x509.NotAfter(time.Now().Add(KubernetesCertificateValidityDuration)),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageClientAuth,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate aggregator cert: %w", err)
	}

	k8sSecrets.FrontProxy = x509.NewCertificateAndKeyFromKeyPair(frontProxy)

	var buf bytes.Buffer

	if err = kubeconfig.Generate(&kubeconfig.GenerateInput{
		ClusterName: k8sRoot.Name,

		CA:                  k8sRoot.CA,
		CertificateLifetime: KubernetesCertificateValidityDuration,

		CommonName:   constants.KubernetesControllerManagerOrganization,
		Organization: constants.KubernetesControllerManagerOrganization,

		Endpoint:    "https://localhost:6443/",
		Username:    constants.KubernetesControllerManagerOrganization,
		ContextName: "default",
	}, &buf); err != nil {
		return fmt.Errorf("failed to generate controller manager kubeconfig: %w", err)
	}

	k8sSecrets.ControllerManagerKubeconfig = buf.String()

	buf.Reset()

	if err = kubeconfig.Generate(&kubeconfig.GenerateInput{
		ClusterName: k8sRoot.Name,

		CA:                  k8sRoot.CA,
		CertificateLifetime: KubernetesCertificateValidityDuration,

		CommonName:   constants.KubernetesSchedulerOrganization,
		Organization: constants.KubernetesSchedulerOrganization,

		Endpoint:    "https://localhost:6443/",
		Username:    constants.KubernetesSchedulerOrganization,
		ContextName: "default",
	}, &buf); err != nil {
		return fmt.Errorf("failed to generate scheduler kubeconfig: %w", err)
	}

	k8sSecrets.SchedulerKubeconfig = buf.String()

	buf.Reset()

	if err = kubeconfig.GenerateAdmin(&generateAdminAdapter{k8sRoot: k8sRoot}, &buf); err != nil {
		return fmt.Errorf("failed to generate admin kubeconfig: %w", err)
	}

	k8sSecrets.AdminKubeconfig = buf.String()

	return nil
}

func (ctrl *KubernetesController) teardownAll(ctx context.Context, r controller.Runtime) error {
	list, err := r.List(ctx, resource.NewMetadata(secrets.NamespaceName, secrets.KubernetesType, "", resource.VersionUndefined))
	if err != nil {
		return err
	}

	// TODO: change this to proper teardown sequence

	for _, res := range list.Items {
		if err = r.Destroy(ctx, res.Metadata()); err != nil {
			return err
		}
	}

	return nil
}

// generateAdminAdapter allows to translate input config into GenerateAdmin input.
type generateAdminAdapter struct {
	k8sRoot *secrets.RootKubernetesSpec
}

func (adapter *generateAdminAdapter) Name() string {
	return adapter.k8sRoot.Name
}

func (adapter *generateAdminAdapter) Endpoint() *url.URL {
	u, _ := url.Parse("https://localhost:6443/") //nolint:errcheck

	return u
}

func (adapter *generateAdminAdapter) CA() *x509.PEMEncodedCertificateAndKey {
	return adapter.k8sRoot.CA
}

func (adapter *generateAdminAdapter) AdminKubeconfig() config.AdminKubeconfig {
	return adapter
}

func (adapter *generateAdminAdapter) CertLifetime() time.Duration {
	// this certificate is not delivered to the user, it's used only internally by control plane components
	return KubernetesCertificateValidityDuration
}
