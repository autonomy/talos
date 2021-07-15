// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build integration_provision

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/talos-systems/go-blockdevice/blockdevice/encryption"
	"github.com/talos-systems/go-retry/retry"
	talosnet "github.com/talos-systems/net"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/talos-systems/talos/cmd/talosctl/pkg/mgmt/helpers"
	"github.com/talos-systems/talos/internal/integration/base"
	"github.com/talos-systems/talos/pkg/cluster/check"
	"github.com/talos-systems/talos/pkg/cluster/kubernetes"
	"github.com/talos-systems/talos/pkg/cluster/sonobuoy"
	"github.com/talos-systems/talos/pkg/images"
	machineapi "github.com/talos-systems/talos/pkg/machinery/api/machine"
	talosclient "github.com/talos-systems/talos/pkg/machinery/client"
	clientconfig "github.com/talos-systems/talos/pkg/machinery/client/config"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/bundle"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/talos-systems/talos/pkg/provision"
	"github.com/talos-systems/talos/pkg/provision/access"
	"github.com/talos-systems/talos/pkg/provision/providers/qemu"
)

//nolint:maligned
type upgradeSpec struct {
	ShortName string

	SourceKernelPath     string
	SourceInitramfsPath  string
	SourceInstallerImage string
	SourceVersion        string
	SourceK8sVersion     string

	TargetInstallerImage string
	TargetVersion        string
	TargetK8sVersion     string

	MasterNodes int
	WorkerNodes int

	UpgradePreserve bool
	UpgradeStage    bool
	WithEncryption  bool
}

const (
	previousRelease = "v0.10.4"
	stableRelease   = "v0.11.1" // or soon-to-be-stable
	// The current version (the one being built on CI) is DefaultSettings.CurrentVersion.

	previousK8sVersion = "1.21.1" // constants.DefaultKubernetesVersion in the previousRelease
	stableK8sVersion   = "1.21.2" // constants.DefaultKubernetesVersion in the stableRelease
	currentK8sVersion  = constants.DefaultKubernetesVersion
)

var defaultNameservers = []net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("1.1.1.1")}

// upgradePreviousToStable upgrades from the previous Talos release to the stable release.
func upgradePreviousToStable() upgradeSpec {
	return upgradeSpec{
		ShortName: fmt.Sprintf("%s-%s", previousRelease, stableRelease),

		SourceKernelPath:     helpers.ArtifactPath(filepath.Join(trimVersion(previousRelease), constants.KernelAsset)),
		SourceInitramfsPath:  helpers.ArtifactPath(filepath.Join(trimVersion(previousRelease), constants.InitramfsAsset)),
		SourceInstallerImage: fmt.Sprintf("%s:%s", "ghcr.io/talos-systems/installer", previousRelease),
		SourceVersion:        previousRelease,
		SourceK8sVersion:     previousK8sVersion,

		TargetInstallerImage: fmt.Sprintf("%s:%s", "ghcr.io/talos-systems/installer", stableRelease),
		TargetVersion:        stableRelease,
		TargetK8sVersion:     stableK8sVersion,

		MasterNodes: DefaultSettings.MasterNodes,
		WorkerNodes: DefaultSettings.WorkerNodes,
	}
}

// upgradeStableToCurrent upgrades from the stable Talos release to the current version.
func upgradeStableToCurrent() upgradeSpec {
	return upgradeSpec{
		ShortName: fmt.Sprintf("%s-%s", stableRelease, DefaultSettings.CurrentVersion),

		SourceKernelPath:     helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.KernelAsset)),
		SourceInitramfsPath:  helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.InitramfsAsset)),
		SourceInstallerImage: fmt.Sprintf("%s:%s", "ghcr.io/talos-systems/installer", stableRelease),
		SourceVersion:        stableRelease,
		SourceK8sVersion:     stableK8sVersion,

		TargetInstallerImage: fmt.Sprintf("%s/%s:%s", DefaultSettings.TargetInstallImageRegistry, images.DefaultInstallerImageName, DefaultSettings.CurrentVersion),
		TargetVersion:        DefaultSettings.CurrentVersion,
		TargetK8sVersion:     currentK8sVersion,

		MasterNodes: DefaultSettings.MasterNodes,
		WorkerNodes: DefaultSettings.WorkerNodes,
	}
}

