// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cluster

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	stdruntime "runtime"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cobra"
	"github.com/talos-systems/go-blockdevice/blockdevice/encryption"
	talosnet "github.com/talos-systems/net"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/talos-systems/talos/cmd/talosctl/pkg/mgmt/helpers"
	"github.com/talos-systems/talos/internal/pkg/kubeconfig"
	"github.com/talos-systems/talos/pkg/cli"
	"github.com/talos-systems/talos/pkg/cluster/check"
	"github.com/talos-systems/talos/pkg/images"
	clientconfig "github.com/talos-systems/talos/pkg/machinery/client/config"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/bundle"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/talos-systems/talos/pkg/provision"
	"github.com/talos-systems/talos/pkg/provision/access"
	"github.com/talos-systems/talos/pkg/provision/providers"
	"github.com/talos-systems/talos/pkg/version"
)

const (
	// gatewayOffset is the offset from the network address of the IP address of the network gateway.
	gatewayOffset = 1

	// nodesOffset is the offset from the network address of the beginning of the IP addresses to be used for nodes.
	nodesOffset = 2

	// vipOffset is the offset from the network address of the CIDR to use for allocating the Virtual (shared) IP address, if enabled.
	vipOffset = 50
)

var (
	talosconfig               string
	nodeImage                 string
	nodeInstallImage          string
	registryMirrors           []string
	registryInsecure          []string
	kubernetesVersion         string
	nodeVmlinuzPath           string
	nodeInitramfsPath         string
	nodeISOPath               string
	nodeDiskImagePath         string
	applyConfigEnabled        bool
	bootloaderEnabled         bool
	uefiEnabled               bool
	configDebug               bool
	networkCIDR               string
	networkMTU                int
	networkIPv4               bool
	networkIPv6               bool
	wireguardCIDR             string
	nameservers               []string
	dnsDomain                 string
	workers                   int
	masters                   int
	clusterCpus               string
	clusterMemory             int
	clusterDiskSize           int
	clusterDisks              []string
	targetArch                string
	clusterWait               bool
	clusterWaitTimeout        time.Duration
	forceInitNodeAsEndpoint   bool
	forceEndpoint             string
	inputDir                  string
	cniBinPath                []string
	cniConfDir                string
	cniCacheDir               string
	cniBundleURL              string
	ports                     string
	dockerHostIP              string
	withInitNode              bool
	customCNIUrl              string
	crashdumpOnFailure        bool
	skipKubeconfig            bool
	skipInjectingConfig       bool
	talosVersion              string
	encryptStatePartition     bool
	encryptEphemeralPartition bool
	useVIP                    bool
	configPatch               string
	configPatchControlPlane   string
	configPatchWorker         string
	configPatchJoin           string
	badRTC                    bool
)

// createCmd represents the cluster up command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a local docker-based or QEMU-based kubernetes cluster",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.WithContext(context.Background(), create)
	},
}

