// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/meta"
)

// K8sControlPlaneType is type of K8sControlPlane resource.
const K8sControlPlaneType = resource.Type("KubernetesControlPlaneConfigs.config.talos.dev")

// K8sControlPlaneAPIServerID is an ID of kube-apiserver config.
const K8sControlPlaneAPIServerID = resource.ID("kube-apiserver")

// K8sControlPlaneControllerManagerID is an ID of kube-controller-manager config.
const K8sControlPlaneControllerManagerID = resource.ID("kube-controller-manager")

// K8sControlPlaneSchedulerID is an ID of kube-scheduler config.
const K8sControlPlaneSchedulerID = resource.ID("kube-scheduler")

// K8sManifestsID is an ID of manifests config.
const K8sManifestsID = resource.ID("system-manifests")

// K8sExtraManifestsID is an ID of extra manifests config.
const K8sExtraManifestsID = resource.ID("extra-manifests")

// K8sControlPlane describes machine type.
type K8sControlPlane struct {
	md resource.Metadata
	// spec stores values of different types depending on ID
	spec interface{}
}

// K8sExtraVolume is a configuration of extra volume.
type K8sExtraVolume struct {
	Name      string `yaml:"name"`
	HostPath  string `yaml:"hostPath"`
	MountPath string `yaml:"mountPath"`
	ReadOnly  bool   `yaml:"readonly"`
}

// K8sControlPlaneAPIServerSpec is configuration for kube-apiserver.
type K8sControlPlaneAPIServerSpec struct {
	Image                string            `yaml:"image"`
	CloudProvider        string            `yaml:"cloudProvider"`
	ControlPlaneEndpoint string            `yaml:"controlPlaneEndpoint"`
	EtcdServers          []string          `yaml:"etcdServers"`
	LocalPort            int               `yaml:"localPort"`
	ServiceCIDR          string            `yaml:"serviceCIDR"`
	ExtraArgs            map[string]string `yaml:"extraArgs"`
	ExtraVolumes         []K8sExtraVolume  `yaml:"extraVolumes"`
}

// K8sControlPlaneControllerManagerSpec is configuration for kube-controller-manager.
type K8sControlPlaneControllerManagerSpec struct {
	Image         string            `yaml:"image"`
	CloudProvider string            `yaml:"cloudProvider"`
	PodCIDR       string            `yaml:"podCIDR"`
	ServiceCIDR   string            `yaml:"serviceCIDR"`
	ExtraArgs     map[string]string `yaml:"extraArgs"`
	ExtraVolumes  []K8sExtraVolume  `yaml:"extraVolumes"`
}

// K8sControlPlaneSchedulerSpec is configuration for kube-scheduler.
type K8sControlPlaneSchedulerSpec struct {
	Image        string            `yaml:"image"`
	ExtraArgs    map[string]string `yaml:"extraArgs"`
	ExtraVolumes []K8sExtraVolume  `yaml:"extraVolumes"`
}

// K8sManifestsSpec is configuration for manifests.
//
//nolint:maligned
type K8sManifestsSpec struct {
	Server        string `yaml:"string"`
	ClusterDomain string `yaml:"clusterDomain"`

	PodCIDRs     string `yaml:"podCIDRs"`
	FirstPodCIDR string `yaml:"firstPodCIDR"`

	ProxyEnabled   bool              `yaml:"proxyEnabled"`
	ProxyImage     string            `yaml:"proxyImage"`
	ProxyMode      string            `yaml:"proxyMode"`
	ProxyExtraArgs map[string]string `yaml:"proxyExtraArgs"`

	CoreDNSEnabled  bool   `yaml:"coreDNSEnabled"`
	CoreDNSImage    string `yaml:"coreDNSImage"`
	CoreDNSReplicas int    `yaml:"coreDNSReplicas"`

	DNSServiceIP   string `yaml:"dnsServiceIP"`
	DNSServiceIPv6 string `yaml:"dnsServiceIPv6"`

	FlannelEnabled  bool   `yaml:"flannelEnabled"`
	FlannelImage    string `yaml:"flannelImage"`
	FlannelCNIImage string `yaml:"flannelCNIImage"`
}

// ExtraManifest defines a single extra manifest to download.
type ExtraManifest struct {
	Name           string            `yaml:"name"`
	URL            string            `yaml:"url"`
	Priority       string            `yaml:"priority"`
	ExtraHeaders   map[string]string `yaml:"extraHeaders"`
	InlineManifest string            `yaml:"inlineManifest"`
}

// K8sExtraManifestsSpec is a configuration for extra manifests.
type K8sExtraManifestsSpec struct {
	ExtraManifests []ExtraManifest `yaml:"extraManifests"`
}

// NewK8sControlPlaneAPIServer initializes a K8sControlPlane resource.
func NewK8sControlPlaneAPIServer() *K8sControlPlane {
	r := &K8sControlPlane{
		md:   resource.NewMetadata(NamespaceName, K8sControlPlaneType, K8sControlPlaneAPIServerID, resource.VersionUndefined),
		spec: K8sControlPlaneAPIServerSpec{},
	}

	r.md.BumpVersion()

	return r
}