// upgradeCurrentToCurrent upgrades the current version to itself.
func upgradeCurrentToCurrent() upgradeSpec {
	installerImage := fmt.Sprintf("%s/%s:%s", DefaultSettings.TargetInstallImageRegistry, images.DefaultInstallerImageName, DefaultSettings.CurrentVersion)

	return upgradeSpec{
		ShortName: fmt.Sprintf("%s-%s", DefaultSettings.CurrentVersion, DefaultSettings.CurrentVersion),

		SourceKernelPath:     helpers.ArtifactPath(constants.KernelAssetWithArch),
		SourceInitramfsPath:  helpers.ArtifactPath(constants.InitramfsAssetWithArch),
		SourceInstallerImage: installerImage,
		SourceVersion:        DefaultSettings.CurrentVersion,
		SourceK8sVersion:     currentK8sVersion,

		TargetInstallerImage: installerImage,
		TargetVersion:        DefaultSettings.CurrentVersion,
		TargetK8sVersion:     currentK8sVersion,

		MasterNodes: DefaultSettings.MasterNodes,
		WorkerNodes: DefaultSettings.WorkerNodes,

		WithEncryption: true,
	}
}

// upgradeStableToCurrentPreserve upgrades from the stable Talos release to the current version for single-node cluster with preserve.
func upgradeStableToCurrentPreserve() upgradeSpec {
	return upgradeSpec{
		ShortName: fmt.Sprintf("prsrv-%s-%s", stableRelease, DefaultSettings.CurrentVersion),

		SourceKernelPath:     helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.KernelAsset)),
		SourceInitramfsPath:  helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.InitramfsAsset)),
		SourceInstallerImage: fmt.Sprintf("%s:%s", "ghcr.io/talos-systems/installer", stableRelease),
		SourceVersion:        stableRelease,
		SourceK8sVersion:     stableK8sVersion,

		TargetInstallerImage: fmt.Sprintf("%s/%s:%s", DefaultSettings.TargetInstallImageRegistry, images.DefaultInstallerImageName, DefaultSettings.CurrentVersion),
		TargetVersion:        DefaultSettings.CurrentVersion,
		TargetK8sVersion:     currentK8sVersion,

		MasterNodes:     1,
		WorkerNodes:     0,
		UpgradePreserve: true,
	}
}

// upgradeStableToCurrentPreserveStage upgrades from the stable Talos release to the current version for single-node cluster with preserve and stage.
func upgradeStableToCurrentPreserveStage() upgradeSpec {
	return upgradeSpec{
		ShortName: fmt.Sprintf("prsrv-stg-%s-%s", stableRelease, DefaultSettings.CurrentVersion),

		SourceKernelPath:     helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.KernelAsset)),
		SourceInitramfsPath:  helpers.ArtifactPath(filepath.Join(trimVersion(stableRelease), constants.InitramfsAsset)),
		SourceInstallerImage: fmt.Sprintf("%s:%s", "ghcr.io/talos-systems/installer", stableRelease),
		SourceVersion:        stableRelease,
		SourceK8sVersion:     stableK8sVersion,

		TargetInstallerImage: fmt.Sprintf("%s/%s:%s", DefaultSettings.TargetInstallImageRegistry, images.DefaultInstallerImageName, DefaultSettings.CurrentVersion),
		TargetVersion:        DefaultSettings.CurrentVersion,
		TargetK8sVersion:     currentK8sVersion,

		MasterNodes:     1,
		WorkerNodes:     0,
		UpgradePreserve: true,
		UpgradeStage:    true,
	}
}

