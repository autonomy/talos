// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package kubernetes

import (
	"os"

	"github.com/talos-systems/talos/internal/app/machined/internal/phase"
	"github.com/talos-systems/talos/internal/pkg/runtime"
	"github.com/talos-systems/talos/pkg/kubernetes"
)

// CordonAndDrain represents the task for stop all containerd tasks in the
// k8s.io namespace.
type CordonAndDrain struct{}

// NewCordonAndDrainTask initializes and returns an Services task.
func NewCordonAndDrainTask() phase.Task {
	return &CordonAndDrain{}
}

// TaskFunc returns the runtime function.
func (task *CordonAndDrain) TaskFunc(mode runtime.Mode) phase.TaskFunc {
	return func(r runtime.Runtime) error {
		return task.standard()
	}
}

func (task *CordonAndDrain) standard() (err error) {
	var hostname string

	if hostname, err = os.Hostname(); err != nil {
		return err
	}

	var kubeHelper *kubernetes.Client

	if kubeHelper, err = kubernetes.NewClientFromKubeletKubeconfig(); err != nil {
		return err
	}

	if err = kubeHelper.CordonAndDrain(hostname); err != nil {
		return err
	}

	return nil
}
