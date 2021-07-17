// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1alpha1

import (
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime"
	machineapi "github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
)

// Sequencer implements the sequencer interface.
type Sequencer struct{}

// NewSequencer intializes and returns a sequencer.
func NewSequencer() *Sequencer {
	return &Sequencer{}
}

// PhaseList represents a list of phases.
type PhaseList []runtime.Phase

// Append appends a task to the phase list.
func (p PhaseList) Append(name string, tasks ...runtime.TaskSetupFunc) PhaseList {
	p = append(p, runtime.Phase{
		Name:  name,
		Tasks: tasks,
	})

	return p
}

// AppendWhen appends a task to the phase list when `when` is `true`.
func (p PhaseList) AppendWhen(when bool, name string, tasks ...runtime.TaskSetupFunc) PhaseList {
	if when {
		p = p.Append(name, tasks...)
	}

	return p
}

// AppendList appends an additional PhaseList to the existing one.
func (p PhaseList) AppendList(list PhaseList) PhaseList {
	return append(p, list...)
}

// ApplyConfiguration defines a sequence which applies a new machine configuration to the node, rebooting to make it active.
func (*Sequencer) ApplyConfiguration(r runtime.Runtime, req *machineapi.ApplyConfigurationRequest) []runtime.Phase {
	phases := PhaseList{}

	phases = phases.Append(
		"saveStateEncryptionConfig",
		SaveStateEncryptionConfig,
	).Append(
		"mountState",
		MountStatePartition,
	).Append(
		"saveConfig",
		SaveConfig,
	).Append(
		"unmountState",
		UnmountStatePartition,
	).Append(
		"cleanup",
		StopAllPods,
	).AppendList(
		stopAllPhaselist(r),
	).Append(
		"reboot",
		Reboot,
	)

	return phases
}

// Initialize is the initialize sequence. The primary goals of this sequence is
// to load the config and enforce kernel security requirements.
func (*Sequencer) Initialize(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		phases = phases.Append(
			"logger",
			SetupLogger,
		).Append(
			"systemRequirements",
			WriteRequiredSysctlsForContainer,
			SetupSystemDirectory,
		).Append(
			"etc",
			CreateOSReleaseFile,
		).Append(
			"config",
			LoadConfig,
		)
	default:
		phases = phases.Append(
			"logger",
			SetupLogger,
		).Append(
			"systemRequirements",
			EnforceKSPPRequirements,
			WriteRequiredSysctls,
			SetupSystemDirectory,
			MountBPFFS,
			MountCgroups,
			MountPseudoFilesystems,
			SetRLimit,
		).Append(
			"integrity",
			WriteIMAPolicy,
			SetupSystemCgroups,
		).Append(
			"etc",
			CreateOSReleaseFile,
		).AppendWhen(
			r.State().Machine().Installed(),
			"mountSystem",
			MountStatePartition,
		).Append(
			"config",
			LoadConfig,
		).AppendWhen(
			r.State().Machine().Installed(),
			"unmountSystem",
			UnmountStatePartition,
		)
	}

	return phases
}

// Install is the install sequence.
func (*Sequencer) Install(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		return nil
	default:
		if !r.State().Machine().Installed() || r.State().Machine().IsInstallStaged() {
			phases = phases.Append(
				"validateConfig",
				ValidateConfig,
			).Append(
				"env",
				SetUserEnvVars,
			).Append(
				"containerd",
				StartContainerd,
			).Append(
				"install",
				Install,
			).Append(
				"saveStateEncryptionConfig",
				SaveStateEncryptionConfig,
			).Append(
				"mountState",
				MountStatePartition,
			).Append(
				"saveConfig",
				SaveConfig,
			).Append(
				"unmountState",
				UnmountStatePartition,
			).Append(
				"stopEverything",
				StopAllServices,
			).Append(
				"reboot",
				Reboot,
			)
		}
	}

	return phases
}

// Boot is the boot sequence. This primary goal if this sequence is to apply
// user supplied settings and start the services for the specific machine type.
// This sequence should never be reached if an installation is not found.
func (*Sequencer) Boot(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}

	phases = phases.AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"saveStateEncryptionConfig",
		SaveStateEncryptionConfig,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"mountState",
		MountStatePartition,
	).Append(
		"validateConfig",
		ValidateConfig,
	).Append(
		"saveConfig",
		SaveConfig,
	).Append(
		"env",
		SetUserEnvVars,
	).Append(
		"containerd",
		StartContainerd,
	).AppendWhen(
		r.State().Platform().Mode() == runtime.ModeContainer,
		"sharedFilesystems",
		SetupSharedFilesystems,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"ephemeral",
		MountEphemeralPartition,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"verifyInstall",
		VerifyInstallation,
	).Append(
		"var",
		SetupVarDirectory,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"overlay",
		MountOverlayFilesystems,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"udevd",
		StartUdevd,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"userDisks",
		MountUserDisks,
	).Append(
		"userSetup",
		WriteUserFiles,
		WriteUserSysctls,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"lvm",
		ActivateLogicalVolumes,
	).Append(
		"startEverything",
		StartAllServices,
	).AppendWhen(
		r.Config().Machine().Type() != machine.TypeWorker,
		"labelMaster",
		LabelNodeAsMaster,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"uncordon",
		UncordonNode,
	).AppendWhen(
		r.State().Platform().Mode() != runtime.ModeContainer,
		"bootloader",
		UpdateBootloader,
	).AppendWhen(
		r.Config().Machine().Type() != machine.TypeWorker,
		"checkControlPlaneStatus",
		CheckControlPlaneStatus,
	)

	return phases
}