// UpgradeSuite ...
type UpgradeSuite struct {
	suite.Suite
	base.TalosSuite

	specGen func() upgradeSpec
	spec    upgradeSpec

	track int

	provisioner provision.Provisioner

	configBundle *v1alpha1.ConfigBundle

	clusterAccess        *access.Adapter
	controlPlaneEndpoint string

	ctx       context.Context
	ctxCancel context.CancelFunc

	stateDir string
	cniDir   string
}

// SetupSuite ...
func (suite *UpgradeSuite) SetupSuite() {
	// call generate late in the flow, as it needs to pick up settings overridden by test runner
	suite.spec = suite.specGen()

	suite.T().Logf("upgrade spec = %v", suite.spec)

	// timeout for the whole test
	suite.ctx, suite.ctxCancel = context.WithTimeout(context.Background(), 30*time.Minute)

	var err error

	suite.provisioner, err = qemu.NewProvisioner(suite.ctx)
	suite.Require().NoError(err)
}

// TearDownSuite ...
func (suite *UpgradeSuite) TearDownSuite() {
	if suite.T().Failed() && DefaultSettings.CrashdumpEnabled && suite.Cluster != nil {
		// for failed tests, produce crash dump for easier debugging,
		// as cluster is going to be torn down below
		suite.provisioner.CrashDump(suite.ctx, suite.Cluster, os.Stderr)

		if suite.clusterAccess != nil {
			suite.clusterAccess.CrashDump(suite.ctx, os.Stderr)
		}
	}

	if suite.clusterAccess != nil {
		suite.Assert().NoError(suite.clusterAccess.Close())
	}

	if suite.Cluster != nil {
		suite.Assert().NoError(suite.provisioner.Destroy(suite.ctx, suite.Cluster))
	}

	suite.ctxCancel()

	if suite.stateDir != "" {
		suite.Assert().NoError(os.RemoveAll(suite.stateDir))
	}

	if suite.provisioner != nil {
		suite.Assert().NoError(suite.provisioner.Close())
	}
}

