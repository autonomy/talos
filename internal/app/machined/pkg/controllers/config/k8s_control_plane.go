// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	talosnet "github.com/talos-systems/net"
	"go.uber.org/zap"

	"github.com/talos-systems/talos/pkg/images"
	talosconfig "github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/talos-systems/talos/pkg/resources/config"
)

// K8sControlPlaneController manages config.K8sControlPlane based on configuration.
type K8sControlPlaneController struct{}

// Name implements controller.Controller interface.
func (ctrl *K8sControlPlaneController) Name() string {
	return "config.K8sControlPlaneController"
}

// Inputs implements controller.Controller interface.
func (ctrl *K8sControlPlaneController) Inputs() []controller.Input {
	return []controller.Input{
		{
			Namespace: config.NamespaceName,
			Type:      config.MachineConfigType,
			ID:        pointer.ToString(config.V1Alpha1ID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: config.NamespaceName,
			Type:      config.MachineTypeType,
			ID:        pointer.ToString(config.MachineTypeID),
			Kind:      controller.InputWeak,
		},
	}
}

// Outputs implements controller.Controller interface.
func (ctrl *K8sControlPlaneController) Outputs() []controller.Output {
	return []controller.Output{
		{
			Type: config.K8sControlPlaneType,
			Kind: controller.OutputExclusive,
		},
	}
}

// Run implements controller.Controller interface.
//
//nolint:gocyclo
func (ctrl *K8sControlPlaneController) Run(ctx context.Context, r controller.Runtime, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		}

		cfg, err := r.Get(ctx, resource.NewMetadata(config.NamespaceName, config.MachineConfigType, config.V1Alpha1ID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				if err = ctrl.teardownAll(ctx, r, logger); err != nil {
					return fmt.Errorf("error destroying resources: %w", err)
				}

				continue
			}

			return fmt.Errorf("error getting config: %w", err)
		}

		cfgProvider := cfg.(*config.MachineConfig).Config()

		machineTypeRes, err := r.Get(ctx, resource.NewMetadata(config.NamespaceName, config.MachineTypeType, config.MachineTypeID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return fmt.Errorf("error getting machine type: %w", err)
		}

		machineType := machineTypeRes.(*config.MachineType).MachineType()

		if machineType == machine.TypeWorker {
			if err = ctrl.teardownAll(ctx, r, logger); err != nil {
				return fmt.Errorf("error destroying resources: %w", err)
			}

			continue
		}

		for _, f := range []func(context.Context, controller.Runtime, *zap.Logger, talosconfig.Provider) error{
			ctrl.manageAPIServerConfig,
			ctrl.manageControllerManagerConfig,
			ctrl.manageSchedulerConfig,
			ctrl.manageManifestsConfig,
			ctrl.manageExtraManifestsConfig,
		} {
			if err = f(ctx, r, logger, cfgProvider); err != nil {
				return fmt.Errorf("error updating objects: %w", err)
			}
		}
	}
}

func convertVolumes(volumes []talosconfig.VolumeMount) []config.K8sExtraVolume {
	result := make([]config.K8sExtraVolume, 0, len(volumes))

	for _, volume := range volumes {
		result = append(result, config.K8sExtraVolume{
			Name:      volume.Name(),
			HostPath:  volume.HostPath(),
			MountPath: volume.MountPath(),
			ReadOnly:  volume.ReadOnly(),
		})
	}

	return result
}

func (ctrl *K8sControlPlaneController) manageAPIServerConfig(ctx context.Context, r controller.Runtime, logger *zap.Logger, cfgProvider talosconfig.Provider) error {
	var cloudProvider string
	if cfgProvider.Cluster().ExternalCloudProvider().Enabled() {
		cloudProvider = "external"
	}

	return r.Modify(ctx, config.NewK8sControlPlaneAPIServer(), func(r resource.Resource) error {
		r.(*config.K8sControlPlane).SetAPIServer(config.K8sControlPlaneAPIServerSpec{
			Image:                cfgProvider.Cluster().APIServer().Image(),
			CloudProvider:        cloudProvider,
			ControlPlaneEndpoint: cfgProvider.Cluster().Endpoint().String(),
			EtcdServers:          []string{"https://127.0.0.1:2379"},
			LocalPort:            cfgProvider.Cluster().LocalAPIServerPort(),
			ServiceCIDR:          cfgProvider.Cluster().Network().ServiceCIDR(),
			ExtraArgs:            cfgProvider.Cluster().APIServer().ExtraArgs(),
			ExtraVolumes:         convertVolumes(cfgProvider.Cluster().APIServer().ExtraVolumes()),
		})

		return nil
	})
}

func (ctrl *K8sControlPlaneController) manageControllerManagerConfig(ctx context.Context, r controller.Runtime, logger *zap.Logger, cfgProvider talosconfig.Provider) error {
	var cloudProvider string
	if cfgProvider.Cluster().ExternalCloudProvider().Enabled() {
		cloudProvider = "external"
	}

	return r.Modify(ctx, config.NewK8sControlPlaneControllerManager(), func(r resource.Resource) error {
		r.(*config.K8sControlPlane).SetControllerManager(config.K8sControlPlaneControllerManagerSpec{
			Image:         cfgProvider.Cluster().ControllerManager().Image(),
			CloudProvider: cloudProvider,
			PodCIDR:       cfgProvider.Cluster().Network().PodCIDR(),
			ServiceCIDR:   cfgProvider.Cluster().Network().ServiceCIDR(),
			ExtraArgs:     cfgProvider.Cluster().ControllerManager().ExtraArgs(),
			ExtraVolumes:  convertVolumes(cfgProvider.Cluster().ControllerManager().ExtraVolumes()),
		})

		return nil
	})
}

