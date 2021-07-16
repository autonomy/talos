// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package v1alpha1 configuration file contains all the options available for configuring a machine.

To generate a set of basic configuration files, run:

	talosctl gen config --version v1alpha1 <cluster name> <cluster endpoint>

This will generate a machine config for each node type, and a talosconfig for the CLI.
*/
package v1alpha1

//go:generate docgen ./v1alpha1_types.go ./v1alpha1_types_doc.go Configuration

//go:generate deepcopy-gen --input-dirs ../v1alpha1/ --go-header-file ../../../../../hack/boilerplate.txt --bounding-dirs ../v1alpha1 -O zz_generated.deepcopy

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	humanize "github.com/dustin/go-humanize"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/go-blockdevice/blockdevice/util/disk"
	yaml "gopkg.in/yaml.v3"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
)

func init() {
	config.Register("v1alpha1", func(version string) (target interface{}) {
		target = &Config{}

		return target
	})
}

func mustParseURL(uri string) *url.URL {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	return u
}

var (
	// Examples section.

	// this is using custom type to avoid generating full example with all the nested structs.
	configExample = struct {
		Version string `yaml:"version"`
		Persist bool
		Machine *yaml.Node
		Cluster *yaml.Node
	}{
		Version: "v1alpha1",
		Persist: true,
		Machine: &yaml.Node{Kind: yaml.ScalarNode, LineComment: "..."},
		Cluster: &yaml.Node{Kind: yaml.ScalarNode, LineComment: "..."},
	}

	machineConfigExample = struct {
		Type    string
		Install *InstallConfig
	}{
		Type:    machine.TypeControlPlane.String(),
		Install: machineInstallExample,
	}

	machineConfigRegistriesExample = &RegistriesConfig{
		RegistryMirrors: map[string]*RegistryMirrorConfig{
			"docker.io": {
				MirrorEndpoints: []string{"https://registry.local"},
			},
		},
		RegistryConfig: map[string]*RegistryConfig{
			"registry.local": {
				RegistryTLS: &RegistryTLSConfig{
					TLSClientIdentity: pemEncodedCertificateExample,
				},
				RegistryAuth: &RegistryAuthConfig{
					RegistryUsername: "username",
					RegistryPassword: "password",
				},
			},
		},
	}

	machineConfigRegistryMirrorsExample = map[string]*RegistryMirrorConfig{
		"ghcr.io": {
			MirrorEndpoints: []string{"https://registry.insecure", "https://ghcr.io/v2/"},
		},
	}

	machineConfigRegistryConfigExample = map[string]*RegistryConfig{
		"registry.insecure": {
			RegistryTLS: &RegistryTLSConfig{
				TLSInsecureSkipVerify: true,
			},
		},
	}

	machineConfigRegistryTLSConfigExample1 = &RegistryTLSConfig{
		TLSClientIdentity: pemEncodedCertificateExample,
	}

	machineConfigRegistryTLSConfigExample2 = &RegistryTLSConfig{
		TLSInsecureSkipVerify: true,
	}

	machineConfigRegistryAuthConfigExample = &RegistryAuthConfig{
		RegistryUsername: "username",
		RegistryPassword: "password",
	}

	pemEncodedCertificateExample *x509.PEMEncodedCertificateAndKey = &x509.PEMEncodedCertificateAndKey{
		Crt: []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJIekNCMHF..."),
		Key: []byte("LS0tLS1CRUdJTiBFRDI1NTE5IFBSSVZBVEUgS0VZLS0tLS0KTUM..."),
	}

	pemEncodedKeyExample *x509.PEMEncodedKey = &x509.PEMEncodedKey{
		Key: []byte("LS0tLS1CRUdJTiBFRDI1NTE5IFBSSVZBVEUgS0VZLS0tLS0KTUM..."),
	}

	machineKubeletExample = &KubeletConfig{
		KubeletImage: (&KubeletConfig{}).Image(),
		KubeletExtraArgs: map[string]string{
			"feature-gates": "ServerSideApply=true",
		},
	}

	kubeletImageExample = (&KubeletConfig{}).Image()

	machineNetworkConfigExample = &NetworkConfig{
		NetworkHostname: "worker-1",
		NetworkInterfaces: []*Device{
			{
				DeviceInterface: "eth0",
				DeviceCIDR:      "192.168.2.0/24",
				DeviceMTU:       1500,
				DeviceRoutes: []*Route{
					{
						RouteNetwork: "0.0.0.0/0",
						RouteGateway: "192.168.2.1",
						RouteMetric:  1024,
					},
				},
			},
		},
		NameServers: []string{"9.8.7.6", "8.7.6.5"},
	}

	machineDisksExample = []*MachineDisk{
		{
			DeviceName: "/dev/sdb",
			DiskPartitions: []*DiskPartition{
				{
					DiskMountPoint: "/var/mnt/extra",
				},
			},
		},
	}

	machineInstallExample = &InstallConfig{
		InstallDisk:            "/dev/sda",
		InstallExtraKernelArgs: []string{"console=ttyS1", "panic=10"},
		InstallImage:           "ghcr.io/talos-systems/installer:latest",
		InstallBootloader:      true,
		InstallWipe:            false,
	}

	machineInstallDiskSelectorExample = &InstallDiskSelector{
		Model: "WDC*",
		Size: &InstallDiskSizeMatcher{
			condition: ">= 1TB",
		},
	}

	machineInstallDiskSizeMatcherExamples = []*InstallDiskSizeMatcher{
		{
			condition: "4GB",
		},
		{
			condition: "> 1TB",
		},
		{
			condition: "<= 2TB",
		},
	}

	machineFilesExample = []*MachineFile{
		{
			FileContent:     "...",
			FilePermissions: 0o666,
			FilePath:        "/tmp/file.txt",
			FileOp:          "append",
		},
	}

	machineEnvExamples = []Env{
		{
			"GRPC_GO_LOG_VERBOSITY_LEVEL": "99",
			"GRPC_GO_LOG_SEVERITY_LEVEL":  "info",
			"https_proxy":                 "http://SERVER:PORT/",
		},
		{
			"GRPC_GO_LOG_SEVERITY_LEVEL": "error",
			"https_proxy":                "https://USERNAME:PASSWORD@SERVER:PORT/",
		},
		{
			"https_proxy": "http://DOMAIN\\USERNAME:PASSWORD@SERVER:PORT/",
		},
	}

	machineTimeExample = &TimeConfig{
		TimeServers: []string{"time.cloudflare.com"},
	}

	machineSysctlsExample = map[string]string{
		"kernel.domainname":   "talos.dev",
		"net.ipv4.ip_forward": "0",
	}

	machineSystemDiskEncryptionExample = &SystemDiskEncryptionConfig{
		EphemeralPartition: &EncryptionConfig{
			EncryptionProvider: "luks2",
			EncryptionKeys: []*EncryptionKey{
				{
					KeyNodeID: &EncryptionKeyNodeID{},
					KeySlot:   0,
				},
			},
		},
	}

	machineFeaturesExample = &FeaturesConfig{
		RBAC: pointer.ToBool(true),
	}

	clusterConfigExample = struct {
		ControlPlane *ControlPlaneConfig   `yaml:"controlPlane"`
		ClusterName  string                `yaml:"clusterName"`
		Network      *ClusterNetworkConfig `yaml:"network"`
	}{
		ControlPlane: clusterControlPlaneExample,
		ClusterName:  "talos.local",
		Network:      clusterNetworkExample,
	}

	clusterControlPlaneExample = &ControlPlaneConfig{
		Endpoint: &Endpoint{
			&url.URL{
				Host:   "1.2.3.4",
				Scheme: "https",
			},
		},
		LocalAPIServerPort: 443,
	}

	clusterNetworkExample = &ClusterNetworkConfig{
		CNI: &CNIConfig{
			CNIName: constants.FlannelCNI,
		},
		DNSDomain:     "cluster.local",
		PodSubnet:     []string{"10.244.0.0/16"},
		ServiceSubnet: []string{"10.96.0.0/12"},
	}

	clusterAPIServerExample = &APIServerConfig{
		ContainerImage: (&APIServerConfig{}).Image(),
		ExtraArgsConfig: map[string]string{
			"feature-gates":                    "ServerSideApply=true",
			"http2-max-streams-per-connection": "32",
		},
		CertSANs: []string{
			"1.2.3.4",
			"4.5.6.7",
		},
	}

	clusterAPIServerImageExample = (&APIServerConfig{}).Image()

	clusterControllerManagerExample = &ControllerManagerConfig{
		ContainerImage: (&ControllerManagerConfig{}).Image(),
		ExtraArgsConfig: map[string]string{
			"feature-gates": "ServerSideApply=true",
		},
	}

	clusterControllerManagerImageExample = (&ControllerManagerConfig{}).Image()

	clusterProxyExample = &ProxyConfig{
		ContainerImage: (&ProxyConfig{}).Image(),
		ExtraArgsConfig: map[string]string{
			"proxy-mode": "iptables",
		},
		ModeConfig: "ipvs",
	}

	clusterProxyImageExample = (&ProxyConfig{}).Image()

	clusterSchedulerExample = &SchedulerConfig{
		ContainerImage: (&SchedulerConfig{}).Image(),
		ExtraArgsConfig: map[string]string{
			"feature-gates": "AllBeta=true",
		},
	}

	clusterSchedulerImageExample = (&SchedulerConfig{}).Image()

	clusterEtcdExample = &EtcdConfig{
		ContainerImage: (&EtcdConfig{}).Image(),
		EtcdExtraArgs: map[string]string{
			"election-timeout": "5000",
		},
		RootCA: pemEncodedCertificateExample,
	}

	clusterEtcdImageExample = (&EtcdConfig{}).Image()

	clusterCoreDNSExample = &CoreDNS{
		CoreDNSImage:    (&CoreDNS{}).Image(),
		CoreDNSReplicas: 2,
	}

	clusterExternalCloudProviderConfigExample = &ExternalCloudProviderConfig{
		ExternalEnabled: true,
		ExternalManifests: []string{
			"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml",
			"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml",
		},
	}

	clusterAdminKubeconfigExample = &AdminKubeconfigConfig{
		AdminKubeconfigCertLifetime: time.Hour,
	}

	clusterEndpointExample1 = &Endpoint{
		mustParseURL("https://1.2.3.4:6443"),
	}

	clusterEndpointExample2 = &Endpoint{
		mustParseURL("https://cluster1.internal:6443"),
	}

	kubeletExtraMountsExample = []specs.Mount{
		{
			Source:      "/var/lib/example",
			Destination: "/var/lib/example",
			Type:        "bind",
			Options: []string{
				"rshared",
				"rw",
			},
		},
	}

	networkConfigExtraHostsExample = []*ExtraHost{
		{
			HostIP: "192.168.1.100",
			HostAliases: []string{
				"example",
				"example.domain.tld",
			},
		},
	}

	networkConfigRoutesExample = []*Route{
		{
			RouteNetwork: "0.0.0.0/0",
			RouteGateway: "10.5.0.1",
		},
		{
			RouteNetwork: "10.2.0.0/16",
			RouteGateway: "10.2.0.1",
		},
	}

	networkConfigBondExample = &Bond{
		BondMode:       "802.3ad",
		BondLACPRate:   "fast",
		BondInterfaces: []string{"eth0", "eth1"},
	}

	networkConfigDHCPOptionsExample = &DHCPOptions{
		DHCPRouteMetric: 1024,
	}

	networkConfigVIPLayer2Example = &DeviceVIPConfig{
		SharedIP: "172.16.199.55",
	}

	networkConfigWireguardHostExample = &DeviceWireguardConfig{
		WireguardPrivateKey: "ABCDEF...",
		WireguardListenPort: 51111,
		WireguardPeers: []*DeviceWireguardPeer{
			{
				WireguardPublicKey:  "ABCDEF...",
				WireguardEndpoint:   "192.168.1.3",
				WireguardAllowedIPs: []string{"192.168.1.0/24"},
			},
		},
	}

	networkConfigWireguardPeerExample = &DeviceWireguardConfig{
		WireguardPrivateKey: "ABCDEF...",
		WireguardPeers: []*DeviceWireguardPeer{
			{
				WireguardPublicKey:                   "ABCDEF...",
				WireguardEndpoint:                    "192.168.1.2",
				WireguardPersistentKeepaliveInterval: time.Second * 10,
				WireguardAllowedIPs:                  []string{"192.168.1.0/24"},
			},
		},
	}

	clusterCustomCNIExample = &CNIConfig{
		CNIName: constants.CustomCNI,
		CNIUrls: []string{
			"https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml",
		},
	}

	clusterInlineManifestsExample = ClusterInlineManifests{
		{
			InlineManifestName: "namespace-ci",
			InlineManifestContents: strings.TrimSpace(`
apiVersion: v1
kind: Namespace
metadata:
	name: ci
`),
		},
	}
)