// NewK8sControlPlaneControllerManager initializes a K8sControlPlane resource.
func NewK8sControlPlaneControllerManager() *K8sControlPlane {
	r := &K8sControlPlane{
		md:   resource.NewMetadata(NamespaceName, K8sControlPlaneType, K8sControlPlaneControllerManagerID, resource.VersionUndefined),
		spec: K8sControlPlaneControllerManagerSpec{},
	}

	r.md.BumpVersion()

	return r
}

// NewK8sControlPlaneScheduler initializes a K8sControlPlane resource.
func NewK8sControlPlaneScheduler() *K8sControlPlane {
	r := &K8sControlPlane{
		md:   resource.NewMetadata(NamespaceName, K8sControlPlaneType, K8sControlPlaneSchedulerID, resource.VersionUndefined),
		spec: K8sControlPlaneSchedulerSpec{},
	}

	r.md.BumpVersion()

	return r
}

// NewK8sManifests initializes a K8sControlPlane resource.
func NewK8sManifests() *K8sControlPlane {
	r := &K8sControlPlane{
		md:   resource.NewMetadata(NamespaceName, K8sControlPlaneType, K8sManifestsID, resource.VersionUndefined),
		spec: K8sManifestsSpec{},
	}

	r.md.BumpVersion()

	return r
}

// NewK8sExtraManifests initializes a K8sControlPlane resource.
func NewK8sExtraManifests() *K8sControlPlane {
	r := &K8sControlPlane{
		md:   resource.NewMetadata(NamespaceName, K8sControlPlaneType, K8sExtraManifestsID, resource.VersionUndefined),
		spec: K8sExtraManifestsSpec{},
	}

	r.md.BumpVersion()

	return r
}

// Metadata implements resource.Resource.
func (r *K8sControlPlane) Metadata() *resource.Metadata {
	return &r.md
}

// Spec implements resource.Resource.
func (r *K8sControlPlane) Spec() interface{} {
	return r.spec
}

func (r *K8sControlPlane) String() string {
	return fmt.Sprintf("config.KubernetesControlPlaneConfig(%q)", r.md.ID())
}

// DeepCopy implements resource.Resource.
func (r *K8sControlPlane) DeepCopy() resource.Resource {
	return &K8sControlPlane{
		md:   r.md,
		spec: r.spec,
	}
}

// ResourceDefinition implements meta.ResourceDefinitionProvider interface.
func (r *K8sControlPlane) ResourceDefinition() meta.ResourceDefinitionSpec {
	return meta.ResourceDefinitionSpec{
		Type:             K8sControlPlaneType,
		Aliases:          []resource.Type{},
		DefaultNamespace: NamespaceName,
	}
}

// APIServer returns K8sControlPlaneApiServerSpec.
func (r *K8sControlPlane) APIServer() K8sControlPlaneAPIServerSpec {
	return r.spec.(K8sControlPlaneAPIServerSpec)
}

// SetAPIServer sets K8sControlPlaneApiServerSpec.
func (r *K8sControlPlane) SetAPIServer(spec K8sControlPlaneAPIServerSpec) {
	r.spec = spec
}

// ControllerManager returns K8sControlPlaneControllerManagerSpec.
func (r *K8sControlPlane) ControllerManager() K8sControlPlaneControllerManagerSpec {
	return r.spec.(K8sControlPlaneControllerManagerSpec)
}

// SetControllerManager sets K8sControlPlaneControllerManagerSpec.
func (r *K8sControlPlane) SetControllerManager(spec K8sControlPlaneControllerManagerSpec) {
	r.spec = spec
}

// Scheduler returns K8sControlPlaneSchedulerSpec.
func (r *K8sControlPlane) Scheduler() K8sControlPlaneSchedulerSpec {
	return r.spec.(K8sControlPlaneSchedulerSpec)
}

// SetScheduler sets K8sControlPlaneSchedulerSpec.
func (r *K8sControlPlane) SetScheduler(spec K8sControlPlaneSchedulerSpec) {
	r.spec = spec
}

// Manifests returns K8sManifestsSpec.
func (r *K8sControlPlane) Manifests() K8sManifestsSpec {
	return r.spec.(K8sManifestsSpec)
}

// SetManifests sets K8sManifestsSpec.
func (r *K8sControlPlane) SetManifests(spec K8sManifestsSpec) {
	r.spec = spec
}

// ExtraManifests returns K8sExtraManifestsSpec.
func (r *K8sControlPlane) ExtraManifests() K8sExtraManifestsSpec {
	return r.spec.(K8sExtraManifestsSpec)
}

// SetExtraManifests sets K8sManifestsSpec.
func (r *K8sControlPlane) SetExtraManifests(spec K8sExtraManifestsSpec) {
	r.spec = spec
}
