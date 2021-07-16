// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/AlekSi/pointer"
	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"go.uber.org/zap"

	"github.com/talos-systems/talos/pkg/resources/config"
	"github.com/talos-systems/talos/pkg/resources/k8s"
	"github.com/talos-systems/talos/pkg/resources/secrets"
)

// ManifestController renders manifests based on templates and config/secrets.
type ManifestController struct{}

// Name implements controller.Controller interface.
func (ctrl *ManifestController) Name() string {
	return "k8s.ManifestController"
}

// Inputs implements controller.Controller interface.
func (ctrl *ManifestController) Inputs() []controller.Input {
	return []controller.Input{
		{
			Namespace: config.NamespaceName,
			Type:      config.K8sControlPlaneType,
			ID:        pointer.ToString(config.K8sManifestsID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: secrets.NamespaceName,
			Type:      secrets.RootType,
			ID:        pointer.ToString(secrets.RootKubernetesID),
			Kind:      controller.InputWeak,
		},
	}
}

// Outputs implements controller.Controller interface.
func (ctrl *ManifestController) Outputs() []controller.Output {
	return []controller.Output{
		{
			Type: k8s.ManifestType,
			Kind: controller.OutputShared,
		},
	}
}

// Run implements controller.Controller interface.
//
//nolint:gocyclo
func (ctrl *ManifestController) Run(ctx context.Context, r controller.Runtime, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		}

		configResource, err := r.Get(ctx, resource.NewMetadata(config.NamespaceName, config.K8sControlPlaneType, config.K8sManifestsID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				if err = ctrl.teardownAll(ctx, r); err != nil {
					return fmt.Errorf("error tearing down: %w", err)
				}

				continue
			}

			return err
		}

		configVersion := configResource.(*config.K8sControlPlane).Metadata().Version().String()
		config := configResource.(*config.K8sControlPlane).Manifests()

		secretsResources, err := r.Get(ctx, resource.NewMetadata(secrets.NamespaceName, secrets.RootType, secrets.RootKubernetesID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				if err = ctrl.teardownAll(ctx, r); err != nil {
					return fmt.Errorf("error tearing down: %w", err)
				}

				continue
			}

			return err
		}

		secrets := secretsResources.(*secrets.Root).KubernetesSpec()

		renderedManifests, err := ctrl.render(configVersion, config, secrets)
		if err != nil {
			return err
		}

		for _, renderedManifest := range renderedManifests {
			renderedManifest := renderedManifest

			if err = r.Modify(ctx, k8s.NewManifest(k8s.ControlPlaneNamespaceName, renderedManifest.name),
				func(r resource.Resource) error {
					return r.(*k8s.Manifest).SetYAML(renderedManifest.data)
				}); err != nil {
				return fmt.Errorf("error updating manifests: %w", err)
			}
		}

		// remove any manifests which weren't rendered
		manifests, err := r.List(ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.ManifestType, "", resource.VersionUndefined))
		if err != nil {
			return fmt.Errorf("error listing manifests: %w", err)
		}

		manifestsToDelete := map[string]struct{}{}

		for _, manifest := range manifests.Items {
			if manifest.Metadata().Owner() != ctrl.Name() {
				continue
			}

			manifestsToDelete[manifest.Metadata().ID()] = struct{}{}
		}

		for _, renderedManifest := range renderedManifests {
			delete(manifestsToDelete, renderedManifest.name)
		}

		for id := range manifestsToDelete {
			if err = r.Destroy(ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.ManifestType, id, resource.VersionUndefined)); err != nil {
				return fmt.Errorf("error cleaning up manifests: %w", err)
			}
		}
	}
}

type renderedManifest struct {
	name string
	data []byte
}

func jsonify(input string) (string, error) {
	out, err := json.Marshal(input)

	return string(out), err
}

func (ctrl *ManifestController) render(version string, cfg config.K8sManifestsSpec, scrt *secrets.RootKubernetesSpec) ([]renderedManifest, error) {
	templateConfig := struct {
		ConfigVersion string
		config.K8sManifestsSpec

		Secrets *secrets.RootKubernetesSpec
	}{
		ConfigVersion:    version,
		K8sManifestsSpec: cfg,
		Secrets:          scrt,
	}

	type manifestDesc struct {
		name     string
		template []byte
	}

	defaultManifests := []manifestDesc{
		{"00-kubelet-bootstrapping-token", kubeletBootstrappingToken},
		{"01-csr-node-bootstrap", csrNodeBootstrapTemplate},
		{"01-csr-approver-role-binding", csrApproverRoleBindingTemplate},
		{"01-csr-renewal-role-binding", csrRenewalRoleBindingTemplate},
		{"02-kube-system-sa-role-binding", kubeSystemSARoleBindingTemplate},
		{"03-default-pod-security-policy", podSecurityPolicy},
		{"11-kube-config-in-cluster", kubeConfigInClusterTemplate},
	}

	if cfg.CoreDNSEnabled {
		defaultManifests = append(defaultManifests,
			[]manifestDesc{
				{"11-core-dns", coreDNSTemplate},
				{"11-core-dns-svc", coreDNSSvcTemplate},
			}...,
		)
	}

	if cfg.FlannelEnabled {
		defaultManifests = append(defaultManifests,
			[]manifestDesc{
				{"05-flannel", flannelTemplate},
			}...,
		)
	}

	if cfg.ProxyEnabled {
		defaultManifests = append(defaultManifests,
			[]manifestDesc{
				{"10-kube-proxy", kubeProxyTemplate},
			}...,
		)
	}

	manifests := make([]renderedManifest, len(defaultManifests))

	for i := range defaultManifests {
		tmpl, err := template.New(defaultManifests[i].name).
			Funcs(template.FuncMap{
				"json": jsonify,
			}).
			Parse(string(defaultManifests[i].template))
		if err != nil {
			return nil, fmt.Errorf("error parsing manifest template %q: %w", defaultManifests[i].name, err)
		}

		var buf bytes.Buffer

		if err = tmpl.Execute(&buf, &templateConfig); err != nil {
			return nil, fmt.Errorf("error executing template %q: %w", defaultManifests[i].name, err)
		}

		manifests[i].name = defaultManifests[i].name
		manifests[i].data = buf.Bytes()
	}

	return manifests, nil
}

//nolint:dupl
func (ctrl *ManifestController) teardownAll(ctx context.Context, r controller.Runtime) error {
	manifests, err := r.List(ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.ManifestType, "", resource.VersionUndefined))
	if err != nil {
		return fmt.Errorf("error listing manifests: %w", err)
	}

	for _, manifest := range manifests.Items {
		if manifest.Metadata().Owner() != ctrl.Name() {
			continue
		}

		if err = r.Destroy(ctx, manifest.Metadata()); err != nil {
			return fmt.Errorf("error destroying manifest: %w", err)
		}
	}

	return nil
}