// Config defines the v1alpha1 configuration file.
//
//  examples:
//     - value: configExample
type Config struct {
	//   description: |
	//     Indicates the schema used to decode the contents.
	//   values:
	//     - "v1alpha1"
	ConfigVersion string `yaml:"version"`
	//   description: |
	//     Enable verbose logging to the console.
	//     All system containers logs will flow into serial console.
	//
	//     > Note: To avoid breaking Talos bootstrap flow enable this option only if serial console can handle high message throughput.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	ConfigDebug bool `yaml:"debug"`
	//   description: |
	//     Indicates whether to pull the machine config upon every boot.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	ConfigPersist bool `yaml:"persist"`
	//   description: |
	//     Provides machine specific configuration options.
	MachineConfig *MachineConfig `yaml:"machine"`
	//   description: |
	//     Provides cluster specific configuration options.
	ClusterConfig *ClusterConfig `yaml:"cluster"`
}

// MachineConfig represents the machine-specific config values.
//
//  examples:
//     - value: machineConfigExample
type MachineConfig struct {
	//   description: |
	//     Defines the role of the machine within the cluster.
	//
	//     #### Init
	//
	//     Init node type designates the first control plane node to come up.
	//     You can think of it like a bootstrap node.
	//     This node will perform the initial steps to bootstrap the cluster -- generation of TLS assets, starting of the control plane, etc.
	//
	//     #### Control Plane
	//
	//     Control Plane node type designates the node as a control plane member.
	//     This means it will host etcd along with the Kubernetes master components such as API Server, Controller Manager, Scheduler.
	//
	//     #### Worker
	//
	//     Worker node type designates the node as a worker node.
	//     This means it will be an available compute node for scheduling workloads.
	//
	//     This node type was previously known as "join"; that value is still supported but deprecated.
	//   values:
	//     - "init"
	//     - "controlplane"
	//     - "worker"
	MachineType string `yaml:"type"`
	//   description: |
	//     The `token` is used by a machine to join the PKI of the cluster.
	//     Using this token, a machine will create a certificate signing request (CSR), and request a certificate that will be used as its' identity.
	//   examples:
	//     - name: example token
	//       value: "\"328hom.uqjzh6jnn2eie9oi\""
	MachineToken string `yaml:"token"` // Warning: It is important to ensure that this token is correct since a machine's certificate has a short TTL by default.
	//   description: |
	//     The root certificate authority of the PKI.
	//     It is composed of a base64 encoded `crt` and `key`.
	//   examples:
	//     - value: pemEncodedCertificateExample
	//       name: machine CA example
	MachineCA *x509.PEMEncodedCertificateAndKey `yaml:"ca,omitempty"`
	//   description: |
	//     Extra certificate subject alternative names for the machine's certificate.
	//     By default, all non-loopback interface IPs are automatically added to the certificate's SANs.
	//   examples:
	//     - name: Uncomment this to enable SANs.
	//       value: '[]string{"10.0.0.10", "172.16.0.10", "192.168.0.10"}'
	MachineCertSANs []string `yaml:"certSANs"`
	//   description: |
	//     Used to provide additional options to the kubelet.
	//   examples:
	//     - name: Kubelet definition example.
	//       value: machineKubeletExample
	MachineKubelet *KubeletConfig `yaml:"kubelet,omitempty"`
	//   description: |
	//     Provides machine specific network configuration options.
	//   examples:
	//     - name: Network definition example.
	//       value: machineNetworkConfigExample
	MachineNetwork *NetworkConfig `yaml:"network,omitempty"`
	//   description: |
	//     Used to partition, format and mount additional disks.
	//     Since the rootfs is read only with the exception of `/var`, mounts are only valid if they are under `/var`.
	//     Note that the partitioning and formating is done only once, if and only if no existing partitions are found.
	//     If `size:` is omitted, the partition is sized to occupy the full disk.
	//   examples:
	//     - name: MachineDisks list example.
	//       value: machineDisksExample
	MachineDisks []*MachineDisk `yaml:"disks,omitempty"` // Note: `size` is in units of bytes.
	//   description: |
	//     Used to provide instructions for installations.
	//   examples:
	//     - name: MachineInstall config usage example.
	//       value: machineInstallExample
	MachineInstall *InstallConfig `yaml:"install,omitempty"`
	//   description: |
	//     Allows the addition of user specified files.
	//     The value of `op` can be `create`, `overwrite`, or `append`.
	//     In the case of `create`, `path` must not exist.
	//     In the case of `overwrite`, and `append`, `path` must be a valid file.
	//     If an `op` value of `append` is used, the existing file will be appended.
	//     Note that the file contents are not required to be base64 encoded.
	//   examples:
	//      - name: MachineFiles usage example.
	//        value: machineFilesExample
	MachineFiles []*MachineFile `yaml:"files,omitempty"` // Note: The specified `path` is relative to `/var`.
	//   description: |
	//     The `env` field allows for the addition of environment variables.
	//     All environment variables are set on PID 1 in addition to every service.
	//   values:
	//     - "`GRPC_GO_LOG_VERBOSITY_LEVEL`"
	//     - "`GRPC_GO_LOG_SEVERITY_LEVEL`"
	//     - "`http_proxy`"
	//     - "`https_proxy`"
	//     - "`no_proxy`"
	//   examples:
	//     - name: Environment variables definition examples.
	//       value: machineEnvExamples[0]
	//     - value: machineEnvExamples[1]
	//     - value: machineEnvExamples[2]
	MachineEnv Env `yaml:"env,omitempty"`
	//   description: |
	//     Used to configure the machine's time settings.
	//   examples:
	//     - name: Example configuration for cloudflare ntp server.
	//       value: machineTimeExample
	MachineTime *TimeConfig `yaml:"time,omitempty"`
	//   description: |
	//     Used to configure the machine's sysctls.
	//   examples:
	//     - name: MachineSysctls usage example.
	//       value: machineSysctlsExample
	MachineSysctls map[string]string `yaml:"sysctls,omitempty"`
	//   description: |
	//     Used to configure the machine's container image registry mirrors.
	//
	//     Automatically generates matching CRI configuration for registry mirrors.
	//
	//     The `mirrors` section allows to redirect requests for images to non-default registry,
	//     which might be local registry or caching mirror.
	//
	//     The `config` section provides a way to authenticate to the registry with TLS client
	//     identity, provide registry CA, or authentication information.
	//     Authentication information has same meaning with the corresponding field in `.docker/config.json`.
	//
	//     See also matching configuration for [CRI containerd plugin](https://github.com/containerd/cri/blob/master/docs/registry.md).
	//   examples:
	//     - value: machineConfigRegistriesExample
	MachineRegistries RegistriesConfig `yaml:"registries,omitempty"`
	//   description: |
	//     Machine system disk encryption configuration.
	//     Defines each system partition encryption parameters.
	//   examples:
	//     - value: machineSystemDiskEncryptionExample
	MachineSystemDiskEncryption *SystemDiskEncryptionConfig `yaml:"systemDiskEncryption,omitempty"`
	//   description: |
	//     Features describe individual Talos features that can be switched on or off.
	//   examples:
	//     - value: machineFeaturesExample
	MachineFeatures *FeaturesConfig `yaml:"features,omitempty"`
}

