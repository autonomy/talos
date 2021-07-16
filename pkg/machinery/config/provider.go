// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/talos-systems/crypto/x509"

	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
)

// Provider defines the configuration consumption interface.
type Provider interface {
	Version() string
	Debug() bool
	Persist() bool
	Machine() MachineConfig
	Cluster() ClusterConfig
	// Validate checks configuration and returns warnings and fatal errors (as multierror).
	Validate(RuntimeMode, ...ValidationOption) ([]string, error)
	ApplyDynamicConfig(context.Context, DynamicConfigProvider) error
	String(encoderOptions ...encoder.Option) (string, error)
	Bytes(encoderOptions ...encoder.Option) ([]byte, error)
}

// MachineConfig defines the requirements for a config that pertains to machine
// related options.
type MachineConfig interface {
	Install() Install
	Security() Security
	Network() MachineNetwork
	Disks() []Disk
	Time() Time
	Env() Env
	Files() ([]File, error)
	Type() machine.Type
	Kubelet() Kubelet
	Sysctls() map[string]string
	Registries() Registries
	SystemDiskEncryption() SystemDiskEncryption
	Features() Features
}

// Disk represents the options available for partitioning, formatting, and
// mounting extra disks.
type Disk interface {
	Device() string
	Partitions() []Partition
}

// Partition represents the options for a device partition.
type Partition interface {
	Size() uint64
	MountPoint() string
}

// Env represents a set of environment variables.
type Env = map[string]string

// File represents a file to write to disk.
type File interface {
	Content() string
	Permissions() os.FileMode
	Path() string
	Op() string
}

// Install defines the requirements for a config that pertains to install
// related options.
type Install interface {
	Image() string
	Disk() (string, error)
	ExtraKernelArgs() []string
	Zero() bool
	LegacyBIOSSupport() bool
	WithBootloader() bool
}

// Security defines the requirements for a config that pertains to security
// related options.
type Security interface {
	CA() *x509.PEMEncodedCertificateAndKey
	Token() string
	CertSANs() []string
}

// MachineNetwork defines the requirements for a config that pertains to network
// related options.
type MachineNetwork interface {
	Hostname() string
	Resolvers() []string
	Devices() []Device
	ExtraHosts() []ExtraHost
}

// ExtraHost represents a host entry in /etc/hosts.
type ExtraHost interface {
	IP() string
	Aliases() []string
}

// Device represents a network interface.
type Device interface {
	Interface() string
	CIDR() string
	Routes() []Route
	Bond() Bond
	Vlans() []Vlan
	MTU() int
	DHCP() bool
	Ignore() bool
	Dummy() bool
	DHCPOptions() DHCPOptions
	VIPConfig() VIPConfig
	WireguardConfig() WireguardConfig
}

// DHCPOptions represents a set of DHCP options.
type DHCPOptions interface {
	RouteMetric() uint32
	IPv4() bool
	IPv6() bool
}

// VIPConfig contains settings for the Virtual (shared) IP setup.
type VIPConfig interface {
	IP() string
}

// WireguardConfig contains settings for configuring Wireguard network interface.
type WireguardConfig interface {
	PrivateKey() string
	ListenPort() int
	FirewallMark() int
	Peers() []WireguardPeer
}

// WireguardPeer a WireGuard device peer configuration.
type WireguardPeer interface {
	PublicKey() string
	Endpoint() string
	PersistentKeepaliveInterval() time.Duration
	AllowedIPs() []string
}

// Bond contains the various options for configuring a
// bonded interface.
type Bond interface {
	Interfaces() []string
	ARPIPTarget() []string
	Mode() string
	HashPolicy() string
	LACPRate() string
	ADActorSystem() string
	ARPValidate() string
	ARPAllTargets() string
	Primary() string
	PrimaryReselect() string
	FailOverMac() string
	ADSelect() string
	MIIMon() uint32
	UpDelay() uint32
	DownDelay() uint32
	ARPInterval() uint32
	ResendIGMP() uint32
	MinLinks() uint32
	LPInterval() uint32
	PacketsPerSlave() uint32
	NumPeerNotif() uint8
	TLBDynamicLB() uint8
	AllSlavesActive() uint8
	UseCarrier() bool
	ADActorSysPrio() uint16
	ADUserPortKey() uint16
	PeerNotifyDelay() uint32
}

// Vlan represents vlan settings for a device.
type Vlan interface {
	CIDR() string
	Routes() []Route
	DHCP() bool
	ID() uint16
}

// Route represents a network route.
type Route interface {
	Network() string
	Gateway() string
	Metric() uint32
}

// Time defines the requirements for a config that pertains to time related
// options.
type Time interface {
	Disabled() bool
	Servers() []string
}

// Kubelet defines the requirements for a config that pertains to kubelet
// related options.
type Kubelet interface {
	Image() string
	ClusterDNS() []string
	ExtraArgs() map[string]string
	ExtraMounts() []specs.Mount
	RegisterWithFQDN() bool
}

// Registries defines the configuration for image fetching.
type Registries interface {
	// Mirror config by registry host (first part of image reference).
	Mirrors() map[string]RegistryMirrorConfig
	// Registry config (auth, TLS) by hostname.
	Config() map[string]RegistryConfig
}

// RegistryMirrorConfig represents mirror configuration for a registry.
type RegistryMirrorConfig interface {
	Endpoints() []string
}

