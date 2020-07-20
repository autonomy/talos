// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build integration_cli

package base

import (
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime"
	"github.com/talos-systems/talos/internal/pkg/cluster"
	"github.com/talos-systems/talos/pkg/retry"
)

// CLISuite is a base suite for CLI tests.
type CLISuite struct {
	suite.Suite
	TalosSuite
}

// DiscoverNodes provides list of Talos nodes in the cluster.
//
// As there's no way to provide this functionality via Talos CLI, it relies on cluster info.
func (cliSuite *CLISuite) DiscoverNodes() cluster.Info {
	discoveredNodes := cliSuite.TalosSuite.DiscoverNodes()
	if discoveredNodes != nil {
		return discoveredNodes
	}

	// still no nodes, skip the test
	cliSuite.T().Skip("no nodes were discovered")

	return nil
}

// RandomNode returns a random node of the specified type (or any type if no types are specified).
func (cliSuite *CLISuite) RandomDiscoveredNode(types ...runtime.MachineType) string {
	nodeInfo := cliSuite.DiscoverNodes()

	var nodes []string

	if len(types) == 0 {
		nodes = nodeInfo.Nodes()
	} else {
		for _, t := range types {
			nodes = append(nodes, nodeInfo.NodesByType(t)...)
		}
	}

	cliSuite.Require().NotEmpty(nodes)

	return nodes[rand.Intn(len(nodes))]
}

func (cliSuite *CLISuite) buildCLICmd(args []string) *exec.Cmd {
	// TODO: add support for calling `talosctl config endpoint` before running talosctl

	args = append([]string{"--talosconfig", cliSuite.TalosConfig}, args...)

	return exec.Command(cliSuite.TalosctlPath, args...)
}

// RunCLI runs talosctl binary with the options provided.
func (cliSuite *CLISuite) RunCLI(args []string, options ...RunOption) {
	Run(&cliSuite.Suite, cliSuite.buildCLICmd(args), options...)
}

func (cliSuite *CLISuite) RunAndWaitForMatch(args []string, regex *regexp.Regexp, duration time.Duration, options ...retry.Option) {
	cliSuite.Assert().NoError(retry.Constant(duration, options...).Retry(func() error {
		stdout, _, err := RunAndWait(&cliSuite.Suite, cliSuite.buildCLICmd(args))
		if err != nil {
			return retry.UnexpectedError(err)
		}

		if !regex.MatchString(stdout.String()) {
			return retry.ExpectedError(fmt.Errorf("stdout doesn't match: %q", stdout))
		}

		return nil
	}))
}