// ClusterConfig represents the cluster-wide config values.
//
//  examples:
//     - value: clusterConfigExample
type ClusterConfig struct {
	//   description: |
	//     Provides control plane specific configuration options.
	//   examples:
	//     - name: Setting controlplane endpoint address to 1.2.3.4 and port to 443 example.
	//       value: clusterControlPlaneExample
	ControlPlane *ControlPlaneConfig `yaml:"controlPlane"`
	//   description: |
	//     Configures the cluster's name.
	ClusterName string `yaml:"clusterName,omitempty"`
	//   description: |
	//     Provides cluster specific network configuration options.
	//   examples:
	//     - name: Configuring with flannel CNI and setting up subnets.
	//       value:  clusterNetworkExample
	ClusterNetwork *ClusterNetworkConfig `yaml:"network,omitempty"`
	//   description: |
	//     The [bootstrap token](https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/) used to join the cluster.
	//   examples:
	//     - name: Bootstrap token example (do not use in production!).
	//       value: '"wlzjyw.bei2zfylhs2by0wd"'
	BootstrapToken string `yaml:"token,omitempty"`
	//   description: |
	//     The key used for the [encryption of secret data at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/).
	//   examples:
	//     - name: Decryption secret example (do not use in production!).
	//       value: '"z01mye6j16bspJYtTB/5SFX8j7Ph4JXxM2Xuu4vsBPM="'
	ClusterAESCBCEncryptionSecret string `yaml:"aescbcEncryptionSecret"`
	//   description: |
	//     The base64 encoded root certificate authority used by Kubernetes.
	//   examples:
	//     - name: ClusterCA example.
	//       value: pemEncodedCertificateExample
	ClusterCA *x509.PEMEncodedCertificateAndKey `yaml:"ca,omitempty"`
	//   description: |
	//     The base64 encoded aggregator certificate authority used by Kubernetes for front-proxy certificate generation.
	//
	//     This CA can be self-signed.
	//   examples:
	//     - name: AggregatorCA example.
	//       value: pemEncodedCertificateExample
	ClusterAggregatorCA *x509.PEMEncodedCertificateAndKey `yaml:"aggregatorCA,omitempty"`
	//   description: |
	//     The base64 encoded private key for service account token generation.
	//   examples:
	//     - name: AggregatorCA example.
	//       value: pemEncodedKeyExample
	ClusterServiceAccount *x509.PEMEncodedKey `yaml:"serviceAccount,omitempty"`
	//   description: |
	//     API server specific configuration options.
	//   examples:
	//     - value: clusterAPIServerExample
	APIServerConfig *APIServerConfig `yaml:"apiServer,omitempty"`
	//   description: |
	//     Controller manager server specific configuration options.
	//   examples:
	//     - value: clusterControllerManagerExample
	ControllerManagerConfig *ControllerManagerConfig `yaml:"controllerManager,omitempty"`
	//   description: |
	//     Kube-proxy server-specific configuration options
	//   examples:
	//     - value: clusterProxyExample
	ProxyConfig *ProxyConfig `yaml:"proxy,omitempty"`
	//   description: |
	//     Scheduler server specific configuration options.
	//   examples:
	//     - value: clusterSchedulerExample
	SchedulerConfig *SchedulerConfig `yaml:"scheduler,omitempty"`
	//   description: |
	//     Etcd specific configuration options.
	//   examples:
	//     - value: clusterEtcdExample
	EtcdConfig *EtcdConfig `yaml:"etcd,omitempty"`
	//   description: |
	//     Core DNS specific configuration options.
	//   examples:
	//     - value: clusterCoreDNSExample
	CoreDNSConfig *CoreDNS `yaml:"coreDNS,omitempty"`
	//   description: |
	//     External cloud provider configuration.
	//   examples:
	//     - value: clusterExternalCloudProviderConfigExample
	ExternalCloudProviderConfig *ExternalCloudProviderConfig `yaml:"externalCloudProvider,omitempty"`
	//   description: |
	//     A list of urls that point to additional manifests.
	//     These will get automatically deployed as part of the bootstrap.
	//   examples:
	//     - value: >
	//        []string{
	//         "https://www.example.com/manifest1.yaml",
	//         "https://www.example.com/manifest2.yaml",
	//        }
	ExtraManifests []string `yaml:"extraManifests,omitempty" talos:"omitonlyifnil"`
	//   description: |
	//     A map of key value pairs that will be added while fetching the extraManifests.
	//   examples:
	//     - value: >
	//         map[string]string{
	//           "Token": "1234567",
	//           "X-ExtraInfo": "info",
	//         }
	ExtraManifestHeaders map[string]string `yaml:"extraManifestHeaders,omitempty"`
	//   description: |
	//     A list of inline Kubernetes manifests.
	//     These will get automatically deployed as part of the bootstrap.
	//   examples:
	//     - value: clusterInlineManifestsExample
	ClusterInlineManifests ClusterInlineManifests `yaml:"inlineManifests,omitempty" talos:"omitonlyifnil"`
	//   description: |
	//     Settings for admin kubeconfig generation.
	//     Certificate lifetime can be configured.
	//   examples:
	//     - value: clusterAdminKubeconfigExample
	AdminKubeconfigConfig *AdminKubeconfigConfig `yaml:"adminKubeconfig,omitempty"`
	//   description: |
	//     Allows running workload on master nodes.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	AllowSchedulingOnMasters bool `yaml:"allowSchedulingOnMasters,omitempty"`
}