// RegistryConfig specifies auth & TLS config per registry.
type RegistryConfig interface {
	TLS() RegistryTLSConfig
	Auth() RegistryAuthConfig
}

// RegistryAuthConfig specifies authentication configuration for a registry.
type RegistryAuthConfig interface {
	Username() string
	Password() string
	Auth() string
	IdentityToken() string
}

// RegistryTLSConfig specifies TLS config for HTTPS registries.
type RegistryTLSConfig interface {
	ClientIdentity() *x509.PEMEncodedCertificateAndKey
	CA() []byte
	InsecureSkipVerify() bool
	GetTLSConfig() (*tls.Config, error)
}

// ClusterConfig defines the requirements for a config that pertains to cluster
// related options.
type ClusterConfig interface {
	Name() string
	APIServer() APIServer
	ControllerManager() ControllerManager
	Proxy() Proxy
	Scheduler() Scheduler
	Endpoint() *url.URL
	Token() Token
	CertSANs() []string
	CA() *x509.PEMEncodedCertificateAndKey
	AggregatorCA() *x509.PEMEncodedCertificateAndKey
	ServiceAccount() *x509.PEMEncodedKey
	AESCBCEncryptionSecret() string
	Config(machine.Type) (string, error)
	Etcd() Etcd
	Network() ClusterNetwork
	LocalAPIServerPort() int
	CoreDNS() CoreDNS
	// ExternalCloudProvider returns external cloud provider settings.
	ExternalCloudProvider() ExternalCloudProvider
	ExtraManifestURLs() []string
	ExtraManifestHeaderMap() map[string]string
	InlineManifests() []InlineManifest
	AdminKubeconfig() AdminKubeconfig
	ScheduleOnMasters() bool
}

// ClusterNetwork defines the requirements for a config that pertains to cluster
// network options.
type ClusterNetwork interface {
	CNI() CNI
	PodCIDR() string
	ServiceCIDR() string
	DNSDomain() string
	// APIServerIPs returns kube-apiserver IPs in the ServiceCIDR.
	APIServerIPs() ([]net.IP, error)
	// DNSServiceIPs returns DNS service IPs in the ServiceCIDR.
	DNSServiceIPs() ([]net.IP, error)
}

// CNI defines the requirements for a config that pertains to Kubernetes
// cni.
type CNI interface {
	Name() string
	URLs() []string
}

// APIServer defines the requirements for a config that pertains to apiserver related
// options.
type APIServer interface {
	Image() string
	ExtraArgs() map[string]string
	ExtraVolumes() []VolumeMount
}

// ControllerManager defines the requirements for a config that pertains to controller manager related
// options.
type ControllerManager interface {
	Image() string
	ExtraArgs() map[string]string
	ExtraVolumes() []VolumeMount
}

// Proxy defines the requirements for a config that pertains to the kube-proxy
// options.
type Proxy interface {
	Enabled() bool

	Image() string

	// Mode indicates the proxy mode for kube-proxy.  By default, this is `iptables`.  Other options include `ipvs`.
	Mode() string

	// ExtraArgs describe an additional set of arguments to be supplied to the execution of `kube-proxy`
	ExtraArgs() map[string]string
}

// Scheduler defines the requirements for a config that pertains to scheduler related
// options.
type Scheduler interface {
	Image() string
	ExtraArgs() map[string]string
	ExtraVolumes() []VolumeMount
}

// Etcd defines the requirements for a config that pertains to etcd related
// options.
type Etcd interface {
	Image() string
	CA() *x509.PEMEncodedCertificateAndKey
	ExtraArgs() map[string]string
}

// Token defines the requirements for a config that pertains to Kubernetes
// bootstrap token.
type Token interface {
	ID() string
	Secret() string
}

// CoreDNS defines the requirements for a config that pertains to CoreDNS
// coredns options.
type CoreDNS interface {
	Enabled() bool
	Image() string
	Replicas() int
}

// ExternalCloudProvider defines settings for external cloud provider.
type ExternalCloudProvider interface {
	// Enabled returns true if external cloud provider is enabled.
	Enabled() bool
	// ManifestURLs returns external cloud provider manifest URLs if it is enabled.
	ManifestURLs() []string
}

// AdminKubeconfig defines settings for admin kubeconfig.
type AdminKubeconfig interface {
	CertLifetime() time.Duration
}

// EncryptionKey defines settings for the partition encryption key handling.
type EncryptionKey interface {
	Static() EncryptionKeyStatic
	NodeID() EncryptionKeyNodeID
	Slot() int
}

// EncryptionKeyStatic ephemeral encryption key.
type EncryptionKeyStatic interface {
	Key() []byte
}

// EncryptionKeyNodeID deterministically generated encryption key.
type EncryptionKeyNodeID interface{}

// Encryption defines settings for the partition encryption.
type Encryption interface {
	Kind() string
	Cipher() string
	Keys() []EncryptionKey
}

// SystemDiskEncryption accumulates settings for all system partitions encryption.
type SystemDiskEncryption interface {
	Get(label string) Encryption
}

// Features describe individual Talos features that can be switched on or off.
type Features interface {
	RBACEnabled() bool
}

// VolumeMount describes extra volume mount for the static pods.
type VolumeMount interface {
	Name() string
	HostPath() string
	MountPath() string
	ReadOnly() bool
}

// InlineManifest describes inline manifest for the cluster boostrap.
type InlineManifest interface {
	Name() string
	Contents() string
}
