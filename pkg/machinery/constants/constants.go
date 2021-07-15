// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package constants

import (
	"time"

	cni "github.com/containerd/go-cni"
	"github.com/talos-systems/crypto/x509"
)

const (
	// DefaultKernelVersion is the default Linux kernel version.
	DefaultKernelVersion = "5.10.45-talos"

	// KernelParamConfig is the kernel parameter name for specifying the URL.
	// to the config.
	KernelParamConfig = "talos.config"

	// ConfigNone indicates no config is required.
	ConfigNone = "none"

	// KernelParamPlatform is the kernel parameter name for specifying the
	// platform.
	KernelParamPlatform = "talos.platform"

	// KernelParamBoard is the kernel parameter name for specifying the
	// SBC.
	KernelParamBoard = "talos.board"

	// BoardNone indicates that the install is not for a specific board.
	BoardNone = "none"

	// BoardLibretechAllH3CCH5 is the  name of the Libre Computer board ALL-H3-CC.
	BoardLibretechAllH3CCH5 = "libretech_all_h3_cc_h5"

	// BoardRPi4 is the  name of the Raspberry Pi 4 Model B.
	BoardRPi4 = "rpi_4"

	// BoardBananaPiM64 is the  name of the Banana Pi M64.
	BoardBananaPiM64 = "bananapi_m64"

	// BoardPine64 is the  name of the Pine64.
	BoardPine64 = "pine64"

	// BoardRock64 is the  name of the Rock64.
	BoardRock64 = "rock64"

	// BoardRockpi4 is the name of the Radxa Rock pi 4.
	BoardRockpi4 = "rockpi_4"

	// KernelParamHostname is the kernel parameter name for specifying the
	// hostname.
	KernelParamHostname = "talos.hostname"

	// KernelParamShutdown is the kernel parameter for specifying the
	// shutdown type (halt/poweroff).
	KernelParamShutdown = "talos.shutdown"

	// KernelParamNetworkInterfaceIgnore is the kernel parameter for specifying network interfaces which should be ignored by talos.
	KernelParamNetworkInterfaceIgnore = "talos.network.interface.ignore"

	// KernelParamPanic is the kernel parameter name for specifying the time to wait until rebooting after kernel panic (0 disables reboot).
	KernelParamPanic = "panic"

	// KernelCurrentRoot is the kernel parameter name for specifying the
	// current root partition.
	//
	// Deprecated: Talos now expects to use an entire disk; this constant will be removed in 0.12
	// (https://github.com/talos-systems/talos/issues/3909).
	KernelCurrentRoot = "talos.root"

	// NewRoot is the path where the switchroot target is mounted.
	NewRoot = "/root"

	// EFIPartitionLabel is the label of the partition to use for mounting at
	// the boot path.
	EFIPartitionLabel = "EFI"

	// EFIMountPoint is the label of the partition to use for mounting at
	// the boot path.
	EFIMountPoint = BootMountPoint + "/EFI"

	// BIOSGrubPartitionLabel is the label of the partition used by grub's second
	// stage bootloader.
	BIOSGrubPartitionLabel = "BIOS"

	// MetaPartitionLabel is the label of the meta partition.
	MetaPartitionLabel = "META"

	// StatePartitionLabel is the label of the state partition.
	StatePartitionLabel = "STATE"

	// StateMountPoint is the label of the partition to use for mounting at
	// the state path.
	StateMountPoint = "/system/state"

	// BootPartitionLabel is the label of the partition to use for mounting at
	// the boot path.
	BootPartitionLabel = "BOOT"

	// BootMountPoint is the label of the partition to use for mounting at
	// the boot path.
	BootMountPoint = "/boot"

	// EphemeralPartitionLabel is the label of the partition to use for
	// mounting at the data path.
	EphemeralPartitionLabel = "EPHEMERAL"

	// EphemeralMountPoint is the label of the partition to use for mounting at
	// the data path.
	EphemeralMountPoint = "/var"

	// RootMountPoint is the label of the partition to use for mounting at
	// the root path.
	RootMountPoint = "/"

	// ISOFilesystemLabel is the label of the ISO file system for the Talos
	// installer.
	ISOFilesystemLabel = "TALOS"

	// PATH defines all locations where executables are stored.
	PATH = "/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:" + cni.DefaultCNIDir

	// KubernetesDefaultCertificateValidityDuration specifies default certificate duration for Kubernetes generated certificates.
	KubernetesDefaultCertificateValidityDuration = time.Hour * 24 * 365

	// DefaultCertificatesDir is the path the the Kubernetes PKI directory.
	DefaultCertificatesDir = "/etc/kubernetes/pki"

	// KubernetesCACert is the path to the root CA certificate.
	KubernetesCACert = DefaultCertificatesDir + "/" + "ca.crt"

	// KubernetesCAKey is the path to the root CA private key.
	KubernetesCAKey = DefaultCertificatesDir + "/" + "ca.key"

	// KubernetesEtcdCACert is the path to the etcd CA certificate.
	KubernetesEtcdCACert = EtcdPKIPath + "/" + "ca.crt"

	// KubernetesEtcdCAKey is the path to the etcd CA private key.
	KubernetesEtcdCAKey = EtcdPKIPath + "/" + "ca.key"

	// KubernetesEtcdCert is the path to the etcd server certificate.
	KubernetesEtcdCert = EtcdPKIPath + "/" + "server.crt"

	// KubernetesEtcdKey is the path to the etcd server private key.
	KubernetesEtcdKey = EtcdPKIPath + "/" + "server.key"

	// KubernetesEtcdPeerCert is the path to the etcd peer certificate.
	KubernetesEtcdPeerCert = EtcdPKIPath + "/" + "peer.crt"

	// KubernetesEtcdPeerKey is the path to the etcd peer private key.
	KubernetesEtcdPeerKey = EtcdPKIPath + "/" + "peer.key"

	// KubernetesEtcdAdminCert is the path to the talos client certificate.
	KubernetesEtcdAdminCert = EtcdPKIPath + "/" + "admin.crt"

	// KubernetesEtcdAdminKey is the path to the talos client private key.
	KubernetesEtcdAdminKey = EtcdPKIPath + "/" + "admin.key"

	// KubernetesEtcdListenClientPort defines the port etcd listen on for client traffic.
	KubernetesEtcdListenClientPort = "2379"

	// KubernetesAdminCertCommonName defines CN property of Kubernetes admin certificate.
	KubernetesAdminCertCommonName = "admin"

	// KubernetesAdminCertOrganization defines Organization values of Kubernetes admin certificate.
	KubernetesAdminCertOrganization = "system:masters"

	// KubernetesAPIServerKubeletClientCommonName defines CN property of Kubernetes API server certificate to access kubelet API.
	KubernetesAPIServerKubeletClientCommonName = "apiserver-kubelet-client"

	// KubernetesControllerManagerOrganization defines Organization value of kube-controller-manager client certificate.
	KubernetesControllerManagerOrganization = "system:kube-controller-manager"

	// KubernetesSchedulerOrganization defines Organization value of kube-scheduler client certificate.
	KubernetesSchedulerOrganization = "system:kube-scheduler"

	// KubernetesAdminCertDefaultLifetime defines default lifetime for Kubernetes generated admin certificate.
	KubernetesAdminCertDefaultLifetime = 365 * 24 * time.Hour

	// KubebernetesStaticSecretsDir defines ephemeral directory which contains rendered secrets for controlplane components.
	KubebernetesStaticSecretsDir = "/system/secrets/kubernetes"

	// KubernetesAPIServerSecretsDir defines ephemeral directory with kube-apiserver secrets.
	KubernetesAPIServerSecretsDir = KubebernetesStaticSecretsDir + "/" + "kube-apiserver"

	// KubernetesControllerManagerSecretsDir defines ephemeral directory with kube-controller-manager secrets.
	KubernetesControllerManagerSecretsDir = KubebernetesStaticSecretsDir + "/" + "kube-controller-manager"

	// KubernetesSchedulerSecretsDir defines ephemeral directory with kube-scheduler secrets.
	KubernetesSchedulerSecretsDir = KubebernetesStaticSecretsDir + "/" + "kube-scheduler"

	// KubernetesRunUser defines UID to run control plane components.
	KubernetesRunUser = 65534

	// KubeletBootstrapKubeconfig is the path to the kubeconfig required to
	// bootstrap the kubelet.
	KubeletBootstrapKubeconfig = "/etc/kubernetes/bootstrap-kubeconfig"

	// KubeletPort is the kubelet port for secure API.
	KubeletPort = 10250

	// KubeletPKIDir is the path to the directory where kubelet stores issued certificates and keys.
	KubeletPKIDir = "/var/lib/kubelet/pki"

	// SystemKubeletPKIDir is the path to the directory where Talos copies kubelet issued certificates and keys.
	SystemKubeletPKIDir = "/system/secrets/kubelet"

	// DefaultKubernetesVersion is the default target version of the control plane.
	DefaultKubernetesVersion = "1.22.0-beta.2"

	// DefaultControlPlanePort is the default port to use for the control plane.
	DefaultControlPlanePort = 6443

	// KubeletImage is the enforced kubelet image to use.
	KubeletImage = "ghcr.io/talos-systems/kubelet"

	// KubeProxyImage is the enforced kube-proxy image to use for the control plane.
	KubeProxyImage = "k8s.gcr.io/kube-proxy"

	// KubernetesAPIServerImage is the enforced apiserver image to use for the control plane.
	KubernetesAPIServerImage = "k8s.gcr.io/kube-apiserver"

	// KubernetesControllerManagerImage is the enforced controllermanager image to use for the control plane.
	KubernetesControllerManagerImage = "k8s.gcr.io/kube-controller-manager"

	// KubernetesProxyImage is the enforced proxy image to use for the control plane.
	KubernetesProxyImage = "k8s.gcr.io/kube-proxy"

	// KubernetesSchedulerImage is the enforced scheduler image to use for the control plane.
	KubernetesSchedulerImage = "k8s.gcr.io/kube-scheduler"

	// CoreDNSImage is the enforced CoreDNS image to use.
	CoreDNSImage = "docker.io/coredns/coredns"

	// DefaultCoreDNSVersion is the default version for the CoreDNS.
	DefaultCoreDNSVersion = "1.8.4"

	// LabelNodeRoleMaster is the node label required by a control plane node.
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	// LabelNodeRoleControlPlane is the node label required by a control plane node.
	LabelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"

	// ManifestsDirectory is the directory that contains all static manifests.
	ManifestsDirectory = "/etc/kubernetes/manifests"

	// TalosManifestPrefix is the prefix for static pod files created in ManifestsDirectory by Talos.
	TalosManifestPrefix = "talos-"

	// KubeletKubeconfig is the generated kubeconfig for kubelet.
	KubeletKubeconfig = "/etc/kubernetes/kubeconfig-kubelet"

	// DefaultEtcdVersion is the default target version of etcd.
	DefaultEtcdVersion = "v3.4.16"

	// EtcdRootTalosKey is the root etcd key for Talos-specific storage.
	EtcdRootTalosKey = "talos:v1"

	// EtcdTalosEtcdUpgradeMutex is the etcd mutex prefix to be used to set an etcd upgrade lock.
	EtcdTalosEtcdUpgradeMutex = EtcdRootTalosKey + ":etcdUpgradeMutex"

	// EtcdTalosManifestApplyMutex is the etcd election .
	EtcdTalosManifestApplyMutex = EtcdRootTalosKey + ":manifestApplyMutex"

	// EtcdImage is the reposistory for the etcd image.
	EtcdImage = "gcr.io/etcd-development/etcd"

	// EtcdPKIPath is the path to the etcd PKI directory.
	EtcdPKIPath = DefaultCertificatesDir + "/etcd"

	// EtcdDataPath is the path where etcd stores its' data.
	EtcdDataPath = "/var/lib/etcd"

	// EtcdRecoverySnapshotPath is the path where etcd snapshot is uploaded for recovery.
	EtcdRecoverySnapshotPath = "/var/lib/etcd.snapshot"

	// ConfigPath is the path to the downloaded config.
	ConfigPath = StateMountPoint + "/config.yaml"

	// MetalConfigISOLabel is the volume label for ISO based configuration.
	MetalConfigISOLabel = "metal-iso"

	// ConfigGuestInfo is the name of the VMware guestinfo config strategy.
	ConfigGuestInfo = "guestinfo"

	// VMwareGuestInfoConfigKey is the guestinfo key used to provide a config file.
	VMwareGuestInfoConfigKey = "talos.config"

	// VMwareGuestInfoFallbackKey is the fallback guestinfo key used to provide a config file.
	VMwareGuestInfoFallbackKey = "userdata"

	// VMwareGuestInfoOvfEnvKey is the guestinfo key used to provide the OVF environment.
	VMwareGuestInfoOvfEnvKey = "ovfenv"

	// AuditPolicyPath is the path to the audit-policy.yaml relative to initramfs.
	AuditPolicyPath = "/etc/kubernetes/audit-policy.yaml"

	// EncryptionConfigPath is the path to the EncryptionConfig relative to initramfs.
	EncryptionConfigPath = "/etc/kubernetes/encryptionconfig.yaml"

	// EncryptionConfigRootfsPath is the path to the EncryptionConfig relative to rootfs.
	EncryptionConfigRootfsPath = "/etc/kubernetes/encryptionconfig.yaml"

	// ApidPort is the port for the apid service.
	ApidPort = 50000

	// TrustdPort is the port for the trustd service.
	TrustdPort = 50001

	// DefaultContainerdVersion is the default container runtime version.
	DefaultContainerdVersion = "1.5.2"

	// SystemContainerdNamespace is the Containerd namespace for Talos services.
	SystemContainerdNamespace = "system"

	// SystemContainerdAddress is the path to the system containerd socket.
	SystemContainerdAddress = SystemRunPath + "/containerd/containerd.sock"

	// CRIContainerdAddress is the path to the CRI containerd socket address.
	CRIContainerdAddress = "/run/containerd/containerd.sock"

	// CRIContainerdConfig is the path to the config for the containerd instance that provides the CRI.
	CRIContainerdConfig = "/etc/cri/containerd.toml"

	// TalosConfigEnvVar is the environment variable for setting the Talos configuration file path.
	TalosConfigEnvVar = "TALOSCONFIG"

	// APISocketPath is the path to file socket of apid.
	APISocketPath = SystemRunPath + "/apid/apid.sock"

	// APIRuntimeSocketPath is the path to file socket of runtime server for apid.
	APIRuntimeSocketPath = SystemRunPath + "/apid/runtime.sock"

	// MachineSocketPath is the path to file socket of machine API.
	MachineSocketPath = SystemRunPath + "/machined/machine.sock"

	// NetworkSocketPath is the path to file socket of network API.
	NetworkSocketPath = SystemRunPath + "/networkd/networkd.sock"

	// ArchVariable is replaced automatically by the target cluster arch.
	ArchVariable = "${ARCH}"

	// KernelAsset defines a well known name for our kernel filename.
	KernelAsset = "vmlinuz"

	// KernelAssetWithArch defines a well known name for our kernel filename with arch variable.
	KernelAssetWithArch = "vmlinuz-" + ArchVariable

	// KernelAssetPath is the path to the kernel on disk.
	KernelAssetPath = "/usr/install/%s/" + KernelAsset

	// InitramfsAsset defines a well known name for our initramfs filename.
	InitramfsAsset = "initramfs.xz"

	// InitramfsAssetWithArch defines a well known name for our initramfs filename with arch variable.
	InitramfsAssetWithArch = "initramfs-" + ArchVariable + ".xz"

	// InitramfsAssetPath is the path to the initramfs on disk.
	InitramfsAssetPath = "/usr/install/%s/" + InitramfsAsset

	// RootfsAsset defines a well known name for our rootfs filename.
	RootfsAsset = "rootfs.sqsh"

	// DefaultCertificateValidityDuration is the default duration for a certificate.
	DefaultCertificateValidityDuration = x509.DefaultCertificateValidityDuration

	// SystemPath is the path to write temporary runtime system related files
	// and directories.
	SystemPath = "/system"

	// SystemOverlaysPath is the path where overlay mounts are created.
	SystemOverlaysPath = "/var/system/overlays"

	// SystemRunPath is the path to the system run directory.
	SystemRunPath = SystemPath + "/run"

	// SystemVarPath is the path to the system var directory.
	SystemVarPath = SystemPath + "/var"

	// SystemEtcPath is the path to the system etc directory.
	SystemEtcPath = SystemPath + "/etc"

	// SystemLibexecPath is the path to the system libexec directory.
	SystemLibexecPath = SystemPath + "/libexec"

	// FlannelCNI is the string to use Tanos-managed Flannel CNI (default).
	FlannelCNI = "flannel"

	// CustomCNI is the string to use custom CNI managed by Tanos with extra manifests.
	CustomCNI = "custom"

	// NoneCNI is the string to indicate that CNI will not be managed by Talos.
	NoneCNI = "none"

	// DefaultIPv4PodNet is the IPv4 network to be used for kubernetes Pods.
	DefaultIPv4PodNet = "10.244.0.0/16"

	// DefaultIPv4ServiceNet is the IPv4 network to be used for kubernetes Services.
	DefaultIPv4ServiceNet = "10.96.0.0/12"

	// DefaultIPv6PodNet is the IPv6 network to be used for kubernetes Pods.
	DefaultIPv6PodNet = "fc00:db8:10::/56"

	// DefaultIPv6ServiceNet is the IPv6 network to be used for kubernetes Services.
	DefaultIPv6ServiceNet = "fc00:db8:20::/112"

	// DefaultDNSDomain is the default DNS domain.
	DefaultDNSDomain = "cluster.local"

	// InitializedKey is the key used to indicate if the cluster has been
	// initialized.
	InitializedKey = "initialized"

	// BootTimeout is the timeout to run all services.
	BootTimeout = 15 * time.Minute

	// NodeReadyTimeout is the timeout to wait for the node to be ready (CNI to be running).
	// For bootstrap API, this includes time to run bootstrap.
	NodeReadyTimeout = BootTimeout

	// AnnotationCordonedKey is the annotation key for the nodes cordoned by Talos.
	AnnotationCordonedKey = "talos.dev/cordoned"

	// AnnotationCordonedValue is the annotation key for the nodes cordoned by Talos.
	AnnotationCordonedValue = "true"

	// AnnotationStaticPodSecretsVersion is the annotation key for the static pod secret version.
	AnnotationStaticPodSecretsVersion = "talos.dev/secrets-version"

	// AnnotationStaticPodConfigVersion is the annotation key for the static pod config version.
	AnnotationStaticPodConfigVersion = "talos.dev/config-version"

	// DefaultNTPServer is the NTP server to use if not configured explicitly.
	//
	// TODO: Once we get naming sorted we need to apply for a project specific address
	// https://manage.ntppool.org/manage/vendor
	DefaultNTPServer = "pool.ntp.org"

	// DefaultPrimaryResolver is the default primary DNS server.
	DefaultPrimaryResolver = "1.1.1.1"

	// DefaultSecondaryResolver is the default secondary DNS server.
	DefaultSecondaryResolver = "8.8.8.8"
)

// See https://linux.die.net/man/3/klogctl
//nolint:stylecheck,revive
const (
	// SYSLOG_ACTION_SIZE_BUFFER is a named type argument to klogctl.
	//nolint:golint
	SYSLOG_ACTION_SIZE_BUFFER = 10

	// SYSLOG_ACTION_READ_ALL is a named type argument to klogctl.
	//nolint:golint
	SYSLOG_ACTION_READ_ALL = 3
)