// Bootstrap is the bootstrap sequence. This primary goal if this sequence is
// to bootstrap Etcd and Kubernetes.
func (*Sequencer) Bootstrap(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}

	phases = phases.Append(
		"etcd",
		BootstrapEtcd,
	)

	return phases
}

// Reboot is the reboot sequence.
func (*Sequencer) Reboot(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}.Append(
		"cleanup",
		StopAllPods,
	).
		AppendList(stopAllPhaselist(r)).
		Append("reboot", Reboot)

	return phases
}

// Reset is the reset sequence.
func (*Sequencer) Reset(r runtime.Runtime, in runtime.ResetOptions) []runtime.Phase {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		phases = phases.AppendList(stopAllPhaselist(r)).
			Append(
				"shutdown",
				Shutdown,
			)
	default:
		phases = phases.AppendWhen(
			in.GetGraceful(),
			"drain",
			CordonAndDrainNode,
		).AppendWhen(
			in.GetGraceful(),
			"cleanup",
			RemoveAllPods,
		).AppendWhen(
			!in.GetGraceful(),
			"cleanup",
			StopAllPods,
		).AppendWhen(
			in.GetGraceful() && (r.Config().Machine().Type() != machine.TypeWorker),
			"leave",
			LeaveEtcd,
		).AppendList(
			stopAllPhaselist(r),
		).AppendWhen(
			len(in.GetSystemDiskTargets()) == 0,
			"reset",
			ResetSystemDisk,
		).AppendWhen(
			len(in.GetSystemDiskTargets()) > 0,
			"resetSpec",
			ResetSystemDiskSpec,
		).AppendWhen(
			in.GetReboot(),
			"reboot",
			Reboot,
		).AppendWhen(
			!in.GetReboot(),
			"shutdown",
			Shutdown,
		)
	}

	return phases
}

// Shutdown is the shutdown sequence.
func (*Sequencer) Shutdown(r runtime.Runtime) []runtime.Phase {
	phases := PhaseList{}.
		Append(
			"cleanup",
			StopAllPods,
		).
		AppendList(stopAllPhaselist(r)).
		Append("shutdown", Shutdown)

	return phases
}

// StageUpgrade is the stage upgrade sequence.
func (*Sequencer) StageUpgrade(r runtime.Runtime, in *machineapi.UpgradeRequest) []runtime.Phase {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		return nil
	default:
		phases = phases.Append(
			"cleanup",
			StopAllPods,
		).AppendWhen(
			!in.GetPreserve() && (r.Config().Machine().Type() != machine.TypeWorker),
			"leave",
			LeaveEtcd,
		).AppendList(
			stopAllPhaselist(r),
		).Append(
			"reboot",
			Reboot,
		)
	}

	return phases
}

// Upgrade is the upgrade sequence.
func (*Sequencer) Upgrade(r runtime.Runtime, in *machineapi.UpgradeRequest) []runtime.Phase {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		return nil
	default:
		phases = phases.Append(
			"drain",
			CordonAndDrainNode,
		).AppendWhen(
			!in.GetPreserve(),
			"cleanup",
			RemoveAllPods,
		).AppendWhen(
			in.GetPreserve(),
			"cleanup",
			StopAllPods,
		).AppendWhen(
			!in.GetPreserve() && (r.Config().Machine().Type() != machine.TypeWorker),
			"leave",
			LeaveEtcd,
		).Append(
			"stopServices",
			StopServicesForUpgrade,
		).Append(
			"unmountUser",
			UnmountUserDisks,
		).Append(
			"unmount",
			UnmountOverlayFilesystems,
			UnmountPodMounts,
		).Append(
			"unmountBind",
			UnmountSystemDiskBindMounts,
		).Append(
			"unmountSystem",
			UnmountEphemeralPartition,
			UnmountStatePartition,
		).Append(
			"verifyDisk",
			VerifyDiskAvailability,
		).Append(
			"upgrade",
			Upgrade,
		).Append(
			"stopEverything",
			StopAllServices,
		).Append(
			"reboot",
			Reboot,
		)
	}

	return phases
}

func stopAllPhaselist(r runtime.Runtime) PhaseList {
	phases := PhaseList{}

	switch r.State().Platform().Mode() { //nolint:exhaustive
	case runtime.ModeContainer:
		phases = phases.Append(
			"stopEverything",
			StopAllServices,
		)
	default:
		phases = phases.Append(
			"stopEverything",
			StopAllServices,
		).Append(
			"unmountUser",
			UnmountUserDisks,
		).Append(
			"umount",
			UnmountOverlayFilesystems,
			UnmountPodMounts,
		).Append(
			"unmountBind",
			UnmountSystemDiskBindMounts,
		).Append(
			"unmountSystem",
			UnmountEphemeralPartition,
			UnmountStatePartition,
		)
	}

	return phases
}