// setupCluster provisions source clusters and waits for health.
func (suite *UpgradeSuite) setupCluster() {
	defaultStateDir, err := clientconfig.GetTalosDirectory()
	suite.Require().NoError(err)

	suite.stateDir = filepath.Join(defaultStateDir, "clusters")
	suite.cniDir = filepath.Join(defaultStateDir, "cni")

	clusterName := suite.spec.ShortName

	_, cidr, err := net.ParseCIDR(DefaultSettings.CIDR)
	suite.Require().NoError(err)

	var gatewayIP net.IP

	gatewayIP, err = talosnet.NthIPInNetwork(cidr, 1)
	suite.Require().NoError(err)

	ips := make([]net.IP, suite.spec.MasterNodes+suite.spec.WorkerNodes)

	for i := range ips {
		ips[i], err = talosnet.NthIPInNetwork(cidr, i+2)
		suite.Require().NoError(err)
	}

	suite.T().Logf("initializing provisioner with cluster name %q, state directory %q", clusterName, suite.stateDir)

	request := provision.ClusterRequest{
		Name: clusterName,

		Network: provision.NetworkRequest{
			Name:         clusterName,
			CIDRs:        []net.IPNet{*cidr},
			GatewayAddrs: []net.IP{gatewayIP},
			MTU:          DefaultSettings.MTU,
			Nameservers:  defaultNameservers,
			CNI: provision.CNIConfig{
				BinPath:  []string{filepath.Join(suite.cniDir, "bin")},
				ConfDir:  filepath.Join(suite.cniDir, "conf.d"),
				CacheDir: filepath.Join(suite.cniDir, "cache"),

				BundleURL: DefaultSettings.CNIBundleURL,
			},
		},

		KernelPath:    suite.spec.SourceKernelPath,
		InitramfsPath: suite.spec.SourceInitramfsPath,

		SelfExecutable: suite.TalosctlPath,
		StateDirectory: suite.stateDir,
	}

	defaultInternalLB, _ := suite.provisioner.GetLoadBalancers(request.Network)
	suite.controlPlaneEndpoint = fmt.Sprintf("https://%s:%d", defaultInternalLB, constants.DefaultControlPlanePort)

	genOptions := suite.provisioner.GenOptions(request.Network)

	for _, registryMirror := range DefaultSettings.RegistryMirrors {
		parts := strings.SplitN(registryMirror, "=", 2)
		suite.Require().Len(parts, 2)

		genOptions = append(genOptions, generate.WithRegistryMirror(parts[0], parts[1]))
	}

	masterEndpoints := make([]string, suite.spec.MasterNodes)
	for i := range masterEndpoints {
		masterEndpoints[i] = ips[i].String()
	}

	if DefaultSettings.CustomCNIURL != "" {
		genOptions = append(genOptions, generate.WithClusterCNIConfig(&v1alpha1.CNIConfig{
			CNIName: constants.CustomCNI,
			CNIUrls: []string{DefaultSettings.CustomCNIURL},
		}))
	}

	if suite.spec.WithEncryption {
		genOptions = append(genOptions, generate.WithSystemDiskEncryption(&v1alpha1.SystemDiskEncryptionConfig{
			StatePartition: &v1alpha1.EncryptionConfig{
				EncryptionProvider: encryption.LUKS2,
				EncryptionKeys: []*v1alpha1.EncryptionKey{
					{
						KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
						KeySlot:   0,
					},
				},
			},
			EphemeralPartition: &v1alpha1.EncryptionConfig{
				EncryptionProvider: encryption.LUKS2,
				EncryptionKeys: []*v1alpha1.EncryptionKey{
					{
						KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
						KeySlot:   0,
					},
				},
			},
		}))
	}

	versionContract, err := config.ParseContractFromVersion(suite.spec.SourceVersion)
	suite.Require().NoError(err)

	suite.configBundle, err = bundle.NewConfigBundle(bundle.WithInputOptions(
		&bundle.InputOptions{
			ClusterName: clusterName,
			Endpoint:    suite.controlPlaneEndpoint,
			KubeVersion: suite.spec.SourceK8sVersion,
			GenOptions: append(
				genOptions,
				generate.WithEndpointList(masterEndpoints),
				generate.WithInstallImage(suite.spec.SourceInstallerImage),
				generate.WithDNSDomain("cluster.local"),
				generate.WithVersionContract(versionContract),
			),
		}))
	suite.Require().NoError(err)

	for i := 0; i < suite.spec.MasterNodes; i++ {
		request.Nodes = append(request.Nodes,
			provision.NodeRequest{
				Name:     fmt.Sprintf("master-%d", i+1),
				Type:     machine.TypeControlPlane,
				IPs:      []net.IP{ips[i]},
				Memory:   DefaultSettings.MemMB * 1024 * 1024,
				NanoCPUs: DefaultSettings.CPUs * 1000 * 1000 * 1000,
				Disks: []*provision.Disk{
					{
						Size: DefaultSettings.DiskGB * 1024 * 1024 * 1024,
					},
				},
				Config: suite.configBundle.ControlPlane(),
			})
	}

	for i := 1; i <= suite.spec.WorkerNodes; i++ {
		request.Nodes = append(request.Nodes,
			provision.NodeRequest{
				Name:     fmt.Sprintf("worker-%d", i),
				Type:     machine.TypeWorker,
				IPs:      []net.IP{ips[suite.spec.MasterNodes+i-1]},
				Memory:   DefaultSettings.MemMB * 1024 * 1024,
				NanoCPUs: DefaultSettings.CPUs * 1000 * 1000 * 1000,
				Disks: []*provision.Disk{
					{
						Size: DefaultSettings.DiskGB * 1024 * 1024 * 1024,
					},
				},
				Config: suite.configBundle.Worker(),
			})
	}

	suite.Cluster, err = suite.provisioner.Create(suite.ctx, request, provision.WithBootlader(true), provision.WithTalosConfig(suite.configBundle.TalosConfig()))
	suite.Require().NoError(err)

	defaultTalosConfig, err := clientconfig.GetDefaultPath()
	suite.Require().NoError(err)

	c, err := clientconfig.Open(defaultTalosConfig)
	suite.Require().NoError(err)

	c.Merge(suite.configBundle.TalosConfig())

	suite.Require().NoError(c.Save(defaultTalosConfig))

	suite.clusterAccess = access.NewAdapter(suite.Cluster, provision.WithTalosConfig(suite.configBundle.TalosConfig()))

	suite.Require().NoError(suite.clusterAccess.Bootstrap(suite.ctx, os.Stdout))

	suite.waitForClusterHealth()
}