// ExtraMount wraps OCI Mount specification.
type ExtraMount struct {
	specs.Mount
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtraMount) DeepCopyInto(out *ExtraMount) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtraMount.
func (in *ExtraMount) DeepCopy() *ExtraMount {
	if in == nil {
		return nil
	}

	out := new(ExtraMount)
	in.DeepCopyInto(out)

	return out
}

// KubeletConfig represents the kubelet config values.
type KubeletConfig struct {
	//   description: |
	//     The `image` field is an optional reference to an alternative kubelet image.
	//   examples:
	//     - value: kubeletImageExample
	KubeletImage string `yaml:"image,omitempty"`
	//   description: |
	//     The `ClusterDNS` field is an optional reference to an alternative kubelet clusterDNS ip list.
	//   examples:
	//     - value: '[]string{"10.96.0.10", "169.254.2.53"}'
	KubeletClusterDNS []string `yaml:"clusterDNS,omitempty"`
	//   description: |
	//     The `extraArgs` field is used to provide additional flags to the kubelet.
	//   examples:
	//     - value: >
	//         map[string]string{
	//           "key": "value",
	//         }
	KubeletExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
	//   description: |
	//     The `extraMounts` field is used to add additional mounts to the kubelet container.
	//   examples:
	//     - value: kubeletExtraMountsExample
	KubeletExtraMounts []ExtraMount `yaml:"extraMounts,omitempty"`
	//   description: |
	//     The `registerWithFQDN` field is used to force kubelet to use the node FQDN for registration.
	//     This is required in clouds like AWS.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	KubeletRegisterWithFQDN bool `yaml:"registerWithFQDN,omitempty"`
}

// NetworkConfig represents the machine's networking config values.
type NetworkConfig struct {
	//   description: |
	//     Used to statically set the hostname for the machine.
	NetworkHostname string `yaml:"hostname,omitempty"`
	//   description: |
	//     `interfaces` is used to define the network interface configuration.
	//     By default all network interfaces will attempt a DHCP discovery.
	//     This can be further tuned through this configuration parameter.
	//   examples:
	//     - value: machineNetworkConfigExample.NetworkInterfaces
	NetworkInterfaces []*Device `yaml:"interfaces,omitempty"`
	//   description: |
	//     Used to statically set the nameservers for the machine.
	//     Defaults to `1.1.1.1` and `8.8.8.8`
	//   examples:
	//     - value: '[]string{"8.8.8.8", "1.1.1.1"}'
	NameServers []string `yaml:"nameservers,omitempty"`
	//   description: |
	//     Allows for extra entries to be added to the `/etc/hosts` file
	//   examples:
	//     - value: networkConfigExtraHostsExample
	ExtraHostEntries []*ExtraHost `yaml:"extraHostEntries,omitempty"`
}

// InstallConfig represents the installation options for preparing a node.
type InstallConfig struct {
	//   description: |
	//     The disk used for installations.
	//   examples:
	//     - value: '"/dev/sda"'
	//     - value: '"/dev/nvme0"'
	InstallDisk string `yaml:"disk,omitempty"`
	//   description: |
	//     Look up disk using disk characteristics like model, size, serial and others.
	//     Always has priority over `disk`.
	//   examples:
	//     - value: machineInstallDiskSelectorExample
	InstallDiskSelector *InstallDiskSelector `yaml:"diskSelector,omitempty"`
	//   description: |
	//     Allows for supplying extra kernel args via the bootloader.
	//   examples:
	//     - value: '[]string{"talos.platform=metal", "reboot=k"}'
	InstallExtraKernelArgs []string `yaml:"extraKernelArgs,omitempty"`
	//   description: |
	//     Allows for supplying the image used to perform the installation.
	//     Image reference for each Talos release can be found on
	//     [GitHub releases page](https://github.com/talos-systems/talos/releases).
	//   examples:
	//     - value: '"ghcr.io/talos-systems/installer:latest"'
	InstallImage string `yaml:"image,omitempty"`
	//   description: |
	//     Indicates if a bootloader should be installed.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	InstallBootloader bool `yaml:"bootloader,omitempty"`
	//   description: |
	//     Indicates if the installation disk should be wiped at installation time.
	//     Defaults to `true`.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	InstallWipe bool `yaml:"wipe"`
	//   description: |
	//     Indicates if MBR partition should be marked as bootable (active).
	//     Should be enabled only for the systems with legacy BIOS that doesn't support GPT partitioning scheme.
	InstallLegacyBIOSSupport bool `yaml:"legacyBIOSSupport,omitempty"`
}

// InstallDiskSizeMatcher disk size condition parser.
type InstallDiskSizeMatcher struct {
	// docgen:nodoc
	Matcher disk.Matcher
	// docgen:nodoc
	condition string
}

// DeepCopyInto implements DeepCopy interface.
func (m *InstallDiskSizeMatcher) DeepCopyInto(out *InstallDiskSizeMatcher) {
	*out = *m
}

// MarshalYAML is a custom marshaller for `InstallDiskSizeMatcher`.
func (m *InstallDiskSizeMatcher) MarshalYAML() (interface{}, error) {
	return m.condition, nil
}

