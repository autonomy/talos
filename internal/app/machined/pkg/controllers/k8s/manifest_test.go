// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//nolint:dupl
package k8s_test

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cosi-project/runtime/pkg/controller/runtime"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/cosi-project/runtime/pkg/state/impl/inmem"
	"github.com/cosi-project/runtime/pkg/state/impl/namespaced"
	"github.com/stretchr/testify/suite"
	"github.com/talos-systems/go-retry/retry"

	k8sctrl "github.com/talos-systems/talos/internal/app/machined/pkg/controllers/k8s"
	"github.com/talos-systems/talos/pkg/logging"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"github.com/talos-systems/talos/pkg/resources/config"
	"github.com/talos-systems/talos/pkg/resources/k8s"
	"github.com/talos-systems/talos/pkg/resources/secrets"
)

type ManifestSuite struct {
	suite.Suite

	state state.State

	runtime *runtime.Runtime
	wg      sync.WaitGroup

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (suite *ManifestSuite) SetupTest() {
	suite.ctx, suite.ctxCancel = context.WithTimeout(context.Background(), 3*time.Minute)

	suite.state = state.WrapCore(namespaced.NewState(inmem.Build))

	var err error

	suite.runtime, err = runtime.NewRuntime(suite.state, logging.Wrap(log.Writer()))
	suite.Require().NoError(err)

	suite.Require().NoError(suite.runtime.RegisterController(&k8sctrl.ManifestController{}))

	suite.startRuntime()
}

func (suite *ManifestSuite) startRuntime() {
	suite.wg.Add(1)

	go func() {
		defer suite.wg.Done()

		suite.Assert().NoError(suite.runtime.Run(suite.ctx))
	}()
}

//nolint:dupl
func (suite *ManifestSuite) assertManifests(manifests []string) error {
	resources, err := suite.state.List(suite.ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.ManifestType, "", resource.VersionUndefined))
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(resources.Items))

	for _, res := range resources.Items {
		ids = append(ids, res.Metadata().ID())
	}

	if !reflect.DeepEqual(manifests, ids) {
		return retry.ExpectedError(fmt.Errorf("expected %q, got %q", manifests, ids))
	}

	return nil
}

var defaultManifestSpec = config.K8sManifestsSpec{
	Server:        "127.0.0.1",
	ClusterDomain: "cluster.",

	PodCIDRs:     constants.DefaultIPv4PodNet,
	FirstPodCIDR: constants.DefaultIPv4PodNet,

	ProxyEnabled: true,
	ProxyImage:   "foo/bar",

	CoreDNSEnabled:  true,
	CoreDNSImage:    "foo/bar",
	CoreDNSReplicas: 1,

	DNSServiceIP: "192.168.0.1",

	FlannelEnabled:  true,
	FlannelImage:    "foo/bar",
	FlannelCNIImage: "foo/bar",
}

func (suite *ManifestSuite) TestReconcileDefaults() {
	rootSecrets := secrets.NewRoot(secrets.RootKubernetesID)
	manifestConfig := config.NewK8sManifests()
	manifestConfig.SetManifests(defaultManifestSpec)

	suite.Require().NoError(suite.state.Create(suite.ctx, rootSecrets))
	suite.Require().NoError(suite.state.Create(suite.ctx, manifestConfig))

	suite.Assert().NoError(retry.Constant(10*time.Second, retry.WithUnits(100*time.Millisecond)).Retry(
		func() error {
			return suite.assertManifests(
				[]string{
					"00-kubelet-bootstrapping-token",
					"01-csr-approver-role-binding",
					"01-csr-node-bootstrap",
					"01-csr-renewal-role-binding",
					"02-kube-system-sa-role-binding", "03-default-pod-security-policy", "05-flannel",
					"10-kube-proxy",
					"11-core-dns",
					"11-core-dns-svc",
					"11-kube-config-in-cluster",
				},
			)
		},
	))
}

func (suite *ManifestSuite) TestReconcileDisableKubeProxy() {
	rootSecrets := secrets.NewRoot(secrets.RootKubernetesID)
	manifestConfig := config.NewK8sManifests()
	spec := defaultManifestSpec
	spec.ProxyEnabled = false
	manifestConfig.SetManifests(spec)

	suite.Require().NoError(suite.state.Create(suite.ctx, rootSecrets))
	suite.Require().NoError(suite.state.Create(suite.ctx, manifestConfig))

	suite.Assert().NoError(retry.Constant(10*time.Second, retry.WithUnits(100*time.Millisecond)).Retry(
		func() error {
			return suite.assertManifests(
				[]string{
					"00-kubelet-bootstrapping-token",
					"01-csr-approver-role-binding",
					"01-csr-node-bootstrap",
					"01-csr-renewal-role-binding",
					"02-kube-system-sa-role-binding", "03-default-pod-security-policy", "05-flannel",
					"11-core-dns",
					"11-core-dns-svc",
					"11-kube-config-in-cluster",
				},
			)
		},
	))
}

func (suite *ManifestSuite) TestReconcileKubeProxyExtraArgs() {
	rootSecrets := secrets.NewRoot(secrets.RootKubernetesID)
	manifestConfig := config.NewK8sManifests()
	spec := defaultManifestSpec
	spec.ProxyExtraArgs = map[string]string{
		"bind-address": "\"::\"",
	}
	manifestConfig.SetManifests(spec)

	suite.Require().NoError(suite.state.Create(suite.ctx, rootSecrets))
	suite.Require().NoError(suite.state.Create(suite.ctx, manifestConfig))

	suite.Assert().NoError(retry.Constant(10*time.Second, retry.WithUnits(100*time.Millisecond)).Retry(
		func() error {
			return suite.assertManifests(
				[]string{
					"00-kubelet-bootstrapping-token",
					"01-csr-approver-role-binding",
					"01-csr-node-bootstrap",
					"01-csr-renewal-role-binding",
					"02-kube-system-sa-role-binding", "03-default-pod-security-policy", "05-flannel",
					"10-kube-proxy",
					"11-core-dns",
					"11-core-dns-svc",
					"11-kube-config-in-cluster",
				},
			)
		},
	))

	r, err := suite.state.Get(suite.ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.ManifestType, "10-kube-proxy", resource.VersionUndefined))
	suite.Require().NoError(err)

	manifest := r.(*k8s.Manifest) //nolint:errcheck,forcetypeassert
	suite.Assert().Len(manifest.Objects(), 3)

	suite.Assert().Equal("DaemonSet", manifest.Objects()[0].GetKind())

	ds := manifest.Objects()[0].Object
	containerSpec := ds["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0]
	args := containerSpec.(map[string]interface{})["command"].([]interface{}) //nolint:errcheck,forcetypeassert

	suite.Assert().Equal("--bind-address=\"::\"", args[len(args)-1])
}

func (suite *ManifestSuite) TearDownTest() {
	suite.T().Log("tear down")

	suite.ctxCancel()

	suite.wg.Wait()

	// trigger updates in resources to stop watch loops
	suite.Assert().NoError(suite.state.Create(context.Background(), secrets.NewRoot(secrets.RootEtcdID)))
	suite.Assert().NoError(suite.state.Create(context.Background(), config.NewK8sControlPlaneAPIServer()))
}

func TestManifestSuite(t *testing.T) {
	suite.Run(t, new(ManifestSuite))
}