// waitForClusterHealth asserts cluster health after any change.
func (suite *UpgradeSuite) waitForClusterHealth() {
	runs := 1

	singleNodeCluster := len(suite.Cluster.Info().Nodes) == 1
	if singleNodeCluster {
		// run health check several times for single node clusters,
		// as self-hosted control plane is not stable after reboot
		runs = 3
	}

	for run := 0; run < runs; run++ {
		if run > 0 {
			time.Sleep(15 * time.Second)
		}

		checkCtx, checkCtxCancel := context.WithTimeout(suite.ctx, 15*time.Minute)
		defer checkCtxCancel()

		suite.Require().NoError(check.Wait(checkCtx, suite.clusterAccess, check.DefaultClusterChecks(), check.StderrReporter()))
	}
}

// runE2E runs e2e test on the cluster.
func (suite *UpgradeSuite) runE2E(k8sVersion string) {
	if suite.spec.WorkerNodes == 0 {
		// no worker nodes, should make masters schedulable
		suite.untaint("master-1")
	}

	options := sonobuoy.DefaultOptions()
	options.KubernetesVersion = k8sVersion

	suite.Assert().NoError(sonobuoy.Run(suite.ctx, suite.clusterAccess, options))
}

func (suite *UpgradeSuite) assertSameVersionCluster(client *talosclient.Client, expectedVersion string) {
	nodes := make([]string, len(suite.Cluster.Info().Nodes))

	for i, node := range suite.Cluster.Info().Nodes {
		nodes[i] = node.IPs[0].String()
	}

	ctx := talosclient.WithNodes(suite.ctx, nodes...)

	var v *machineapi.VersionResponse

	err := retry.Constant(
		time.Minute,
	).Retry(func() error {
		var e error
		v, e = client.Version(ctx)

		return retry.ExpectedError(e)
	})

	suite.Require().NoError(err)

	suite.Require().Len(v.Messages, len(nodes))

	for _, version := range v.Messages {
		suite.Assert().Equal(expectedVersion, version.Version.Tag)
	}
}

func (suite *UpgradeSuite) readVersion(nodeCtx context.Context, client *talosclient.Client) (version string, err error) {
	var v *machineapi.VersionResponse

	v, err = client.Version(nodeCtx)
	if err != nil {
		return
	}

	version = v.Messages[0].Version.Tag

	return
}

func (suite *UpgradeSuite) upgradeNode(client *talosclient.Client, node provision.NodeInfo) {
	suite.T().Logf("upgrading node %s", node.IPs[0])

	nodeCtx := talosclient.WithNodes(suite.ctx, node.IPs[0].String())

	var (
		resp *machineapi.UpgradeResponse
		err  error
	)

	err = retry.Constant(time.Minute, retry.WithUnits(10*time.Second)).Retry(func() error {
		resp, err = client.Upgrade(nodeCtx, suite.spec.TargetInstallerImage, suite.spec.UpgradePreserve, suite.spec.UpgradeStage, false)
		if err != nil {
			if strings.Contains(err.Error(), "leader changed") {
				return retry.ExpectedError(err)
			}

			return err
		}

		return nil
	})

	err = base.IgnoreGRPCUnavailable(err)
	suite.Require().NoError(err)

	if resp != nil {
		suite.Require().Equal("Upgrade request received", resp.Messages[0].Ack)
	}

	// wait for the upgrade to be kicked off
	time.Sleep(10 * time.Second)

	// wait for the version to be equal to target version
	suite.Require().NoError(retry.Constant(10 * time.Minute).Retry(func() error {
		var version string

		version, err = suite.readVersion(nodeCtx, client)
		if err != nil {
			// API might be unresponsive during upgrade
			return retry.ExpectedError(err)
		}

		if version != suite.spec.TargetVersion {
			// upgrade not finished yet
			return retry.ExpectedError(fmt.Errorf("node %q version doesn't match expected: expected %q, got %q", node.IPs[0].String(), suite.spec.TargetVersion, version))
		}

		return nil
	}))

	suite.waitForClusterHealth()
}