//nolint:gocyclo,cyclop
func create(ctx context.Context) (err error) {
	if configPatchJoin != "" {
		if configPatchWorker != "" {
			return fmt.Errorf("both --config-patch-join and --config-patch-worker are passed")
		}

		configPatchWorker = configPatchJoin
	}

	if masters < 1 {
		return fmt.Errorf("number of masters can't be less than 1")
	}

	nanoCPUs, err := parseCPUShare()
	if err != nil {
		return fmt.Errorf("error parsing --cpus: %s", err)
	}

	memory := int64(clusterMemory) * 1024 * 1024

	// Validate CIDR range and allocate IPs
	fmt.Println("validating CIDR and reserving IPs")

	_, cidr4, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return fmt.Errorf("error validating cidr block: %w", err)
	}

	if cidr4.IP.To4() == nil {
		return fmt.Errorf("--cidr is expected to be IPV4 CIDR")
	}

	// use ULA IPv6 network fd00::/8, add 'TAL' in hex to build /32 network, add IPv4 CIDR to build /64 unique network
	_, cidr6, err := net.ParseCIDR(fmt.Sprintf("fd74:616c:%02x%02x:%02x%02x::/64", cidr4.IP[0], cidr4.IP[1], cidr4.IP[2], cidr4.IP[3]))
	if err != nil {
		return fmt.Errorf("error validating cidr IPv6 block: %w", err)
	}

	var cidrs []net.IPNet

	if networkIPv4 {
		cidrs = append(cidrs, *cidr4)
	}

	if networkIPv6 {
		cidrs = append(cidrs, *cidr6)
	}

	if len(cidrs) == 0 {
		return fmt.Errorf("neither IPv4 nor IPv6 network was enabled")
	}

	// Gateway addr at 1st IP in range, ex. 192.168.0.1
	gatewayIPs := make([]net.IP, len(cidrs))

	for j := range gatewayIPs {
		gatewayIPs[j], err = talosnet.NthIPInNetwork(&cidrs[j], gatewayOffset)
		if err != nil {
			return err
		}
	}

	// Set starting ip at 2nd ip in range, ex: 192.168.0.2
	ips := make([][]net.IP, len(cidrs))

	for j := range cidrs {
		ips[j] = make([]net.IP, masters+workers)

		for i := range ips[j] {
			ips[j][i], err = talosnet.NthIPInNetwork(&cidrs[j], nodesOffset+i)
			if err != nil {
				return err
			}
		}
	}

	// Parse nameservers
	nameserverIPs := make([]net.IP, len(nameservers))

	for i := range nameserverIPs {
		nameserverIPs[i] = net.ParseIP(nameservers[i])
		if nameserverIPs[i] == nil {
			return fmt.Errorf("failed parsing nameserver IP %q", nameservers[i])
		}
	}

	// Virtual (shared) IP at the vipOffset IP in range, ex. 192.168.0.50
	var vip net.IP

	if useVIP {
		vip, err = talosnet.NthIPInNetwork(&cidrs[0], vipOffset)
		if err != nil {
			return err
		}
	}

	provisioner, err := providers.Factory(ctx, provisionerName)
	if err != nil {
		return err
	}

	defer provisioner.Close() //nolint:errcheck

	// Craft cluster and node requests
	request := provision.ClusterRequest{
		Name: clusterName,

		Network: provision.NetworkRequest{
			Name:         clusterName,
			CIDRs:        cidrs,
			GatewayAddrs: gatewayIPs,
			MTU:          networkMTU,
			Nameservers:  nameserverIPs,
			CNI: provision.CNIConfig{
				BinPath:  cniBinPath,
				ConfDir:  cniConfDir,
				CacheDir: cniCacheDir,

				BundleURL: cniBundleURL,
			},
		},

		Image:         nodeImage,
		KernelPath:    nodeVmlinuzPath,
		InitramfsPath: nodeInitramfsPath,
		ISOPath:       nodeISOPath,
		DiskImagePath: nodeDiskImagePath,

		SelfExecutable: os.Args[0],
		StateDirectory: stateDir,
	}

	provisionOptions := []provision.Option{
		provision.WithDockerPortsHostIP(dockerHostIP),
		provision.WithBootlader(bootloaderEnabled),
		provision.WithUEFI(uefiEnabled),
		provision.WithTargetArch(targetArch),
	}
	configBundleOpts := []bundle.Option{}

	if ports != "" {
		if provisionerName != "docker" {
			return fmt.Errorf("exposed-ports flag only supported with docker provisioner")
		}

		portList := strings.Split(ports, ",")
		provisionOptions = append(provisionOptions, provision.WithDockerPorts(portList))
	}

	disks, err := getDisks()
	if err != nil {
		return err
	}

	if inputDir != "" {
		configBundleOpts = append(configBundleOpts, bundle.WithExistingConfigs(inputDir))
	} else {
		genOptions := []generate.GenOption{
			generate.WithInstallImage(nodeInstallImage),
			generate.WithDebug(configDebug),
			generate.WithDNSDomain(dnsDomain),
		}

		for _, registryMirror := range registryMirrors {
			components := strings.SplitN(registryMirror, "=", 2)
			if len(components) != 2 {
				return fmt.Errorf("invalid registry mirror spec: %q", registryMirror)
			}

			genOptions = append(genOptions, generate.WithRegistryMirror(components[0], components[1]))
		}

		for _, registryHost := range registryInsecure {
			genOptions = append(genOptions, generate.WithRegistryInsecureSkipVerify(registryHost))
		}

		genOptions = append(genOptions, provisioner.GenOptions(request.Network)...)

		if customCNIUrl != "" {
			genOptions = append(genOptions, generate.WithClusterCNIConfig(&v1alpha1.CNIConfig{
				CNIName: constants.CustomCNI,
				CNIUrls: []string{customCNIUrl},
			}))
		}

		if len(disks) > 1 {
			// convert provision disks to machine disks
			machineDisks := make([]*v1alpha1.MachineDisk, len(disks)-1)
			for i, disk := range disks[1:] {
				machineDisks[i] = &v1alpha1.MachineDisk{
					DeviceName:     provisioner.UserDiskName(i + 1),
					DiskPartitions: disk.Partitions,
				}
			}

			genOptions = append(genOptions, generate.WithUserDisks(machineDisks))
		}

		if talosVersion == "" {
			if provisionerName == "docker" {
				parts := strings.Split(nodeImage, ":")

				talosVersion = parts[len(parts)-1]
			} else {
				parts := strings.Split(nodeInstallImage, ":")

				talosVersion = parts[len(parts)-1]
			}
		}

		if talosVersion != "latest" {
			var versionContract *config.VersionContract

			versionContract, err = config.ParseContractFromVersion(talosVersion)
			if err != nil {
				return fmt.Errorf("error parsing Talos version %q: %w", talosVersion, err)
			}

			genOptions = append(genOptions, generate.WithVersionContract(versionContract))
		}

		if encryptStatePartition || encryptEphemeralPartition {
			diskEncryptionConfig := &v1alpha1.SystemDiskEncryptionConfig{}

			if encryptStatePartition {
				diskEncryptionConfig.StatePartition = &v1alpha1.EncryptionConfig{
					EncryptionProvider: encryption.LUKS2,
					EncryptionKeys: []*v1alpha1.EncryptionKey{
						{
							KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
							KeySlot:   0,
						},
					},
				}
			}

			if encryptEphemeralPartition {
				diskEncryptionConfig.EphemeralPartition = &v1alpha1.EncryptionConfig{
					EncryptionProvider: encryption.LUKS2,
					EncryptionKeys: []*v1alpha1.EncryptionKey{
						{
							KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
							KeySlot:   0,
						},
					},
				}
			}

			genOptions = append(genOptions, generate.WithSystemDiskEncryption(diskEncryptionConfig))
		}

		if useVIP {
			genOptions = append(genOptions,
				generate.WithNetworkOptions(
					v1alpha1.WithNetworkInterfaceVirtualIP(provisioner.GetFirstInterface(), vip.String()),
				),
			)
		}

		defaultInternalLB, defaultEndpoint := provisioner.GetLoadBalancers(request.Network)

		if defaultInternalLB == "" {
			// provisioner doesn't provide internal LB, so use first master node
			defaultInternalLB = ips[0][0].String()
		}

		if useVIP {
			defaultInternalLB = vip.String()
		}

		var endpointList []string

		switch {
		case defaultEndpoint != "":
			if forceEndpoint == "" {
				forceEndpoint = defaultEndpoint
			}

			fallthrough
		case forceEndpoint != "":
			endpointList = []string{forceEndpoint}
			// using non-default endpoints, provision additional cert SANs and fix endpoint list
			provisionOptions = append(provisionOptions, provision.WithEndpoint(forceEndpoint))
			genOptions = append(genOptions, generate.WithAdditionalSubjectAltNames(endpointList))
		case forceInitNodeAsEndpoint:
			endpointList = []string{ips[0][0].String()}
		default:
			// use control plane nodes as endpoints, client-side load-balancing
			for i := 0; i < masters; i++ {
				endpointList = append(endpointList, ips[0][i].String())
			}
		}

		genOptions = append(genOptions, generate.WithEndpointList(endpointList))
		configBundleOpts = append(configBundleOpts,
			bundle.WithInputOptions(
				&bundle.InputOptions{
					ClusterName: clusterName,
					Endpoint:    fmt.Sprintf("https://%s:%d", defaultInternalLB, constants.DefaultControlPlanePort),
					KubeVersion: strings.TrimPrefix(kubernetesVersion, "v"),
					GenOptions:  genOptions,
				}),
		)
	}

	addConfigPatch := func(configPatch string, configOpt func(jsonpatch.Patch) bundle.Option) error {
		if configPatch == "" {
			return nil
		}

		var jsonPatch jsonpatch.Patch

		jsonPatch, err = jsonpatch.DecodePatch([]byte(configPatch))
		if err != nil {
			return fmt.Errorf("error parsing config JSON patch: %w", err)
		}

		configBundleOpts = append(configBundleOpts, configOpt(jsonPatch))

		return nil
	}

	if err = addConfigPatch(configPatch, bundle.WithJSONPatch); err != nil {
		return err
	}

	if err = addConfigPatch(configPatchControlPlane, bundle.WithJSONPatchControlPlane); err != nil {
		return err
	}

	if err = addConfigPatch(configPatchWorker, bundle.WithJSONPatchWorker); err != nil {
		return err
	}

	configBundle, err := bundle.NewConfigBundle(configBundleOpts...)
	if err != nil {
		return err
	}

	if skipInjectingConfig {
		types := []machine.Type{machine.TypeControlPlane, machine.TypeWorker}

		if withInitNode {
			types = append([]machine.Type{machine.TypeInit}, types...)
		}

		if err = configBundle.Write(".", encoder.CommentsAll, types...); err != nil {
			return err
		}
	}

	// Wireguard configuration.
	var wireguardConfigBundle *helpers.WireguardConfigBundle
	if wireguardCIDR != "" {
		wireguardConfigBundle, err = helpers.NewWireguardConfigBundle(ips[0], wireguardCIDR, 51111, masters)
		if err != nil {
			return err
		}
	}

	// Add talosconfig to provision options so we'll have it to parse there
	provisionOptions = append(provisionOptions, provision.WithTalosConfig(configBundle.TalosConfig()))

	// Create the master nodes.
	for i := 0; i < masters; i++ {
		var cfg config.Provider

		nodeIPs := make([]net.IP, len(cidrs))
		for j := range nodeIPs {
			nodeIPs[j] = ips[j][i]
		}

		nodeReq := provision.NodeRequest{
			Name:                fmt.Sprintf("%s-master-%d", clusterName, i+1),
			Type:                machine.TypeControlPlane,
			IPs:                 nodeIPs,
			Memory:              memory,
			NanoCPUs:            nanoCPUs,
			Disks:               disks,
			SkipInjectingConfig: skipInjectingConfig,
			BadRTC:              badRTC,
		}

		if i == 0 {
			nodeReq.Ports = []string{"50000:50000/tcp", fmt.Sprintf("%d:%d/tcp", constants.DefaultControlPlanePort, constants.DefaultControlPlanePort)}
		}

		if withInitNode && i == 0 {
			cfg = configBundle.Init()
			nodeReq.Type = machine.TypeInit
		} else {
			cfg = configBundle.ControlPlane()
		}

		if wireguardConfigBundle != nil {
			cfg, err = wireguardConfigBundle.PatchConfig(nodeIPs[0], cfg)
			if err != nil {
				return err
			}
		}

		nodeReq.Config = cfg
		request.Nodes = append(request.Nodes, nodeReq)
	}

	for i := 1; i <= workers; i++ {
		name := fmt.Sprintf("%s-worker-%d", clusterName, i)

		cfg := configBundle.Worker()

		nodeIPs := make([]net.IP, len(cidrs))
		for j := range nodeIPs {
			nodeIPs[j] = ips[j][masters+i-1]
		}

		if wireguardConfigBundle != nil {
			cfg, err = wireguardConfigBundle.PatchConfig(nodeIPs[0], cfg)
			if err != nil {
				return err
			}
		}

		request.Nodes = append(request.Nodes,
			provision.NodeRequest{
				Name:                name,
				Type:                machine.TypeWorker,
				IPs:                 nodeIPs,
				Memory:              memory,
				NanoCPUs:            nanoCPUs,
				Disks:               disks,
				Config:              cfg,
				SkipInjectingConfig: skipInjectingConfig,
				BadRTC:              badRTC,
			})
	}

	cluster, err := provisioner.Create(ctx, request, provisionOptions...)
	if err != nil {
		return err
	}

	// Create and save the talosctl configuration file.
	if err = saveConfig(configBundle.TalosConfig()); err != nil {
		return err
	}

	clusterAccess := access.NewAdapter(cluster, provisionOptions...)
	defer clusterAccess.Close() //nolint:errcheck

	if applyConfigEnabled {
		err = clusterAccess.ApplyConfig(ctx, request.Nodes, os.Stdout)
		if err != nil {
			return err
		}
	}

	if err = postCreate(ctx, clusterAccess); err != nil {
		if crashdumpOnFailure {
			provisioner.CrashDump(ctx, cluster, os.Stderr)
		}

		return err
	}

	return showCluster(cluster)
}

