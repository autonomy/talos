// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1alpha1

import (
	"context"
	"crypto/tls"
	stdx509 "crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/go-blockdevice/blockdevice/util/disk"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
)

const (
	// Version is the version string for v1alpha1.
	Version = "v1alpha1"
)

// Version implements the config.Provider interface.
func (c *Config) Version() string {
	return Version
}

// Debug implements the config.Provider interface.
func (c *Config) Debug() bool {
	return c.ConfigDebug
}

// Persist implements the config.Provider interface.
func (c *Config) Persist() bool {
	return c.ConfigPersist
}

// Machine implements the config.Provider interface.
func (c *Config) Machine() config.MachineConfig {
	return c.MachineConfig
}

// Cluster implements the config.Provider interface.
func (c *Config) Cluster() config.ClusterConfig {
	return c.ClusterConfig
}

// String implements the config.Provider interface.
func (c *Config) String(options ...encoder.Option) (string, error) {
	b, err := c.Bytes(options...)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Bytes implements the config.Provider interface.
func (c *Config) Bytes(options ...encoder.Option) ([]byte, error) {
	return encoder.NewEncoder(c, options...).Encode()
}

// ApplyDynamicConfig implements the config.Provider interface.
//nolint:gocyclo
func (c *Config) ApplyDynamicConfig(ctx context.Context, dynamicProvider config.DynamicConfigProvider) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if c.MachineConfig == nil {
		c.MachineConfig = &MachineConfig{}
	}

	hostname, err := dynamicProvider.Hostname(ctx)
	if err != nil {
		return err
	}

	if hostname != nil {
		if c.MachineConfig.MachineNetwork == nil {
			c.MachineConfig.MachineNetwork = &NetworkConfig{}
		}

		c.MachineConfig.MachineNetwork.NetworkHostname = string(hostname)
	}

	addrs, err := dynamicProvider.ExternalIPs(ctx)
	if err != nil {
		log.Printf("certificates will be created without external IPs: %v", err)
	}

	if c.MachineConfig.MachineNetwork != nil {
		for _, nc := range c.MachineConfig.MachineNetwork.NetworkInterfaces {
			if nc.VIPConfig() != nil {
				sharedIP := net.ParseIP(nc.VIPConfig().IP())
				if sharedIP != nil {
					addrs = append(addrs, sharedIP)
				}
			}
		}
	}

	existingSANs := map[string]bool{}
	for _, addr := range c.MachineConfig.MachineCertSANs {
		existingSANs[addr] = true
	}

	sans := make([]string, 0, len(addrs))
	for i, addr := range addrs {
		sans = append(sans, addr.String())

		if existingSANs[sans[i]] {
			continue
		}

		c.MachineConfig.MachineCertSANs = append(c.MachineConfig.MachineCertSANs, sans[i])
	}

	if c.ClusterConfig == nil {
		c.ClusterConfig = &ClusterConfig{}
	}

	if c.ClusterConfig.APIServerConfig == nil {
		c.ClusterConfig.APIServerConfig = &APIServerConfig{}
	}

	existingCertSANs := map[string]bool{}
	for _, certSAN := range c.ClusterConfig.APIServerConfig.CertSANs {
		existingCertSANs[certSAN] = true
	}

	for _, certSAN := range sans {
		if existingCertSANs[certSAN] {
			continue
		}

		c.ClusterConfig.APIServerConfig.CertSANs = append(c.ClusterConfig.APIServerConfig.CertSANs, certSAN)
	}

	return nil
}

// Install implements the config.Provider interface.
func (m *MachineConfig) Install() config.Install {
	if m.MachineInstall == nil {
		return &InstallConfig{}
	}

	return m.MachineInstall
}

// Security implements the config.Provider interface.
func (m *MachineConfig) Security() config.Security {
	return m
}

// Disks implements the config.Provider interface.
func (m *MachineConfig) Disks() []config.Disk {
	disks := make([]config.Disk, len(m.MachineDisks))

	for i := 0; i < len(m.MachineDisks); i++ {
		disks[i] = m.MachineDisks[i]
	}

	return disks
}