func (suite *UpgradeSuite) upgradeKubernetes(fromVersion, toVersion string) {
	if fromVersion == toVersion {
		suite.T().Logf("skipping Kubernetes upgrade, as versions are equal %q -> %q", fromVersion, toVersion)

		return
	}

	suite.T().Logf("upgrading Kubernetes: %q -> %q", fromVersion, toVersion)

	options := kubernetes.UpgradeOptions{
		FromVersion: fromVersion,
		ToVersion:   toVersion,

		ControlPlaneEndpoint: suite.controlPlaneEndpoint,
	}

	suite.Require().NoError(kubernetes.UpgradeTalosManaged(suite.ctx, suite.clusterAccess, options))
}

func (suite *UpgradeSuite) untaint(name string) {
	client, err := suite.clusterAccess.K8sClient(suite.ctx)
	suite.Require().NoError(err)

	n, err := client.CoreV1().Nodes().Get(suite.ctx, name, metav1.GetOptions{})
	suite.Require().NoError(err)

	oldData, err := json.Marshal(n)
	suite.Require().NoError(err)

	k := 0

	for _, taint := range n.Spec.Taints {
		if taint.Key != constants.LabelNodeRoleMaster {
			n.Spec.Taints[k] = taint
			k++
		}
	}

	n.Spec.Taints = n.Spec.Taints[:k]

	newData, err := json.Marshal(n)
	suite.Require().NoError(err)

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Node{})
	suite.Require().NoError(err)

	_, err = client.CoreV1().Nodes().Patch(suite.ctx, n.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	suite.Require().NoError(err)
}

// TestRolling performs rolling upgrade starting with master nodes.
func (suite *UpgradeSuite) TestRolling() {
	suite.setupCluster()

	client, err := suite.clusterAccess.Client()
	suite.Require().NoError(err)

	// verify initial cluster version
	suite.assertSameVersionCluster(client, suite.spec.SourceVersion)

	// upgrade master nodes
	for _, node := range suite.Cluster.Info().Nodes {
		if node.Type == machine.TypeInit || node.Type == machine.TypeControlPlane {
			suite.upgradeNode(client, node)
		}
	}

	// upgrade worker nodes
	for _, node := range suite.Cluster.Info().Nodes {
		if node.Type == machine.TypeWorker {
			suite.upgradeNode(client, node)
		}
	}

	// verify final cluster version
	suite.assertSameVersionCluster(client, suite.spec.TargetVersion)

	// upgrade Kubernetes if required
	suite.upgradeKubernetes(suite.spec.SourceK8sVersion, suite.spec.TargetK8sVersion)

	// run e2e test
	suite.runE2E(suite.spec.TargetK8sVersion)
}

// SuiteName ...
func (suite *UpgradeSuite) SuiteName() string {
	if suite.spec.ShortName == "" {
		suite.spec = suite.specGen()
	}

	return fmt.Sprintf("provision.UpgradeSuite.%s-TR%d", suite.spec.ShortName, suite.track)
}

func init() {
	allSuites = append(allSuites,
		&UpgradeSuite{specGen: upgradePreviousToStable, track: 0},
		&UpgradeSuite{specGen: upgradeStableToCurrent, track: 1},
		&UpgradeSuite{specGen: upgradeCurrentToCurrent, track: 2},
		&UpgradeSuite{specGen: upgradeStableToCurrentPreserve, track: 0},
		&UpgradeSuite{specGen: upgradeStableToCurrentPreserveStage, track: 1},
	)
}
