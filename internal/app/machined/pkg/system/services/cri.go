// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package services

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/defaults"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/events"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/health"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/runner"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/runner/process"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/runner/restart"
	"github.com/talos-systems/talos/internal/pkg/conditions"
	"github.com/talos-systems/talos/pkg/universe"
)

// CRI implements the Service interface. It serves as the concrete type with
// the required methods.
type CRI struct{}

// ID implements the Service interface.
func (c *CRI) ID(config runtime.Configurator) string {
	return "cri"
}

// PreFunc implements the Service interface.
func (c *CRI) PreFunc(ctx context.Context, config runtime.Configurator) error {
	return os.MkdirAll(defaults.DefaultRootDir, os.ModeDir)
}

// PostFunc implements the Service interface.
func (c *CRI) PostFunc(config runtime.Configurator, state events.ServiceState) (err error) {
	return nil
}

// Condition implements the Service interface.
func (c *CRI) Condition(config runtime.Configurator) conditions.Condition {
	return nil
}

// DependsOn implements the Service interface.
func (c *CRI) DependsOn(config runtime.Configurator) []string {
	return []string{"networkd"}
}

// Runner implements the Service interface.
func (c *CRI) Runner(config runtime.Configurator) (runner.Runner, error) {
	// Set the process arguments.
	args := &runner.Args{
		ID: c.ID(config),
		ProcessArgs: []string{
			"/bin/containerd",
			"--address",
			universe.ContainerdAddress,
			"--config",
			universe.CRIContainerdConfig,
		},
	}

	env := []string{}
	for key, val := range config.Machine().Env() {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}

	return restart.New(process.NewRunner(
		config.Debug(),
		args,
		runner.WithEnv(env),
	),
		restart.WithType(restart.Forever),
	), nil
}

// HealthFunc implements the HealthcheckedService interface
func (c *CRI) HealthFunc(runtime.Configurator) health.Check {
	return func(ctx context.Context) error {
		client, err := containerd.New(universe.ContainerdAddress)
		if err != nil {
			return err
		}
		// nolint: errcheck
		defer client.Close()

		resp, err := client.HealthService().Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return err
		}

		if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return fmt.Errorf("unexpected serving status: %d", resp.Status)
		}

		return nil
	}
}

// HealthSettings implements the HealthcheckedService interface
func (c *CRI) HealthSettings(runtime.Configurator) *health.Settings {
	return &health.DefaultSettings
}
