// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1alpha1

import (
	"github.com/talos-systems/talos/api/machine"
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime"
)

// Sequencer is an implementation of a `Sequencer`.
type Sequencer struct{}

// Initialize is the initialize sequence.
func (*Sequencer) Initialize(r runtime.Runtime) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return []runtime.Phase{
			{
				WriteRequiredSysctlsForContainer,
				SetupSystemDirectory,
			},
			{
				CreatOSReleaseFile,
			},
			{
				SaveConfig,
			},
		}
	default:
		return []runtime.Phase{
			{
				EnforceKSPPRequirements,
				WriteRequiredSysctls,
				SetupSystemDirectory,
				MountBPFFS,
				MountCgroups,
				MountSubDevices,
				SetFileLimit,
			},
			{
				WriteIMAPolicy,
			},
			{
				CreateEtcNetworkFiles,
				CreatOSReleaseFile,
			},
			{
				SetupDiscoveryNetwork,
			},
			{
				MountBootPartition,
			},
			{
				SaveConfig,
			},
			{
				ResetNetwork,
			},
			{
				SetupDiscoveryNetwork,
			},
		}
	}
}

// Boot is the boot sequence.
func (*Sequencer) Boot(r runtime.Runtime) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return []runtime.Phase{
			{
				ValidateConfig,
			},
			{
				SetUserEnvVars,
			},
			{
				StartStage1SystemServices,
			},
			{
				InitializePlatform,
			},
			{
				VerifyInstallation,
			},
			{
				MountVolumesAsSharedForContainer,
				SetupVarDirectory,
			},
			{
				WriteUserFiles,
				WriteUserSysctls,
			},
			{
				StartStage2SystemServices,
				StartOrchestrationServices,
			},
			{
				LabelNodeAsMaster,
			},
		}
	default:
		return []runtime.Phase{
			{
				ValidateConfig,
			},
			{
				SetUserEnvVars,
			},
			{
				StartStage1SystemServices,
			},
			{
				InitializePlatform,
			},
			{
				VerifyInstallation,
			},
			{
				MountOverlayFilesystems,
				SetupVarDirectory,
			},
			{
				MountUserDisks,
			},
			{
				WriteUserFiles,
				WriteUserSysctls,
			},
			{
				StartStage2SystemServices,
				StartOrchestrationServices,
			},
			{
				LabelNodeAsMaster,
			},
			{
				UpdateBootloader,
			},
		}
	}
}

// Reboot is the reboot sequence.
func (*Sequencer) Reboot(r runtime.Runtime) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return []runtime.Phase{
			{
				StopServices,
			},
			{
				Reboot,
			},
		}
	default:
		return []runtime.Phase{
			{
				StopServices,
			},
			{
				UnmountOverlayFilesystems,
				UnmountPodMounts,
			},
			{
				UnmountSystemDisks,
			},
			{
				UnmountSystemDiskBindMounts,
			},
			{
				Reboot,
			},
		}
	}
}

// Reset is the reset sequence.
func (*Sequencer) Reset(r runtime.Runtime, in *machine.ResetRequest) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return []runtime.Phase{
			{
				StopServices,
			},
			{
				Shutdown,
			},
		}
	default:
		if in.GetGraceful() {
			return []runtime.Phase{
				{
					CordonAndDrainNode,
				},
				{
					LeaveEtcd,
				},
				{
					RemoveAllPods,
				},
				{
					StopServices,
				},
				{
					UnmountOverlayFilesystems,
					UnmountPodMounts,
				},
				{
					UnmountSystemDisks,
				},
				{
					UnmountSystemDiskBindMounts,
				},
				{
					ResetSystemDisk,
				},
			}
		}

		return []runtime.Phase{
			{
				StopServices,
			},
			{
				UnmountOverlayFilesystems,
				UnmountPodMounts,
			},
			{
				UnmountSystemDisks,
			},
			{
				UnmountSystemDiskBindMounts,
			},
			{
				ResetSystemDisk,
			},
		}
	}
}

// Shutdown is the shutdown sequence.
func (*Sequencer) Shutdown(r runtime.Runtime) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return []runtime.Phase{
			{
				StopServices,
			},
			{
				Shutdown,
			},
		}
	default:
		return []runtime.Phase{
			{
				StopServices,
			},
			{
				UnmountOverlayFilesystems,
				UnmountPodMounts,
			},
			{
				UnmountSystemDisks,
			},
			{
				UnmountSystemDiskBindMounts,
			},
			{
				Shutdown,
			},
		}
	}
}

// Upgrade is the upgrade sequence.
func (*Sequencer) Upgrade(r runtime.Runtime, in *machine.UpgradeRequest) []runtime.Phase {
	switch r.Platform().Mode() {
	case runtime.Container:
		return nil
	default:
		return []runtime.Phase{
			{
				CordonAndDrainNode,
			},
			{
				LeaveEtcd,
			},
			{
				RemoveAllPods,
			},
			{
				StopServices,
			},
			{
				UnmountOverlayFilesystems,
				UnmountPodMounts,
			},
			{
				UnmountSystemDisks,
			},
			{
				VerifyDiskAvailability,
			},
			{
				Upgrade,
			},
			{
				StopServices,
			},
			{
				Reboot,
			},
		}
	}
}