func (ctrl *K8sControlPlaneController) manageSchedulerConfig(ctx context.Context, r controller.Runtime, logger *zap.Logger, cfgProvider talosconfig.Provider) error {
	return r.Modify(ctx, config.NewK8sControlPlaneScheduler(), func(r resource.Resource) error {
		r.(*config.K8sControlPlane).SetScheduler(config.K8sControlPlaneSchedulerSpec{
			Image:        cfgProvider.Cluster().Scheduler().Image(),
			ExtraArgs:    cfgProvider.Cluster().Scheduler().ExtraArgs(),
			ExtraVolumes: convertVolumes(cfgProvider.Cluster().Scheduler().ExtraVolumes()),
		})

		return nil
	})
}

func (ctrl *K8sControlPlaneController) manageManifestsConfig(ctx context.Context, r controller.Runtime, logger *zap.Logger, cfgProvider talosconfig.Provider) error {
	dnsServiceIPs, err := cfgProvider.Cluster().Network().DNSServiceIPs()
	if err != nil {
		return fmt.Errorf("error calculating DNS service IPs: %w", err)
	}

	dnsServiceIP := ""
	dnsServiceIPv6 := ""

	if len(dnsServiceIPs) == 1 {
		dnsServiceIP = dnsServiceIPs[0].String()
	} else {
		for _, ip := range dnsServiceIPs {
			if dnsServiceIP == "" && ip.To4().Equal(ip) {
				dnsServiceIP = ip.String()
			}

			if dnsServiceIPv6 == "" && talosnet.IsNonLocalIPv6(ip) {
				dnsServiceIPv6 = ip.String()
			}
		}
	}

	return r.Modify(ctx, config.NewK8sManifests(), func(r resource.Resource) error {
		images := images.List(cfgProvider)

		r.(*config.K8sControlPlane).SetManifests(config.K8sManifestsSpec{
			Server:        cfgProvider.Cluster().Endpoint().String(),
			ClusterDomain: cfgProvider.Cluster().Network().DNSDomain(),

			PodCIDRs:     cfgProvider.Cluster().Network().PodCIDR(),
			FirstPodCIDR: strings.Split(cfgProvider.Cluster().Network().PodCIDR(), ",")[0],

			ProxyEnabled:   cfgProvider.Cluster().Proxy().Enabled(),
			ProxyImage:     cfgProvider.Cluster().Proxy().Image(),
			ProxyMode:      cfgProvider.Cluster().Proxy().Mode(),
			ProxyExtraArgs: cfgProvider.Cluster().Proxy().ExtraArgs(),

			CoreDNSEnabled:  cfgProvider.Cluster().CoreDNS().Enabled(),
			CoreDNSImage:    cfgProvider.Cluster().CoreDNS().Image(),
			CoreDNSReplicas: cfgProvider.Cluster().CoreDNS().Replicas(),

			DNSServiceIP:   dnsServiceIP,
			DNSServiceIPv6: dnsServiceIPv6,

			FlannelEnabled:  cfgProvider.Cluster().Network().CNI().Name() == constants.FlannelCNI,
			FlannelImage:    images.Flannel,
			FlannelCNIImage: images.FlannelCNI,
		})

		return nil
	})
}

func (ctrl *K8sControlPlaneController) manageExtraManifestsConfig(ctx context.Context, r controller.Runtime, logger *zap.Logger, cfgProvider talosconfig.Provider) error {
	return r.Modify(ctx, config.NewK8sExtraManifests(), func(r resource.Resource) error {
		spec := config.K8sExtraManifestsSpec{}

		for _, url := range cfgProvider.Cluster().Network().CNI().URLs() {
			spec.ExtraManifests = append(spec.ExtraManifests, config.ExtraManifest{
				Name:     url,
				URL:      url,
				Priority: "05", // push CNI to the top
			})
		}

		for _, url := range cfgProvider.Cluster().ExternalCloudProvider().ManifestURLs() {
			spec.ExtraManifests = append(spec.ExtraManifests, config.ExtraManifest{
				Name:     url,
				URL:      url,
				Priority: "30", // after default manifests
			})
		}

		for _, url := range cfgProvider.Cluster().ExtraManifestURLs() {
			spec.ExtraManifests = append(spec.ExtraManifests, config.ExtraManifest{
				Name:         url,
				URL:          url,
				Priority:     "99", // make sure extra manifests come last, when PSP is already created
				ExtraHeaders: cfgProvider.Cluster().ExtraManifestHeaderMap(),
			})
		}

		for _, manifest := range cfgProvider.Cluster().InlineManifests() {
			spec.ExtraManifests = append(spec.ExtraManifests, config.ExtraManifest{
				Name:           manifest.Name(),
				Priority:       "99", // make sure extra manifests come last, when PSP is already created
				InlineManifest: manifest.Contents(),
			})
		}

		r.(*config.K8sControlPlane).SetExtraManifests(spec)

		return nil
	})
}

func (ctrl *K8sControlPlaneController) teardownAll(ctx context.Context, r controller.Runtime, logger *zap.Logger) error {
	return nil
}