func postCreate(ctx context.Context, clusterAccess *access.Adapter) error {
	if !withInitNode {
		if err := clusterAccess.Bootstrap(ctx, os.Stdout); err != nil {
			return fmt.Errorf("bootstrap error: %w", err)
		}
	}

	if !clusterWait {
		return nil
	}

	// Run cluster readiness checks
	checkCtx, checkCtxCancel := context.WithTimeout(ctx, clusterWaitTimeout)
	defer checkCtxCancel()

	if err := check.Wait(checkCtx, clusterAccess, append(check.DefaultClusterChecks(), check.ExtraClusterChecks()...), check.StderrReporter()); err != nil {
		return err
	}

	if !skipKubeconfig {
		if err := mergeKubeconfig(ctx, clusterAccess); err != nil {
			return err
		}
	}

	return nil
}

func saveConfig(talosConfigObj *clientconfig.Config) (err error) {
	c, err := clientconfig.Open(talosconfig)
	if err != nil {
		return err
	}

	renames := c.Merge(talosConfigObj)
	for _, rename := range renames {
		fmt.Printf("renamed talosconfig context %s\n", rename.String())
	}

	return c.Save(talosconfig)
}

func mergeKubeconfig(ctx context.Context, clusterAccess *access.Adapter) error {
	kubeconfigPath, err := kubeconfig.DefaultPath()
	if err != nil {
		return err
	}

	fmt.Printf("\nmerging kubeconfig into %q\n", kubeconfigPath)

	k8sconfig, err := clusterAccess.Kubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("error fetching kubeconfig: %w", err)
	}

	config, err := clientcmd.Load(k8sconfig)
	if err != nil {
		return fmt.Errorf("error parsing kubeconfig: %w", err)
	}

	if clusterAccess.ForceEndpoint != "" {
		for name := range config.Clusters {
			config.Clusters[name].Server = fmt.Sprintf("https://%s:%d", clusterAccess.ForceEndpoint, constants.DefaultControlPlanePort)
		}
	}

	_, err = os.Stat(kubeconfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		return clientcmd.WriteToFile(*config, kubeconfigPath)
	}

	merger, err := kubeconfig.Load(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error loading existing kubeconfig: %w", err)
	}

	err = merger.Merge(config, kubeconfig.MergeOptions{
		ActivateContext: true,
		OutputWriter:    os.Stdout,
		ConflictHandler: func(component kubeconfig.ConfigComponent, name string) (kubeconfig.ConflictDecision, error) {
			return kubeconfig.RenameDecision, nil
		},
	})
	if err != nil {
		return fmt.Errorf("error merging kubeconfig: %w", err)
	}

	return merger.Write(kubeconfigPath)
}