// UnmarshalYAML is a custom unmarshaller for `InstallDiskSizeMatcher`.
func (m *InstallDiskSizeMatcher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&m.condition); err != nil {
		return err
	}

	m.condition = strings.TrimSpace(m.condition)

	re := regexp.MustCompile(`(>=|<=|>|<|==)?\b*(.*)$`)

	parts := re.FindStringSubmatch(m.condition)
	if len(parts) < 2 {
		return fmt.Errorf("failed to parse the condition: expected [>=|<=|>|<|==]<size>[units], got %s", m.condition)
	}

	var compare func(*disk.Disk, uint64) bool

	switch parts[1] {
	case ">=":
		compare = func(d *disk.Disk, size uint64) bool {
			return d.Size >= size
		}
	case "<=":
		compare = func(d *disk.Disk, size uint64) bool {
			return d.Size <= size
		}
	case ">":
		compare = func(d *disk.Disk, size uint64) bool {
			return d.Size > size
		}
	case "<":
		compare = func(d *disk.Disk, size uint64) bool {
			return d.Size < size
		}
	case "":
		fallthrough
	case "==":
		compare = func(d *disk.Disk, size uint64) bool {
			return d.Size == size
		}
	default:
		return fmt.Errorf("unknown binary operator %s", parts[1])
	}

	size, err := humanize.ParseBytes(strings.TrimSpace(parts[2]))
	if err != nil {
		return fmt.Errorf("failed to parse disk size %s: %s", parts[2], err)
	}

	m.Matcher = func(d *disk.Disk) bool {
		return compare(d, size)
	}

	return nil
}

// InstallDiskType custom type for disk type selector.
type InstallDiskType disk.Type

// MarshalYAML is a custom marshaller for `InstallDiskSizeMatcher`.
func (it InstallDiskType) MarshalYAML() (interface{}, error) {
	return disk.Type(it).String(), nil
}

// UnmarshalYAML is a custom unmarshaller for `Endpoint`.
func (it *InstallDiskType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		t   string
		err error
	)

	if err = unmarshal(&t); err != nil {
		return err
	}

	if dt, err := disk.ParseType(t); err == nil {
		*it = InstallDiskType(dt)
	} else {
		return err
	}

	return nil
}

// InstallDiskSelector represents a disk query parameters for the install disk lookup.
type InstallDiskSelector struct {
	//   description: Disk size.
	//   examples:
	//     - name: Select a disk which size is equal to 4GB.
	//       value: machineInstallDiskSizeMatcherExamples[0]
	//     - name: Select a disk which size is greater than 1TB.
	//       value: machineInstallDiskSizeMatcherExamples[1]
	//     - name: Select a disk which size is less or equal than 2TB.
	//       value: machineInstallDiskSizeMatcherExamples[2]
	Size *InstallDiskSizeMatcher `yaml:"size,omitempty"`
	//   description: Disk name `/sys/block/<dev>/device/name`.
	Name string `yaml:"name,omitempty"`
	//   description: Disk model `/sys/block/<dev>/device/model`.
	Model string `yaml:"model,omitempty"`
	//   description: Disk serial number `/sys/block/<dev>/serial`.
	Serial string `yaml:"serial,omitempty"`
	//   description: Disk modalias `/sys/block/<dev>/device/modalias`.
	Modalias string `yaml:"modalias,omitempty"`
	//   description: Disk UUID `/sys/block/<dev>/uuid`.
	UUID string `yaml:"uuid,omitempty"`
	//   description: Disk WWID `/sys/block/<dev>/wwid`.
	WWID string `yaml:"wwid,omitempty"`
	//   description: Disk Type.
	//   values:
	//     - ssd
	//     - hdd
	//     - nvme
	//     - sd
	Type InstallDiskType `yaml:"type,omitempty"`
}

// TimeConfig represents the options for configuring time on a machine.
type TimeConfig struct {
	//   description: |
	//     Indicates if the time service is disabled for the machine.
	//     Defaults to `false`.
	TimeDisabled bool `yaml:"disabled"`
	//   description: |
	//     Specifies time (NTP) servers to use for setting the system time.
	//     Defaults to `pool.ntp.org`
	TimeServers []string `yaml:"servers,omitempty"` // This parameter only supports a single time server.
}

// RegistriesConfig represents the image pull options.
type RegistriesConfig struct {
	//   description: |
	//     Specifies mirror configuration for each registry.
	//     This setting allows to use local pull-through caching registires,
	//     air-gapped installations, etc.
	//
	//     Registry name is the first segment of image identifier, with 'docker.io'
	//     being default one.
	//     To catch any registry names not specified explicitly, use '*'.
	//   examples:
	//     - value: machineConfigRegistryMirrorsExample
	RegistryMirrors map[string]*RegistryMirrorConfig `yaml:"mirrors,omitempty"`
	//   description: |
	//     Specifies TLS & auth configuration for HTTPS image registries.
	//     Mutual TLS can be enabled with 'clientIdentity' option.
	//
	//     TLS configuration can be skipped if registry has trusted
	//     server certificate.
	//   examples:
	//     - value: machineConfigRegistryConfigExample
	RegistryConfig map[string]*RegistryConfig `yaml:"config,omitempty"`
}

// PodCheckpointer represents the pod-checkpointer config values.
type PodCheckpointer struct {
	//   description: |
	//     The `image` field is an override to the default pod-checkpointer image.
	PodCheckpointerImage string `yaml:"image,omitempty"`
}

// CoreDNS represents the CoreDNS config values.
type CoreDNS struct {
	//   description: |
	//     Disable coredns deployment on cluster bootstrap.
	CoreDNSDisabled bool `yaml:"disabled,omitempty"`
	//   description: |
	//     The `image` field is an override to the default coredns image.
	CoreDNSImage string `yaml:"image,omitempty"`
	//   description: |
	//     The `replicas` field is an override to the default coredns number of desired replicas.
	//     The default is `2`.
	CoreDNSReplicas int `yaml:"replicas,omitempty"`
}

// Endpoint represents the endpoint URL parsed out of the machine config.
type Endpoint struct {
	*url.URL
}

// UnmarshalYAML is a custom unmarshaller for `Endpoint`.
func (e *Endpoint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var endpoint string

	if err := unmarshal(&endpoint); err != nil {
		return err
	}

	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	*e = Endpoint{url}

	return nil
}