// Network implements the config.Provider interface.
func (m *MachineConfig) Network() config.MachineNetwork {
	if m.MachineNetwork == nil {
		return &NetworkConfig{}
	}

	return m.MachineNetwork
}

// Time implements the config.Provider interface.
func (m *MachineConfig) Time() config.Time {
	if m.MachineTime == nil {
		return &TimeConfig{}
	}

	return m.MachineTime
}

// Kubelet implements the config.Provider interface.
func (m *MachineConfig) Kubelet() config.Kubelet {
	if m.MachineKubelet == nil {
		return &KubeletConfig{}
	}

	return m.MachineKubelet
}

// Env implements the config.Provider interface.
func (m *MachineConfig) Env() config.Env {
	return m.MachineEnv
}

// Files implements the config.Provider interface.
func (m *MachineConfig) Files() ([]config.File, error) {
	files := make([]config.File, len(m.MachineFiles))

	for i := 0; i < len(m.MachineFiles); i++ {
		files[i] = m.MachineFiles[i]
	}

	return files, nil
}

// Type implements the config.Provider interface.
func (m *MachineConfig) Type() machine.Type {
	t, _ := machine.ParseType(m.MachineType) //nolint:errcheck

	return t
}

// Server implements the config.Provider interface.
func (m *MachineConfig) Server() string {
	return ""
}

// Sysctls implements the config.Provider interface.
func (m *MachineConfig) Sysctls() map[string]string {
	if m.MachineSysctls == nil {
		return make(map[string]string)
	}

	return m.MachineSysctls
}

// CA implements the config.Provider interface.
func (m *MachineConfig) CA() *x509.PEMEncodedCertificateAndKey {
	return m.MachineCA
}

// Token implements the config.Provider interface.
func (m *MachineConfig) Token() string {
	return m.MachineToken
}

// CertSANs implements the config.Provider interface.
func (m *MachineConfig) CertSANs() []string {
	return m.MachineCertSANs
}

// Registries implements the config.Provider interface.
func (m *MachineConfig) Registries() config.Registries {
	return &m.MachineRegistries
}

// SystemDiskEncryption implements the config.Provider interface.
func (m *MachineConfig) SystemDiskEncryption() config.SystemDiskEncryption {
	if m.MachineSystemDiskEncryption == nil {
		return &SystemDiskEncryptionConfig{}
	}

	return m.MachineSystemDiskEncryption
}

// Features implements the config.MachineConfig interface.
func (m *MachineConfig) Features() config.Features {
	if m.MachineFeatures == nil {
		return &FeaturesConfig{}
	}

	return m.MachineFeatures
}

// Image implements the config.Provider interface.
func (k *KubeletConfig) Image() string {
	image := k.KubeletImage

	if image == "" {
		image = fmt.Sprintf("%s:v%s", constants.KubeletImage, constants.DefaultKubernetesVersion)
	}

	return image
}

// ClusterDNS implements the config.Provider interface.
func (k *KubeletConfig) ClusterDNS() []string {
	if k == nil || k.KubeletClusterDNS == nil {
		return nil
	}

	return k.KubeletClusterDNS
}

// ExtraArgs implements the config.Provider interface.
func (k *KubeletConfig) ExtraArgs() map[string]string {
	if k == nil || k.KubeletExtraArgs == nil {
		return make(map[string]string)
	}

	return k.KubeletExtraArgs
}

// ExtraMounts implements the config.Provider interface.
func (k *KubeletConfig) ExtraMounts() []specs.Mount {
	if k.KubeletExtraMounts == nil {
		return nil
	}

	out := make([]specs.Mount, len(k.KubeletExtraMounts))

	for i := range k.KubeletExtraMounts {
		out[i] = k.KubeletExtraMounts[i].Mount
	}

	return out
}

// RegisterWithFQDN implements the config.Provider interface.
func (k *KubeletConfig) RegisterWithFQDN() bool {
	return k.KubeletRegisterWithFQDN
}

// Mirrors implements the Registries interface.
func (r *RegistriesConfig) Mirrors() map[string]config.RegistryMirrorConfig {
	mirrors := make(map[string]config.RegistryMirrorConfig, len(r.RegistryMirrors))

	for k, v := range r.RegistryMirrors {
		mirrors[k] = v
	}

	return mirrors
}