func parseCPUShare() (int64, error) {
	cpu, ok := new(big.Rat).SetString(clusterCpus)
	if !ok {
		return 0, fmt.Errorf("failed to parsing as a rational number: %s", clusterCpus)
	}

	nano := cpu.Mul(cpu, big.NewRat(1e9, 1))
	if !nano.IsInt() {
		return 0, errors.New("value is too precise")
	}

	return nano.Num().Int64(), nil
}

func getDisks() ([]*provision.Disk, error) {
	// should have at least a single primary disk
	disks := []*provision.Disk{
		{
			Size: uint64(clusterDiskSize) * 1024 * 1024,
		},
	}

	for _, disk := range clusterDisks {
		var (
			partitions     = strings.Split(disk, ":")
			diskPartitions = make([]*v1alpha1.DiskPartition, len(partitions)/2)
			diskSize       uint64
		)

		if len(partitions)%2 != 0 {
			return nil, fmt.Errorf("failed to parse malformed partition definitions")
		}

		partitionIndex := 0

		for j := 0; j < len(partitions); j += 2 {
			partitionPath := partitions[j]

			if !strings.HasPrefix(partitionPath, "/var") {
				return nil, fmt.Errorf("user disk partitions can only be mounted into /var folder")
			}

			value, e := strconv.ParseInt(partitions[j+1], 10, -1)
			partitionSize := uint64(value)

			if e != nil {
				partitionSize, e = humanize.ParseBytes(partitions[j+1])

				if e != nil {
					return nil, fmt.Errorf("failed to parse partition size")
				}
			}

			diskPartitions[partitionIndex] = &v1alpha1.DiskPartition{
				DiskSize:       v1alpha1.DiskSize(partitionSize),
				DiskMountPoint: partitionPath,
			}
			diskSize += partitionSize
			partitionIndex++ //nolint:wastedassign
		}

		disks = append(disks, &provision.Disk{
			// add 1 MB to make extra room for GPT
			Size:       diskSize + 1024*1024,
			Partitions: diskPartitions,
		})
	}

	return disks, nil
}

