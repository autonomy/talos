// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package qemu

import (
	"context"

	"github.com/talos-systems/talos/internal/pkg/provision"
	"github.com/talos-systems/talos/internal/pkg/provision/providers/vm"
	"github.com/talos-systems/talos/pkg/config/types/v1alpha1/generate"
)

type provisioner struct {
	vm.Provisioner
}

// NewProvisioner initializes qemu provisioner.
func NewProvisioner(ctx context.Context) (provision.Provisioner, error) {
	p := &provisioner{
		vm.Provisioner{
			Name: "qemu",
		},
	}

	return p, nil
}

// Close and release resources.
func (p *provisioner) Close() error {
	return nil
}

// GenOptions provides a list of additional config generate options.
func (p *provisioner) GenOptions(networkReq provision.NetworkRequest) []generate.GenOption {
	nameservers := make([]string, len(networkReq.Nameservers))
	for i := range nameservers {
		nameservers[i] = networkReq.Nameservers[i].String()
	}

	return []generate.GenOption{
		generate.WithInstallDisk("/dev/vda"),
		generate.WithInstallExtraKernelArgs([]string{
			"console=ttyS0",
			// reboot configuration
			"reboot=k",
			"panic=1",
			// Talos-specific
			"talos.platform=metal",
		}),
	}
}

// GetLoadBalancers returns internal/external loadbalancer endpoints.
func (p *provisioner) GetLoadBalancers(networkReq provision.NetworkRequest) (internalEndpoint, externalEndpoint string) {
	// firecracker runs loadbalancer on the bridge, which is good for both internal & external access
	return networkReq.GatewayAddr.String(), networkReq.GatewayAddr.String()
}