// Config implements the Registries interface.
func (r *RegistriesConfig) Config() map[string]config.RegistryConfig {
	registries := make(map[string]config.RegistryConfig, len(r.RegistryConfig))

	for k, v := range r.RegistryConfig {
		registries[k] = v
	}

	return registries
}

// TLS implements the Registries interface.
func (r *RegistryConfig) TLS() config.RegistryTLSConfig {
	if r.RegistryTLS == nil {
		return nil
	}

	return r.RegistryTLS
}

// Auth implements the Registries interface.
func (r *RegistryConfig) Auth() config.RegistryAuthConfig {
	if r.RegistryAuth == nil {
		return nil
	}

	return r.RegistryAuth
}

// Username implements the Registries interface.
func (r *RegistryAuthConfig) Username() string {
	return r.RegistryUsername
}

// Password implements the Registries interface.
func (r *RegistryAuthConfig) Password() string {
	return r.RegistryPassword
}

// Auth implements the Registries interface.
func (r *RegistryAuthConfig) Auth() string {
	return r.RegistryAuth
}

// IdentityToken implements the Registries interface.
func (r *RegistryAuthConfig) IdentityToken() string {
	return r.RegistryIdentityToken
}

// ClientIdentity implements the Registries interface.
func (r *RegistryTLSConfig) ClientIdentity() *x509.PEMEncodedCertificateAndKey {
	return r.TLSClientIdentity
}

// CA implements the Registries interface.
func (r *RegistryTLSConfig) CA() []byte {
	return r.TLSCA
}

// InsecureSkipVerify implements the Registries interface.
func (r *RegistryTLSConfig) InsecureSkipVerify() bool {
	return r.TLSInsecureSkipVerify
}