func trimVersion(version string) string {
	// remove anything extra after semantic version core, `v0.3.2-1-abcd` -> `v0.3.2`
	return regexp.MustCompile(`(-\d+(-g[0-9a-f]+)?(-dirty)?)$`).ReplaceAllString(version, "")
}

func init() {
	defaultTalosConfig, err := clientconfig.GetDefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find default Talos config path: %s", err)
	}

	createCmd.Flags().StringVar(&talosconfig, "talosconfig", defaultTalosConfig, "The path to the Talos configuration file")
	createCmd.Flags().StringVar(&nodeImage, "image", helpers.DefaultImage(images.DefaultTalosImageRepository), "the image to use")
	createCmd.Flags().StringVar(&nodeInstallImage, "install-image", helpers.DefaultImage(images.DefaultInstallerImageRepository), "the installer image to use")
	createCmd.Flags().StringVar(&nodeVmlinuzPath, "vmlinuz-path", helpers.ArtifactPath(constants.KernelAssetWithArch), "the compressed kernel image to use")
	createCmd.Flags().StringVar(&nodeISOPath, "iso-path", "", "the ISO path to use for the initial boot (VM only)")
	createCmd.Flags().StringVar(&nodeInitramfsPath, "initrd-path", helpers.ArtifactPath(constants.InitramfsAssetWithArch), "initramfs image to use")
	createCmd.Flags().StringVar(&nodeDiskImagePath, "disk-image-path", "", "disk image to use")
	createCmd.Flags().BoolVar(&applyConfigEnabled, "with-apply-config", false, "enable apply config when the VM is starting in maintenance mode")
	createCmd.Flags().BoolVar(&bootloaderEnabled, "with-bootloader", true, "enable bootloader to load kernel and initramfs from disk image after install")
	createCmd.Flags().BoolVar(&uefiEnabled, "with-uefi", false, "enable UEFI on x86_64 architecture (always enabled for arm64)")
	createCmd.Flags().StringSliceVar(&registryMirrors, "registry-mirror", []string{}, "list of registry mirrors to use in format: <registry host>=<mirror URL>")
	createCmd.Flags().StringSliceVar(&registryInsecure, "registry-insecure-skip-verify", []string{}, "list of registry hostnames to skip TLS verification for")
	createCmd.Flags().BoolVar(&configDebug, "with-debug", false, "enable debug in Talos config to send service logs to the console")
	createCmd.Flags().IntVar(&networkMTU, "mtu", 1500, "MTU of the cluster network")
	createCmd.Flags().StringVar(&networkCIDR, "cidr", "10.5.0.0/24", "CIDR of the cluster network (IPv4, ULA network for IPv6 is derived in automated way)")
	createCmd.Flags().BoolVar(&networkIPv4, "ipv4", true, "enable IPv4 network in the cluster")
	createCmd.Flags().BoolVar(&networkIPv6, "ipv6", false, "enable IPv6 network in the cluster (QEMU provisioner only)")
	createCmd.Flags().StringVar(&wireguardCIDR, "wireguard-cidr", "", "CIDR of the wireguard network")
	createCmd.Flags().StringSliceVar(&nameservers, "nameservers", []string{"8.8.8.8", "1.1.1.1", "2001:4860:4860::8888", "2606:4700:4700::1111"}, "list of nameservers to use")
	createCmd.Flags().IntVar(&workers, "workers", 1, "the number of workers to create")
	createCmd.Flags().IntVar(&masters, "masters", 1, "the number of masters to create")
	createCmd.Flags().StringVar(&clusterCpus, "cpus", "2.0", "the share of CPUs as fraction (each container/VM)")
	createCmd.Flags().IntVar(&clusterMemory, "memory", 2048, "the limit on memory usage in MB (each container/VM)")
	createCmd.Flags().IntVar(&clusterDiskSize, "disk", 6*1024, "default limit on disk size in MB (each VM)")
	createCmd.Flags().StringSliceVar(&clusterDisks, "user-disk", []string{}, "list of disks to create for each VM in format: <mount_point1>:<size1>:<mount_point2>:<size2>")
	createCmd.Flags().StringVar(&targetArch, "arch", stdruntime.GOARCH, "cluster architecture")
	createCmd.Flags().BoolVar(&clusterWait, "wait", true, "wait for the cluster to be ready before returning")
	createCmd.Flags().DurationVar(&clusterWaitTimeout, "wait-timeout", 20*time.Minute, "timeout to wait for the cluster to be ready")
	createCmd.Flags().BoolVar(&forceInitNodeAsEndpoint, "init-node-as-endpoint", false, "use init node as endpoint instead of any load balancer endpoint")
	createCmd.Flags().StringVar(&forceEndpoint, "endpoint", "", "use endpoint instead of provider defaults")
	createCmd.Flags().StringVar(&kubernetesVersion, "kubernetes-version", "", fmt.Sprintf("desired kubernetes version to run (default %q)", constants.DefaultKubernetesVersion))
	createCmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "location of pre-generated config files")
	createCmd.Flags().StringSliceVar(&cniBinPath, "cni-bin-path", []string{filepath.Join(defaultCNIDir, "bin")}, "search path for CNI binaries (VM only)")
	createCmd.Flags().StringVar(&cniConfDir, "cni-conf-dir", filepath.Join(defaultCNIDir, "conf.d"), "CNI config directory path (VM only)")
	createCmd.Flags().StringVar(&cniCacheDir, "cni-cache-dir", filepath.Join(defaultCNIDir, "cache"), "CNI cache directory path (VM only)")
	createCmd.Flags().StringVar(&cniBundleURL, "cni-bundle-url", fmt.Sprintf("https://github.com/talos-systems/talos/releases/download/%s/talosctl-cni-bundle-%s.tar.gz",
		trimVersion(version.Tag), constants.ArchVariable), "URL to download CNI bundle from (VM only)")
	createCmd.Flags().StringVarP(&ports,
		"exposed-ports",
		"p",
		"",
		"Comma-separated list of ports/protocols to expose on init node. Ex -p <hostPort>:<containerPort>/<protocol (tcp or udp)> (Docker provisioner only)",
	)
	createCmd.Flags().StringVar(&dockerHostIP, "docker-host-ip", "0.0.0.0", "Host IP to forward exposed ports to (Docker provisioner only)")
	createCmd.Flags().BoolVar(&withInitNode, "with-init-node", false, "create the cluster with an init node")
	createCmd.Flags().StringVar(&customCNIUrl, "custom-cni-url", "", "install custom CNI from the URL (Talos cluster)")
	createCmd.Flags().StringVar(&dnsDomain, "dns-domain", "cluster.local", "the dns domain to use for cluster")
	createCmd.Flags().BoolVar(&crashdumpOnFailure, "crashdump", false, "print debug crashdump to stderr when cluster startup fails")
	createCmd.Flags().BoolVar(&skipKubeconfig, "skip-kubeconfig", false, "skip merging kubeconfig from the created cluster")
	createCmd.Flags().BoolVar(&skipInjectingConfig, "skip-injecting-config", false, "skip injecting config from embedded metadata server, write config files to current directory")
	createCmd.Flags().BoolVar(&encryptStatePartition, "encrypt-state", false, "enable state partition encryption")
	createCmd.Flags().BoolVar(&encryptEphemeralPartition, "encrypt-ephemeral", false, "enable ephemeral partition encryption")
	createCmd.Flags().StringVar(&talosVersion, "talos-version", "", "the desired Talos version to generate config for (if not set, defaults to image version)")
	createCmd.Flags().BoolVar(&useVIP, "use-vip", false, "use a virtual IP for the controlplane endpoint instead of the loadbalancer")
	createCmd.Flags().StringVar(&configPatch, "config-patch", "", "patch generated machineconfigs (applied to all node types)")
	createCmd.Flags().StringVar(&configPatchControlPlane, "config-patch-control-plane", "", "patch generated machineconfigs (applied to 'init' and 'controlplane' types)")
	createCmd.Flags().StringVar(&configPatchWorker, "config-patch-worker", "", "patch generated machineconfigs (applied to 'worker' type)")
	createCmd.Flags().BoolVar(&badRTC, "bad-rtc", false, "launch VM with bad RTC state (QEMU only)")

	// remove in 0.13: https://github.com/talos-systems/talos/issues/3910
	createCmd.Flags().StringVar(&configPatchJoin, "config-patch-join", "", "")
	cli.Should(createCmd.Flags().MarkDeprecated("config-patch-join", "use --config-patch-worker instead"))

	Cmd.AddCommand(createCmd)
}