// MarshalYAML is a custom marshaller for `Endpoint`.
func (e *Endpoint) MarshalYAML() (interface{}, error) {
	return e.URL.String(), nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (e *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *e

	if e.URL != nil {
		in, out := &e.URL, &out.URL
		*out = new(url.URL)
		*out = *in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (e *Endpoint) DeepCopy() *Endpoint {
	if e == nil {
		return nil
	}

	out := new(Endpoint)
	e.DeepCopyInto(out)

	return out
}

// ControlPlaneConfig represents the control plane configuration options.
type ControlPlaneConfig struct {
	//   description: |
	//     Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
	//     It is single-valued, and may optionally include a port number.
	//   examples:
	//     - value: clusterEndpointExample1
	//     - value: clusterEndpointExample2
	Endpoint *Endpoint `yaml:"endpoint"`
	//   description: |
	//     The port that the API server listens on internally.
	//     This may be different than the port portion listed in the endpoint field above.
	//     The default is `6443`.
	LocalAPIServerPort int `yaml:"localAPIServerPort,omitempty"`
}

// APIServerConfig represents the kube apiserver configuration options.
type APIServerConfig struct {
	//   description: |
	//     The container image used in the API server manifest.
	//   examples:
	//     - value: clusterAPIServerImageExample
	ContainerImage string `yaml:"image,omitempty"`
	//   description: |
	//     Extra arguments to supply to the API server.
	ExtraArgsConfig map[string]string `yaml:"extraArgs,omitempty"`
	//   description: |
	//     Extra volumes to mount to the API server static pod.
	ExtraVolumesConfig []VolumeMountConfig `yaml:"extraVolumes,omitempty"`
	//   description: |
	//     Extra certificate subject alternative names for the API server's certificate.
	CertSANs []string `yaml:"certSANs,omitempty"`
}

// ControllerManagerConfig represents the kube controller manager configuration options.
type ControllerManagerConfig struct {
	//   description: |
	//     The container image used in the controller manager manifest.
	//   examples:
	//     - value: clusterControllerManagerImageExample
	ContainerImage string `yaml:"image,omitempty"`
	//   description: |
	//     Extra arguments to supply to the controller manager.
	ExtraArgsConfig map[string]string `yaml:"extraArgs,omitempty"`
	//   description: |
	//     Extra volumes to mount to the controller manager static pod.
	ExtraVolumesConfig []VolumeMountConfig `yaml:"extraVolumes,omitempty"`
}

// ProxyConfig represents the kube proxy configuration options.
type ProxyConfig struct {
	//   description: |
	//     Disable kube-proxy deployment on cluster bootstrap.
	//   examples:
	//     - value: false
	Disabled bool `yaml:"disabled,omitempty"`
	//   description: |
	//     The container image used in the kube-proxy manifest.
	//   examples:
	//     - value: clusterProxyImageExample
	ContainerImage string `yaml:"image,omitempty"`
	//   description: |
	//     proxy mode of kube-proxy.
	//     The default is 'iptables'.
	ModeConfig string `yaml:"mode,omitempty"`
	//   description: |
	//     Extra arguments to supply to kube-proxy.
	ExtraArgsConfig map[string]string `yaml:"extraArgs,omitempty"`
}

// SchedulerConfig represents the kube scheduler configuration options.
type SchedulerConfig struct {
	//   description: |
	//     The container image used in the scheduler manifest.
	//   examples:
	//     - value: clusterSchedulerImageExample
	ContainerImage string `yaml:"image,omitempty"`
	//   description: |
	//     Extra arguments to supply to the scheduler.
	ExtraArgsConfig map[string]string `yaml:"extraArgs,omitempty"`
	//   description: |
	//     Extra volumes to mount to the scheduler static pod.
	ExtraVolumesConfig []VolumeMountConfig `yaml:"extraVolumes,omitempty"`
}

// EtcdConfig represents the etcd configuration options.
type EtcdConfig struct {
	//   description: |
	//     The container image used to create the etcd service.
	//   examples:
	//     - value: clusterEtcdImageExample
	ContainerImage string `yaml:"image,omitempty"`
	//   description: |
	//     The `ca` is the root certificate authority of the PKI.
	//     It is composed of a base64 encoded `crt` and `key`.
	//   examples:
	//     - value: pemEncodedCertificateExample
	RootCA *x509.PEMEncodedCertificateAndKey `yaml:"ca"`
	//   description: |
	//     Extra arguments to supply to etcd.
	//     Note that the following args are not allowed:
	//
	//     - `name`
	//     - `data-dir`
	//     - `initial-cluster-state`
	//     - `listen-peer-urls`
	//     - `listen-client-urls`
	//     - `cert-file`
	//     - `key-file`
	//     - `trusted-ca-file`
	//     - `peer-client-cert-auth`
	//     - `peer-cert-file`
	//     - `peer-trusted-ca-file`
	//     - `peer-key-file`
	//   examples:
	//     - values: >
	//         map[string]string{
	//           "initial-cluster": "https://1.2.3.4:2380",
	//           "advertise-client-urls": "https://1.2.3.4:2379",
	//         }
	EtcdExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}

// ClusterNetworkConfig represents kube networking configuration options.
type ClusterNetworkConfig struct {
	//   description: |
	//     The CNI used.
	//     Composed of "name" and "urls".
	//     The "name" key supports the following options: "flannel", "custom", and "none".
	//     "flannel" uses Talos-managed Flannel CNI, and that's the default option.
	//     "custom" uses custom manifests that should be provided in "urls".
	//     "none" indicates that Talos will not manage any CNI installation.
	//   examples:
	//     - value: clusterCustomCNIExample
	CNI *CNIConfig `yaml:"cni,omitempty"`
	//   description: |
	//     The domain used by Kubernetes DNS.
	//     The default is `cluster.local`
	//   examples:
	//     - value: '"cluser.local"'
	DNSDomain string `yaml:"dnsDomain"`
	//   description: |
	//     The pod subnet CIDR.
	//   examples:
	//     -  value: >
	//          []string{"10.244.0.0/16"}
	PodSubnet []string `yaml:"podSubnets"`
	//   description: |
	//     The service subnet CIDR.
	//   examples:
	//   examples:
	//     -  value: >
	//          []string{"10.96.0.0/12"}
	ServiceSubnet []string `yaml:"serviceSubnets"`
}

// CNIConfig represents the CNI configuration options.
type CNIConfig struct {
	//   description: |
	//     Name of CNI to use.
	//   values:
	//     - flannel
	//     - custom
	//     - none
	CNIName string `yaml:"name,omitempty"`
	//   description: |
	//     URLs containing manifests to apply for the CNI.
	//     Should be present for "custom", must be empty for "flannel" and "none".
	CNIUrls []string `yaml:"urls,omitempty"`
}

// ExternalCloudProviderConfig contains external cloud provider configuration.
type ExternalCloudProviderConfig struct {
	//   description: |
	//     Enable external cloud provider.
	//   values:
	//     - true
	//     - yes
	//     - false
	//     - no
	ExternalEnabled bool `yaml:"enabled,omitempty"`
	//   description: |
	//     A list of urls that point to additional manifests for an external cloud provider.
	//     These will get automatically deployed as part of the bootstrap.
	//   examples:
	//     - value: >
	//        []string{
	//         "https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml",
	//         "https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml",
	//        }
	ExternalManifests []string `yaml:"manifests,omitempty"`
}

// AdminKubeconfigConfig contains admin kubeconfig settings.
type AdminKubeconfigConfig struct {
	//   description: |
	//     Admin kubeconfig certificate lifetime (default is 1 year).
	//     Field format accepts any Go time.Duration format ('1h' for one hour, '10m' for ten minutes).
	AdminKubeconfigCertLifetime time.Duration `yaml:"certLifetime,omitempty"`
}

// MachineDisk represents the options available for partitioning, formatting, and
// mounting extra disks.
type MachineDisk struct {
	//   description: The name of the disk to use.
	DeviceName string `yaml:"device,omitempty"`
	//   description: A list of partitions to create on the disk.
	DiskPartitions []*DiskPartition `yaml:"partitions,omitempty"`
}

// DiskSize partition size in bytes.
type DiskSize uint64

// MarshalYAML write as human readable string.
func (ds DiskSize) MarshalYAML() (interface{}, error) {
	if ds%DiskSize(1000) == 0 {
		bytesString := humanize.Bytes(uint64(ds))
		// ensure that stringifying bytes as human readable string
		// doesn't lose precision
		parsed, err := humanize.ParseBytes(bytesString)
		if err == nil && parsed == uint64(ds) {
			return bytesString, nil
		}
	}

	return uint64(ds), nil
}

// UnmarshalYAML read from human readable string.
func (ds *DiskSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var size string

	if err := unmarshal(&size); err != nil {
		return err
	}

	s, err := humanize.ParseBytes(size)
	if err != nil {
		return err
	}

	*ds = DiskSize(s)

	return nil
}

// DiskPartition represents the options for a disk partition.
type DiskPartition struct {
	//   description: >
	//     The size of partition: either bytes or human readable representation. If `size:`
	//     is omitted, the partition is sized to occupy the full disk.
	//   examples:
	//     - name: Human readable representation.
	//       value: DiskSize(100000000)
	//     - name: Precise value in bytes.
	//       value: 1024 * 1024 * 1024
	DiskSize DiskSize `yaml:"size,omitempty"`
	//   description:
	//     Where to mount the partition.
	DiskMountPoint string `yaml:"mountpoint,omitempty"`
}

// EncryptionConfig represents partition encryption settings.
type EncryptionConfig struct {
	//   description: >
	//     Encryption provider to use for the encryption.
	//   examples:
	//     - value: '"luks2"'
	EncryptionProvider string `yaml:"provider"`
	//   description: >
	//     Defines the encryption keys generation and storage method.
	EncryptionKeys []*EncryptionKey `yaml:"keys"`
	//   description: >
	//     Cipher kind to use for the encryption.
	//     Depends on the encryption provider.
	EncryptionCipher string `yaml:"cipher,omitempty"`
}

// EncryptionKey represents configuration for disk encryption key.
type EncryptionKey struct {
	//   description: >
	//     Key which value is stored in the configuration file.
	KeyStatic *EncryptionKeyStatic `yaml:"static,omitempty"`
	//   description: >
	//     Deterministically generated key from the node UUID and PartitionLabel.
	KeyNodeID *EncryptionKeyNodeID `yaml:"nodeID,omitempty"`
	//   description: >
	//     Key slot number for luks2 encryption.
	KeySlot int `yaml:"slot"`
}

// EncryptionKeyStatic represents throw away key type.
type EncryptionKeyStatic struct {
	//   description: >
	//     Defines the static passphrase value.
	KeyData string `yaml:"passphrase,omitempty"`
}

// EncryptionKeyNodeID represents deterministically generated key from the node UUID and PartitionLabel.
type EncryptionKeyNodeID struct{}

// Env represents a set of environment variables.
type Env = map[string]string

// FileMode represents file's permissions.
type FileMode os.FileMode

// String convert file mode to octal string.
func (fm FileMode) String() string {
	return "0o" + strconv.FormatUint(uint64(fm), 8)
}

// MarshalYAML encodes as an octal value.
func (fm FileMode) MarshalYAML() (interface{}, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!int",
		Value: fm.String(),
	}, nil
}

// MachineFile represents a file to write to disk.
type MachineFile struct {
	//   description: The contents of the file.
	FileContent string `yaml:"content"`
	//   description: The file's permissions in octal.
	FilePermissions FileMode `yaml:"permissions"`
	//   description: The path of the file.
	FilePath string `yaml:"path"`
	//   description: The operation to use
	//   values:
	//     - create
	//     - append
	//     - overwrite
	FileOp string `yaml:"op"`
}

// ExtraHost represents a host entry in /etc/hosts.
type ExtraHost struct {
	//   description: The IP of the host.
	HostIP string `yaml:"ip"`
	//   description: The host alias.
	HostAliases []string `yaml:"aliases"`
}

// Device represents a network interface.
type Device struct {
	//   description: The interface name.
	//   examples:
	//     - value: '"eth0"'
	DeviceInterface string `yaml:"interface"`
	//   description: |
	//     Assigns a static IP address to the interface.
	//     This should be in proper CIDR notation.
	//
	//     > Note: This option is mutually exclusive with DHCP option.
	//   examples:
	//     - value: '"10.5.0.0/16"'
	DeviceCIDR string `yaml:"cidr,omitempty"`
	//   description: |
	//     A list of routes associated with the interface.
	//     If used in combination with DHCP, these routes will be appended to routes returned by DHCP server.
	//   examples:
	//     - value: networkConfigRoutesExample
	DeviceRoutes []*Route `yaml:"routes,omitempty"`
	//   description: Bond specific options.
	//   examples:
	//     - value: networkConfigBondExample
	DeviceBond *Bond `yaml:"bond,omitempty"`
	//   description: VLAN specific options.
	DeviceVlans []*Vlan `yaml:"vlans,omitempty"`
	//   description: |
	//     The interface's MTU.
	//     If used in combination with DHCP, this will override any MTU settings returned from DHCP server.
	DeviceMTU int `yaml:"mtu"`
	//   description: |
	//     Indicates if DHCP should be used to configure the interface.
	//     The following DHCP options are supported:
	//
	//     - `OptionClasslessStaticRoute`
	//     - `OptionDomainNameServer`
	//     - `OptionDNSDomainSearchList`
	//     - `OptionHostName`
	//
	//     > Note: This option is mutually exclusive with CIDR.
	//     >
	//     > Note: To configure an interface with *only* IPv6 SLAAC addressing, CIDR should be set to "" and DHCP to false
	//     > in order for Talos to skip configuration of addresses.
	//     > All other options will still apply.
	//   examples:
	//     - value: true
	DeviceDHCP bool `yaml:"dhcp,omitempty"`
	//   description: Indicates if the interface should be ignored (skips configuration).
	DeviceIgnore bool `yaml:"ignore,omitempty"`
	//   description: |
	//     Indicates if the interface is a dummy interface.
	//     `dummy` is used to specify that this interface should be a virtual-only, dummy interface.
	DeviceDummy bool `yaml:"dummy,omitempty"`
	//   description: |
	//     DHCP specific options.
	//     `dhcp` *must* be set to true for these to take effect.
	//   examples:
	//     - value: networkConfigDHCPOptionsExample
	DeviceDHCPOptions *DHCPOptions `yaml:"dhcpOptions,omitempty"`
	//   description: |
	//     Wireguard specific configuration.
	//     Includes things like private key, listen port, peers.
	//   examples:
	//     - name: wireguard server example
	//       value: networkConfigWireguardHostExample
	//     - name: wireguard peer example
	//       value: networkConfigWireguardPeerExample
	DeviceWireguardConfig *DeviceWireguardConfig `yaml:"wireguard,omitempty"`
	//   description: Virtual (shared) IP address configuration.
	//   examples:
	//     - name: layer2 vip example
	//     - value: networkConfigVIPLayer2Example
	DeviceVIPConfig *DeviceVIPConfig `yaml:"vip,omitempty"`
}

// DHCPOptions contains options for configuring the DHCP settings for a given interface.
type DHCPOptions struct {
	//   description: The priority of all routes received via DHCP.
	DHCPRouteMetric uint32 `yaml:"routeMetric"`
	//   description: Enables DHCPv4 protocol for the interface (default is enabled).
	DHCPIPv4 *bool `yaml:"ipv4,omitempty"`
	//   description: Enables DHCPv6 protocol for the interface (default is disabled).
	DHCPIPv6 *bool `yaml:"ipv6,omitempty"`
}

// DeviceWireguardConfig contains settings for configuring Wireguard network interface.
type DeviceWireguardConfig struct {
	//   description: |
	//     Specifies a private key configuration (base64 encoded).
	//     Can be generated by `wg genkey`.
	WireguardPrivateKey string `yaml:"privateKey,omitempty"`
	//   description: Specifies a device's listening port.
	WireguardListenPort int `yaml:"listenPort,omitempty"`
	//   description: Specifies a device's firewall mark.
	WireguardFirewallMark int `yaml:"firewallMark,omitempty"`
	//   description: Specifies a list of peer configurations to apply to a device.
	WireguardPeers []*DeviceWireguardPeer `yaml:"peers,omitempty"`
}

// DeviceWireguardPeer a WireGuard device peer configuration.
type DeviceWireguardPeer struct {
	//   description: |
	//     Specifies the public key of this peer.
	//     Can be extracted from private key by running `wg pubkey < private.key > public.key && cat public.key`.
	WireguardPublicKey string `yaml:"publicKey,omitempty"`
	//   description: Specifies the endpoint of this peer entry.
	WireguardEndpoint string `yaml:"endpoint,omitempty"`
	//   description: |
	//     Specifies the persistent keepalive interval for this peer.
	//     Field format accepts any Go time.Duration format ('1h' for one hour, '10m' for ten minutes).
	WireguardPersistentKeepaliveInterval time.Duration `yaml:"persistentKeepaliveInterval,omitempty"`
	//   description: AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
	WireguardAllowedIPs []string `yaml:"allowedIPs,omitempty"`
}

// DeviceVIPConfig contains settings for configuring a Virtual Shared IP on an interface.
type DeviceVIPConfig struct {
	// description: Specifies the IP address to be used.
	SharedIP string `yaml:"ip,omitempty"`
}

// Bond contains the various options for configuring a bonded interface.
type Bond struct {
	//   description: The interfaces that make up the bond.
	BondInterfaces []string `yaml:"interfaces"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	//     Not supported at the moment.
	BondARPIPTarget []string `yaml:"arpIPTarget,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondMode string `yaml:"mode"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondHashPolicy string `yaml:"xmitHashPolicy,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondLACPRate string `yaml:"lacpRate,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	//     Not supported at the moment.
	BondADActorSystem string `yaml:"adActorSystem,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondARPValidate string `yaml:"arpValidate,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondARPAllTargets string `yaml:"arpAllTargets,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondPrimary string `yaml:"primary,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondPrimaryReselect string `yaml:"primaryReselect,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondFailOverMac string `yaml:"failOverMac,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondADSelect string `yaml:"adSelect,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondMIIMon uint32 `yaml:"miimon,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondUpDelay uint32 `yaml:"updelay,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondDownDelay uint32 `yaml:"downdelay,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondARPInterval uint32 `yaml:"arpInterval,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondResendIGMP uint32 `yaml:"resendIgmp,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondMinLinks uint32 `yaml:"minLinks,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondLPInterval uint32 `yaml:"lpInterval,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondPacketsPerSlave uint32 `yaml:"packetsPerSlave,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondNumPeerNotif uint8 `yaml:"numPeerNotif,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondTLBDynamicLB uint8 `yaml:"tlbDynamicLb,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondAllSlavesActive uint8 `yaml:"allSlavesActive,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondUseCarrier *bool `yaml:"useCarrier,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondADActorSysPrio uint16 `yaml:"adActorSysPrio,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondADUserPortKey uint16 `yaml:"adUserPortKey,omitempty"`
	//   description: |
	//     A bond option.
	//     Please see the official kernel documentation.
	BondPeerNotifyDelay uint32 `yaml:"peerNotifyDelay,omitempty"`
}

// Vlan represents vlan settings for a device.
type Vlan struct {
	//   description: The CIDR to use.
	VlanCIDR string `yaml:"cidr"`
	//   description: A list of routes associated with the VLAN.
	VlanRoutes []*Route `yaml:"routes"`
	//   description: Indicates if DHCP should be used.
	VlanDHCP bool `yaml:"dhcp"`
	//   description: The VLAN's ID.
	VlanID uint16 `yaml:"vlanId"`
}

// Route represents a network route.
type Route struct {
	//   description: The route's network.
	RouteNetwork string `yaml:"network"`
	//   description: The route's gateway.
	RouteGateway string `yaml:"gateway"`
	//   description: The optional metric for the route.
	RouteMetric uint32 `yaml:"metric,omitempty"`
}

// RegistryMirrorConfig represents mirror configuration for a registry.
type RegistryMirrorConfig struct {
	//   description: |
	//     List of endpoints (URLs) for registry mirrors to use.
	//     Endpoint configures HTTP/HTTPS access mode, host name,
	//     port and path (if path is not set, it defaults to `/v2`).
	MirrorEndpoints []string `yaml:"endpoints"`
}

// RegistryConfig specifies auth & TLS config per registry.
type RegistryConfig struct {
	//   description: |
	//     The TLS configuration for the registry.
	//   examples:
	//     - value: machineConfigRegistryTLSConfigExample1
	//     - value: machineConfigRegistryTLSConfigExample2
	RegistryTLS *RegistryTLSConfig `yaml:"tls,omitempty"`
	//   description: The auth configuration for this registry.
	//   examples:
	//     - value: machineConfigRegistryAuthConfigExample
	RegistryAuth *RegistryAuthConfig `yaml:"auth,omitempty"`
}

// RegistryAuthConfig specifies authentication configuration for a registry.
type RegistryAuthConfig struct {
	//   description: |
	//     Optional registry authentication.
	//     The meaning of each field is the same with the corresponding field in .docker/config.json.
	RegistryUsername string `yaml:"username,omitempty"`
	//   description: |
	//     Optional registry authentication.
	//     The meaning of each field is the same with the corresponding field in .docker/config.json.
	RegistryPassword string `yaml:"password,omitempty"`
	//   description: |
	//     Optional registry authentication.
	//     The meaning of each field is the same with the corresponding field in .docker/config.json.
	RegistryAuth string `yaml:"auth,omitempty"`
	//   description: |
	//     Optional registry authentication.
	//     The meaning of each field is the same with the corresponding field in .docker/config.json.
	RegistryIdentityToken string `yaml:"identityToken,omitempty"`
}

// RegistryTLSConfig specifies TLS config for HTTPS registries.
type RegistryTLSConfig struct {
	//   description: |
	//     Enable mutual TLS authentication with the registry.
	//     Client certificate and key should be base64-encoded.
	//   examples:
	//     - value: pemEncodedCertificateExample
	TLSClientIdentity *x509.PEMEncodedCertificateAndKey `yaml:"clientIdentity,omitempty"`
	//   description: |
	//     CA registry certificate to add the list of trusted certificates.
	//     Certificate should be base64-encoded.
	TLSCA Base64Bytes `yaml:"ca,omitempty"`
	//   description: |
	//     Skip TLS server certificate verification (not recommended).
	TLSInsecureSkipVerify bool `yaml:"insecureSkipVerify,omitempty"`
}

// SystemDiskEncryptionConfig specifies system disk partitions encryption settings.
type SystemDiskEncryptionConfig struct {
	//   description: |
	//     State partition encryption.
	StatePartition *EncryptionConfig `yaml:"state,omitempty"`
	//   description: |
	//     Ephemeral partition encryption.
	EphemeralPartition *EncryptionConfig `yaml:"ephemeral,omitempty"`
}

// FeaturesConfig describe individual Talos features that can be switched on or off.
type FeaturesConfig struct {
	//   description: |
	//     Enable role-based access control (RBAC).
	RBAC *bool `yaml:"rbac,omitempty"`
}

// VolumeMountConfig struct describes extra volume mount for the static pods.
type VolumeMountConfig struct {
	//   description: |
	//     Path on the host.
	//   examples:
	//     - value: '"/var/lib/auth"'
	VolumeHostPath string `yaml:"hostPath"`
	//   description: |
	//     Path in the container.
	//   examples:
	//     - value: '"/etc/kubernetes/auth"'
	VolumeMountPath string `yaml:"mountPath"`
	//   description: |
	//     Mount the volume read only.
	//   examples:
	//     - value: true
	VolumeReadOnly bool `yaml:"readonly,omitempty"`
}

// ClusterInlineManifests is a list of ClusterInlineManifest.
type ClusterInlineManifests []ClusterInlineManifest

// ClusterInlineManifest struct describes inline bootstrap manifests for the user.
type ClusterInlineManifest struct {
	//   description: |
	//     Name of the manifest.
	//     Name should be unique.
	//   examples:
	//     - value: '"csi"'
	InlineManifestName string `yaml:"name"`
	//   description: |
	//     Manifest contents as a string.
	//   examples:
	//     - value: '"/etc/kubernetes/auth"'
	InlineManifestContents string `yaml:"contents"`
}