// GetTLSConfig prepares TLS configuration for connection.
func (r *RegistryTLSConfig) GetTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if r.TLSClientIdentity != nil {
		cert, err := tls.X509KeyPair(r.TLSClientIdentity.Crt, r.TLSClientIdentity.Key)
		if err != nil {
			return nil, fmt.Errorf("error parsing client identity: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if r.CA() != nil {
		tlsConfig.RootCAs = stdx509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(r.TLSCA)
	}

	if r.TLSInsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}

// Hostname implements the config.Provider interface.
func (n *NetworkConfig) Hostname() string {
	return n.NetworkHostname
}

// Devices implements the config.Provider interface.
func (n *NetworkConfig) Devices() []config.Device {
	interfaces := make([]config.Device, len(n.NetworkInterfaces))

	for i := 0; i < len(n.NetworkInterfaces); i++ {
		interfaces[i] = n.NetworkInterfaces[i]
	}

	return interfaces
}

// getDevice adds or returns existing Device by name.
//
// This method mutates configuration, but it's only used in config generation.
func (n *NetworkConfig) getDevice(name string) *Device {
	for _, dev := range n.NetworkInterfaces {
		if dev.DeviceInterface == name {
			return dev
		}
	}

	dev := &Device{
		DeviceInterface: name,
	}

	n.NetworkInterfaces = append(n.NetworkInterfaces, dev)

	return dev
}

// Resolvers implements the config.Provider interface.
func (n *NetworkConfig) Resolvers() []string {
	return n.NameServers
}

// ExtraHosts implements the config.Provider interface.
func (n *NetworkConfig) ExtraHosts() []config.ExtraHost {
	hosts := make([]config.ExtraHost, len(n.ExtraHostEntries))

	for i := 0; i < len(n.ExtraHostEntries); i++ {
		hosts[i] = n.ExtraHostEntries[i]
	}

	return hosts
}

// IP implements the MachineNetwork interface.
func (e *ExtraHost) IP() string {
	return e.HostIP
}

// Aliases implements the MachineNetwork interface.
func (e *ExtraHost) Aliases() []string {
	return e.HostAliases
}

// Interface implements the MachineNetwork interface.
func (d *Device) Interface() string {
	return d.DeviceInterface
}

// CIDR implements the MachineNetwork interface.
func (d *Device) CIDR() string {
	return d.DeviceCIDR
}

// Routes implements the MachineNetwork interface.
func (d *Device) Routes() []config.Route {
	routes := make([]config.Route, len(d.DeviceRoutes))

	for i := 0; i < len(d.DeviceRoutes); i++ {
		routes[i] = d.DeviceRoutes[i]
	}

	return routes
}

// Bond implements the MachineNetwork interface.
func (d *Device) Bond() config.Bond {
	if d.DeviceBond == nil {
		return nil
	}

	return d.DeviceBond
}

// Vlans implements the MachineNetwork interface.
func (d *Device) Vlans() []config.Vlan {
	vlans := make([]config.Vlan, len(d.DeviceVlans))

	for i := 0; i < len(d.DeviceVlans); i++ {
		vlans[i] = d.DeviceVlans[i]
	}

	return vlans
}

// MTU implements the MachineNetwork interface.
func (d *Device) MTU() int {
	return d.DeviceMTU
}

// DHCP implements the MachineNetwork interface.
func (d *Device) DHCP() bool {
	return d.DeviceDHCP
}

// Ignore implements the MachineNetwork interface.
func (d *Device) Ignore() bool {
	return d.DeviceIgnore
}

// Dummy implements the MachineNetwork interface.
func (d *Device) Dummy() bool {
	return d.DeviceDummy
}

// DHCPOptions implements the MachineNetwork interface.
func (d *Device) DHCPOptions() config.DHCPOptions {
	// Default route metric on systemd is 1024. This sets the same.
	if d.DeviceDHCPOptions == nil {
		return &DHCPOptions{
			DHCPRouteMetric: uint32(0),
		}
	}

	return d.DeviceDHCPOptions
}

// VIPConfig implements the MachineNetwork interface.
func (d *Device) VIPConfig() config.VIPConfig {
	if d.DeviceVIPConfig == nil {
		return nil
	}

	return d.DeviceVIPConfig
}

// IP implements the config.VIPConfig interface.
func (d *DeviceVIPConfig) IP() string {
	return d.SharedIP
}

// WireguardConfig implements the MachineNetwork interface.
func (d *Device) WireguardConfig() config.WireguardConfig {
	if d.DeviceWireguardConfig == nil {
		return nil
	}

	return d.DeviceWireguardConfig
}

// RouteMetric implements the DHCPOptions interface.
func (d *DHCPOptions) RouteMetric() uint32 {
	return d.DHCPRouteMetric
}

// IPv4 implements the DHCPOptions interface.
func (d *DHCPOptions) IPv4() bool {
	if d.DHCPIPv4 == nil {
		return true
	}

	return *d.DHCPIPv4
}

// IPv6 implements the DHCPOptions interface.
func (d *DHCPOptions) IPv6() bool {
	if d.DHCPIPv6 == nil {
		return false
	}

	return *d.DHCPIPv6
}

// PrivateKey implements the MachineNetwork interface.
func (wc *DeviceWireguardConfig) PrivateKey() string {
	return wc.WireguardPrivateKey
}

// ListenPort implements the MachineNetwork interface.
func (wc *DeviceWireguardConfig) ListenPort() int {
	return wc.WireguardListenPort
}

// FirewallMark implements the MachineNetwork interface.
func (wc *DeviceWireguardConfig) FirewallMark() int {
	return wc.WireguardFirewallMark
}

// Peers implements the MachineNetwork interface.
func (wc *DeviceWireguardConfig) Peers() []config.WireguardPeer {
	peers := make([]config.WireguardPeer, len(wc.WireguardPeers))

	for i := 0; i < len(wc.WireguardPeers); i++ {
		peers[i] = wc.WireguardPeers[i]
	}

	return peers
}

// PublicKey implements the MachineNetwork interface.
func (wd *DeviceWireguardPeer) PublicKey() string {
	return wd.WireguardPublicKey
}

// Endpoint implements the MachineNetwork interface.
func (wd *DeviceWireguardPeer) Endpoint() string {
	return wd.WireguardEndpoint
}

// PersistentKeepaliveInterval implements the MachineNetwork interface.
func (wd *DeviceWireguardPeer) PersistentKeepaliveInterval() time.Duration {
	return wd.WireguardPersistentKeepaliveInterval
}

// AllowedIPs implements the MachineNetwork interface.
func (wd *DeviceWireguardPeer) AllowedIPs() []string {
	return wd.WireguardAllowedIPs
}

// Network implements the MachineNetwork interface.
func (r *Route) Network() string {
	return r.RouteNetwork
}

// Gateway implements the MachineNetwork interface.
func (r *Route) Gateway() string {
	return r.RouteGateway
}

// Metric implements the MachineNetwork interface.
func (r *Route) Metric() uint32 {
	return r.RouteMetric
}

// Interfaces implements the MachineNetwork interface.
func (b *Bond) Interfaces() []string {
	if b == nil {
		return nil
	}

	return b.BondInterfaces
}

// ARPIPTarget implements the MachineNetwork interface.
func (b *Bond) ARPIPTarget() []string {
	if b == nil {
		return nil
	}

	return b.BondARPIPTarget
}

// Mode implements the MachineNetwork interface.
func (b *Bond) Mode() string {
	return b.BondMode
}

// HashPolicy implements the MachineNetwork interface.
func (b *Bond) HashPolicy() string {
	return b.BondHashPolicy
}

// LACPRate implements the MachineNetwork interface.
func (b *Bond) LACPRate() string {
	return b.BondLACPRate
}

// ADActorSystem implements the MachineNetwork interface.
func (b *Bond) ADActorSystem() string {
	return b.BondADActorSystem
}

// ARPValidate implements the MachineNetwork interface.
func (b *Bond) ARPValidate() string {
	return b.BondARPValidate
}

// ARPAllTargets implements the MachineNetwork interface.
func (b *Bond) ARPAllTargets() string {
	return b.BondARPAllTargets
}

// Primary implements the MachineNetwork interface.
func (b *Bond) Primary() string {
	return b.BondPrimary
}

// PrimaryReselect implements the MachineNetwork interface.
func (b *Bond) PrimaryReselect() string {
	return b.BondPrimaryReselect
}

// FailOverMac implements the MachineNetwork interface.
func (b *Bond) FailOverMac() string {
	return b.BondFailOverMac
}

// ADSelect implements the MachineNetwork interface.
func (b *Bond) ADSelect() string {
	return b.BondADSelect
}

// MIIMon implements the MachineNetwork interface.
func (b *Bond) MIIMon() uint32 {
	return b.BondMIIMon
}

// UpDelay implements the MachineNetwork interface.
func (b *Bond) UpDelay() uint32 {
	return b.BondUpDelay
}

// DownDelay implements the MachineNetwork interface.
func (b *Bond) DownDelay() uint32 {
	return b.BondDownDelay
}

// ARPInterval implements the MachineNetwork interface.
func (b *Bond) ARPInterval() uint32 {
	return b.BondARPInterval
}

// ResendIGMP implements the MachineNetwork interface.
func (b *Bond) ResendIGMP() uint32 {
	return b.BondResendIGMP
}

// MinLinks implements the MachineNetwork interface.
func (b *Bond) MinLinks() uint32 {
	return b.BondMinLinks
}

// LPInterval implements the MachineNetwork interface.
func (b *Bond) LPInterval() uint32 {
	return b.BondLPInterval
}

// PacketsPerSlave implements the MachineNetwork interface.
func (b *Bond) PacketsPerSlave() uint32 {
	return b.BondPacketsPerSlave
}

// NumPeerNotif implements the MachineNetwork interface.
func (b *Bond) NumPeerNotif() uint8 {
	return b.BondNumPeerNotif
}

// TLBDynamicLB implements the MachineNetwork interface.
func (b *Bond) TLBDynamicLB() uint8 {
	return b.BondTLBDynamicLB
}

// AllSlavesActive implements the MachineNetwork interface.
func (b *Bond) AllSlavesActive() uint8 {
	return b.BondAllSlavesActive
}

// UseCarrier implements the MachineNetwork interface.
func (b *Bond) UseCarrier() bool {
	if b.BondUseCarrier == nil {
		return true
	}

	return *b.BondUseCarrier
}

// ADActorSysPrio implements the MachineNetwork interface.
func (b *Bond) ADActorSysPrio() uint16 {
	return b.BondADActorSysPrio
}

// ADUserPortKey implements the MachineNetwork interface.
func (b *Bond) ADUserPortKey() uint16 {
	return b.BondADUserPortKey
}

// PeerNotifyDelay implements the MachineNetwork interface.
func (b *Bond) PeerNotifyDelay() uint32 {
	return b.BondPeerNotifyDelay
}

// CIDR implements the MachineNetwork interface.
func (v *Vlan) CIDR() string {
	return v.VlanCIDR
}

// Routes implements the MachineNetwork interface.
func (v *Vlan) Routes() []config.Route {
	routes := make([]config.Route, len(v.VlanRoutes))

	for i := 0; i < len(v.VlanRoutes); i++ {
		routes[i] = v.VlanRoutes[i]
	}

	return routes
}

// DHCP implements the MachineNetwork interface.
func (v *Vlan) DHCP() bool {
	return v.VlanDHCP
}

// ID implements the MachineNetwork interface.
func (v *Vlan) ID() uint16 {
	return v.VlanID
}

// Disabled implements the config.Provider interface.
func (t *TimeConfig) Disabled() bool {
	return t.TimeDisabled
}

// Servers implements the config.Provider interface.
func (t *TimeConfig) Servers() []string {
	return t.TimeServers
}

// Image implements the config.Provider interface.
func (i *InstallConfig) Image() string {
	return i.InstallImage
}

// Disk implements the config.Provider interface.
func (i *InstallConfig) Disk() (string, error) {
	matchers := i.DiskMatchers()
	if len(matchers) > 0 {
		d, err := disk.Find(matchers...)
		if err != nil {
			return "", err
		}

		if d != nil {
			return d.DeviceName, nil
		}

		return "", fmt.Errorf("no disk found matching provided parameters")
	}

	return i.InstallDisk, nil
}

// DiskMatchers implements the config.Provider interface.
func (i *InstallConfig) DiskMatchers() []disk.Matcher {
	if i.InstallDiskSelector != nil {
		selector := i.InstallDiskSelector

		matchers := []disk.Matcher{}
		if selector.Size != nil {
			matchers = append(matchers, selector.Size.Matcher)
		}

		if selector.UUID != "" {
			matchers = append(matchers, disk.WithUUID(selector.UUID))
		}

		if selector.WWID != "" {
			matchers = append(matchers, disk.WithWWID(selector.WWID))
		}

		if selector.Model != "" {
			matchers = append(matchers, disk.WithModel(selector.Model))
		}

		if selector.Name != "" {
			matchers = append(matchers, disk.WithName(selector.Name))
		}

		if selector.Serial != "" {
			matchers = append(matchers, disk.WithSerial(selector.Serial))
		}

		if selector.Modalias != "" {
			matchers = append(matchers, disk.WithModalias(selector.Modalias))
		}

		if disk.Type(selector.Type) != disk.TypeUnknown {
			matchers = append(matchers, disk.WithType(disk.Type(selector.Type)))
		}

		return matchers
	}

	return nil
}

// ExtraKernelArgs implements the config.Provider interface.
func (i *InstallConfig) ExtraKernelArgs() []string {
	return i.InstallExtraKernelArgs
}

// Zero implements the config.Provider interface.
func (i *InstallConfig) Zero() bool {
	return i.InstallWipe
}

// LegacyBIOSSupport implements the config.Provider interface.
func (i *InstallConfig) LegacyBIOSSupport() bool {
	return i.InstallLegacyBIOSSupport
}

// WithBootloader implements the config.Provider interface.
func (i *InstallConfig) WithBootloader() bool {
	return i.InstallBootloader
}

// Enabled implements the config.Provider interface.
func (c *CoreDNS) Enabled() bool {
	return !c.CoreDNSDisabled
}

// Image implements the config.Provider interface.
func (c *CoreDNS) Image() string {
	coreDNSImage := fmt.Sprintf("%s:%s", constants.CoreDNSImage, constants.DefaultCoreDNSVersion)

	if c.CoreDNSImage != "" {
		coreDNSImage = c.CoreDNSImage
	}

	return coreDNSImage
}

// Replicas implements the config.Provider interface.
func (c *CoreDNS) Replicas() int {
	coreDNSReplicas := constants.DefaultCoreDNSReplicas

	if c.CoreDNSReplicas > 0 {
		coreDNSReplicas = c.CoreDNSReplicas
	}

	return coreDNSReplicas
}

// CertLifetime implements the config.Provider interface.
func (a *AdminKubeconfigConfig) CertLifetime() time.Duration {
	if a.AdminKubeconfigCertLifetime == 0 {
		return constants.KubernetesAdminCertDefaultLifetime
	}

	return a.AdminKubeconfigCertLifetime
}

// Endpoints implements the config.Provider interface.
func (r *RegistryMirrorConfig) Endpoints() []string {
	return r.MirrorEndpoints
}

// Content implements the config.Provider interface.
func (f *MachineFile) Content() string {
	return f.FileContent
}

// Permissions implements the config.Provider interface.
func (f *MachineFile) Permissions() os.FileMode {
	return os.FileMode(f.FilePermissions)
}

// Path implements the config.Provider interface.
func (f *MachineFile) Path() string {
	return f.FilePath
}

// Op implements the config.Provider interface.
func (f *MachineFile) Op() string {
	return f.FileOp
}

// Device implements the config.Provider interface.
func (d *MachineDisk) Device() string {
	return d.DeviceName
}

// Partitions implements the config.Provider interface.
func (d *MachineDisk) Partitions() []config.Partition {
	partitions := make([]config.Partition, len(d.DiskPartitions))

	for i := 0; i < len(d.DiskPartitions); i++ {
		partitions[i] = d.DiskPartitions[i]
	}

	return partitions
}

// Size implements the config.Provider interface.
func (p *DiskPartition) Size() uint64 {
	return uint64(p.DiskSize)
}

// MountPoint implements the config.Provider interface.
func (p *DiskPartition) MountPoint() string {
	return p.DiskMountPoint
}

// Kind implements the config.Provider interface.
func (e *EncryptionConfig) Kind() string {
	return e.EncryptionProvider
}

// Cipher implements the config.Provider interface.
func (e *EncryptionConfig) Cipher() string {
	return e.EncryptionCipher
}

// Keys implements the config.Provider interface.
func (e *EncryptionConfig) Keys() []config.EncryptionKey {
	keys := make([]config.EncryptionKey, len(e.EncryptionKeys))

	for i, key := range e.EncryptionKeys {
		keys[i] = key
	}

	return keys
}

// Static implements the config.Provider interface.
func (e *EncryptionKey) Static() config.EncryptionKeyStatic {
	if e.KeyStatic == nil {
		return nil
	}

	return e.KeyStatic
}

// NodeID implements the config.Provider interface.
func (e *EncryptionKey) NodeID() config.EncryptionKeyNodeID {
	if e.KeyNodeID == nil {
		return nil
	}

	return e.KeyNodeID
}

// Slot implements the config.Provider interface.
func (e *EncryptionKey) Slot() int {
	return e.KeySlot
}

// Key implements the config.Provider interface.
func (e *EncryptionKeyStatic) Key() []byte {
	return []byte(e.KeyData)
}

// Get implements the config.Provider interface.
func (e *SystemDiskEncryptionConfig) Get(label string) config.Encryption {
	switch label {
	case constants.StatePartitionLabel:
		if e.StatePartition == nil {
			return nil
		}

		return e.StatePartition
	case constants.EphemeralPartitionLabel:
		if e.EphemeralPartition == nil {
			return nil
		}

		return e.EphemeralPartition
	}

	return nil
}

// HostPath implements the config.VolumeMount interface.
func (v VolumeMountConfig) HostPath() string {
	return v.VolumeHostPath
}

// MountPath implements the config.VolumeMount interface.
func (v VolumeMountConfig) MountPath() string {
	return v.VolumeMountPath
}

// Name implements the config.VolumeMount interface.
func (v VolumeMountConfig) Name() string {
	return strings.Trim(strings.ReplaceAll(strings.ReplaceAll(v.VolumeMountPath, "/", "-"), "_", "-"), "-")
}

// ReadOnly implements the config.VolumeMount interface.
func (v VolumeMountConfig) ReadOnly() bool {
	return v.VolumeReadOnly
}
