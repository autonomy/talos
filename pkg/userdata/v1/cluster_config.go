/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package v1

import "github.com/talos-systems/talos/pkg/userdata/token"

// ClusterConfig reperesents the cluster-wide config values
type ClusterConfig struct {
	ControlPlane      *ControlPlaneConfig      `yaml:"controlPlane"`
	ClusterName       string                   `yaml:"clusterName,omitempty"`
	Network           *ClusterNetworkConfig    `yaml:"network,omitempty"`
	Token             string                   `yaml:"token,omitempty"`
	InitToken         *token.Token             `yaml:"initToken,omitempty"`
	CA                *ClusterCAConfig         `yaml:"ca,omitempty"`
	APIServer         *APIServerConfig         `yaml:"apiServer,omitempty"`
	ControllerManager *ControllerManagerConfig `yaml:"controllerManager,omitempty"`
	Scheduler         *SchedulerConfig         `yaml:"scheduler,omitempty"`
	Etcd              *EtcdConfig              `yaml:"etcd,omitempty"`
}

// ControlPlaneConfig represents control plane config vals
type ControlPlaneConfig struct {
	IPs   []string `yaml:"ips"`
	Index int      `yaml:"index,omitempty"`
}

// APIServerConfig represents kube apiserver config vals
type APIServerConfig struct {
	Image     string            `yaml:"image,omitempty"`
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
	CertSANs  []string          `yaml:"certSANs,omitempty"`
}

// ControllerManagerConfig represents kube controller manager config vals
type ControllerManagerConfig struct {
	Image     string            `yaml:"image,omitempty"`
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}

// SchedulerConfig represents kube scheduler config vals
type SchedulerConfig struct {
	Image     string            `yaml:"image,omitempty"`
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}

// EtcdConfig represents etcd config vals
type EtcdConfig struct {
	Image string `yaml:"image,omitempty"`
}

// ClusterNetworkConfig represents kube networking config vals
type ClusterNetworkConfig struct {
	DNSDomain     string   `yaml:"dnsDomain"`
	PodSubnet     []string `yaml:"podSubnets"`
	ServiceSubnet []string `yaml:"serviceSubnets"`
}

// ClusterCAConfig represents kube cert config vals
type ClusterCAConfig struct {
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}
