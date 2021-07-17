// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1alpha1

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/talos-systems/go-blockdevice/blockdevice"
	"github.com/talos-systems/go-blockdevice/blockdevice/partition/gpt"
	"github.com/talos-systems/go-blockdevice/blockdevice/util"
	"github.com/talos-systems/go-cmd/pkg/cmd"
	"github.com/talos-systems/go-kmsg"
	"github.com/talos-systems/go-procfs/procfs"
	"github.com/talos-systems/go-retry/retry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sys/unix"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	installer "github.com/talos-systems/talos/cmd/installer/pkg/install"
	"github.com/talos-systems/talos/internal/app/machined/internal/install"
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime"
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime/v1alpha1/bootloader"
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime/v1alpha1/bootloader/adv"
	"github.com/talos-systems/talos/internal/app/machined/pkg/runtime/v1alpha1/bootloader/grub"
	perrors "github.com/talos-systems/talos/internal/app/machined/pkg/runtime/v1alpha1/platform/errors"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/events"
	"github.com/talos-systems/talos/internal/app/machined/pkg/system/services"
	"github.com/talos-systems/talos/internal/app/maintenance"
	"github.com/talos-systems/talos/internal/pkg/containers/cri/containerd"
	"github.com/talos-systems/talos/internal/pkg/cri"
	"github.com/talos-systems/talos/internal/pkg/etcd"
	"github.com/talos-systems/talos/internal/pkg/kernel/kspp"
	"github.com/talos-systems/talos/internal/pkg/mount"
	"github.com/talos-systems/talos/internal/pkg/partition"
	"github.com/talos-systems/talos/pkg/conditions"
	"github.com/talos-systems/talos/pkg/images"
	"github.com/talos-systems/talos/pkg/kubernetes"
	machineapi "github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/configloader"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	resourcev1alpha1 "github.com/talos-systems/talos/pkg/resources/v1alpha1"
	"github.com/talos-systems/talos/pkg/sysctl"
	"github.com/talos-systems/talos/pkg/version"
)

// SetupLogger represents the SetupLogger task.
func SetupLogger(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		machinedLog, err := r.Logging().ServiceLog("machined").Writer()
		if err != nil {
			return err
		}

		if r.State().Platform().Mode() == runtime.ModeContainer {
			// send all the logs to machinedLog as well, but skip /dev/kmsg logging
			log.SetOutput(io.MultiWriter(log.Writer(), machinedLog))
			log.SetPrefix("[talos] ")

			return nil
		}

		// disable ratelimiting for kmsg, otherwise logs might be not visible.
		// this should be set via kernel arg, but in case it's not set, try to force it.
		if err = sysctl.WriteSystemProperty(&sysctl.SystemProperty{
			Key:   "kernel.printk_devkmsg",
			Value: "on\n",
		}); err != nil {
			var serr syscall.Errno

			if !(errors.As(err, &serr) && serr == syscall.EINVAL) { // ignore EINVAL which is returned when kernel arg is set
				log.Printf("failed setting kernel.printk_devkmsg: %s, error ignored", err)
			}
		}

		if err = kmsg.SetupLogger(nil, "[talos]", machinedLog); err != nil {
			return fmt.Errorf("failed to setup logging: %w", err)
		}

		return nil
	}, "setupLogger"
}

// EnforceKSPPRequirements represents the EnforceKSPPRequirements task.
func EnforceKSPPRequirements(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if err = kspp.EnforceKSPPKernelParameters(); err != nil {
			return err
		}

		return kspp.EnforceKSPPSysctls()
	}, "enforceKSPPRequirements"
}

// SetupSystemDirectory represents the SetupSystemDirectory task.
func SetupSystemDirectory(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		for _, p := range []string{constants.SystemEtcPath, constants.SystemRunPath, constants.SystemVarPath, constants.StateMountPoint} {
			if err = os.MkdirAll(p, 0o700); err != nil {
				return err
			}
		}

		return nil
	}, "setupSystemDirectory"
}

// SetupSystemCgroups represents the SetupSystemCgroups task.
func SetupSystemCgroups(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {

		initPid := []byte(strconv.Itoa(1))
		groups := []string{
			"init",
			"kubelet",
		}

		for _, c := range groups {
			p := path.Join("/sys/fs/cgroup/", c)
			if err = os.MkdirAll(p, 0o755); err != nil {
				return fmt.Errorf("failed to create default cgroup: %w", err)
			}
		}

		if err = ioutil.WriteFile("/sys/fs/cgroup/init/cgroup.procs", initPid, os.FileMode(0o644)); err != nil {
			return fmt.Errorf("failed to move init process to cgroup: %w", err)
		}

		return nil
	}, "SetupSystemCgroups"
}

// MountBPFFS represents the MountBPFFS task.
func MountBPFFS(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var mountpoints *mount.Points

		mountpoints, err = mount.BPFMountPoints()
		if err != nil {
			return err
		}

		return mount.Mount(mountpoints)
	}, "mountBPFFS"
}

const (
	memoryCgroup                  = "memory"
	memoryUseHierarchy            = "memory.use_hierarchy"
	memoryUseHierarchyPermissions = os.FileMode(0o400)
)

var memoryUseHierarchyContents = []byte(strconv.Itoa(1))

// MountCgroups represents the MountCgroups task.
func MountCgroups(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var mountpoints *mount.Points

		mountpoints, err = mount.CGroupMountPoints()
		if err != nil {
			return err
		}

		if err = mount.Mount(mountpoints); err != nil {
			return err
		}

		return nil
	}, "mountCgroups"
}

// MountPseudoFilesystems represents the MountPseudoFilesystems task.
func MountPseudoFilesystems(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var mountpoints *mount.Points

		mountpoints, err = mount.PseudoSubMountPoints()
		if err != nil {
			return err
		}

		return mount.Mount(mountpoints)
	}, "mountPseudoFilesystems"
}

// WriteRequiredSysctlsForContainer represents the WriteRequiredSysctlsForContainer task.
func WriteRequiredSysctlsForContainer(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var multiErr *multierror.Error

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.ipv4.ip_forward", Value: "1"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.ipv4.ip_forward: %w", err))
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.ipv6.conf.default.forwarding", Value: "1"}); err != nil {
			if !errors.Is(err, os.ErrNotExist) { // ignore error if ipv6 is disabled
				multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.ipv6.conf.default.forwarding: %w", err))
			}
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "kernel.pid_max", Value: "262144"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set kernel.pid_max: %w", err))
		}

		return multiErr.ErrorOrNil()
	}, "writeRequiredSysctlsForContainer"
}

// WriteRequiredSysctls represents the WriteRequiredSysctls task.
func WriteRequiredSysctls(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var multiErr *multierror.Error

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.ipv4.ip_forward", Value: "1"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.ipv4.ip_forward: %w", err))
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.bridge.bridge-nf-call-iptables", Value: "1"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.bridge.bridge-nf-call-iptables: %w", err))
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.bridge.bridge-nf-call-ip6tables", Value: "1"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.bridge.bridge-nf-call-ip6tables: %w", err))
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "net.ipv6.conf.default.forwarding", Value: "1"}); err != nil {
			if !errors.Is(err, os.ErrNotExist) { // ignore error if ipv6 is disabled
				multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set net.ipv6.conf.default.forwarding: %w", err))
			}
		}

		if err := sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: "kernel.pid_max", Value: "262144"}); err != nil {
			multiErr = multierror.Append(multiErr, fmt.Errorf("failed to set kernel.pid_max: %w", err))
		}

		return multiErr.ErrorOrNil()
	}, "writeRequiredSysctls"
}

// SetRLimit represents the SetRLimit task.
func SetRLimit(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		// TODO(andrewrynhard): Should we read limit from /proc/sys/fs/nr_open?
		return unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: 1048576, Max: 1048576})
	}, "setRLimit"
}

// See https://www.kernel.org/doc/Documentation/ABI/testing/ima_policy
var rules = []string{
	"dont_measure fsmagic=0x9fa0",     // PROC_SUPER_MAGIC
	"dont_measure fsmagic=0x62656572", // SYSFS_MAGIC
	"dont_measure fsmagic=0x64626720", // DEBUGFS_MAGIC
	"dont_measure fsmagic=0x1021994",  // TMPFS_MAGIC
	"dont_measure fsmagic=0x1cd1",     // DEVPTS_SUPER_MAGIC
	"dont_measure fsmagic=0x42494e4d", // BINFMTFS_MAGIC
	"dont_measure fsmagic=0x73636673", // SECURITYFS_MAGIC
	"dont_measure fsmagic=0xf97cff8c", // SELINUX_MAGIC
	"dont_measure fsmagic=0x43415d53", // SMACK_MAGIC
	"dont_measure fsmagic=0x27e0eb",   // CGROUP_SUPER_MAGIC
	"dont_measure fsmagic=0x63677270", // CGROUP2_SUPER_MAGIC
	"dont_measure fsmagic=0x6e736673", // NSFS_MAGIC
	"dont_measure fsmagic=0xde5e81e4", // EFIVARFS_MAGIC
	"dont_measure fsmagic=0x58465342", // XFS_MAGIC
	"dont_measure fsmagic=0x794c7630", // OVERLAYFS_SUPER_MAGIC
	"measure func=MMAP_CHECK mask=MAY_EXEC",
	"measure func=BPRM_CHECK mask=MAY_EXEC",
	"measure func=FILE_CHECK mask=^MAY_READ euid=0",
	"measure func=FILE_CHECK mask=^MAY_READ uid=0",
	"measure func=MODULE_CHECK",
	"measure func=FIRMWARE_CHECK",
	"measure func=POLICY_CHECK",
}

// WriteIMAPolicy represents the WriteIMAPolicy task.
func WriteIMAPolicy(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if _, err = os.Stat("/sys/kernel/security/ima/policy"); os.IsNotExist(err) {
			return fmt.Errorf("policy file does not exist: %w", err)
		}

		f, err := os.OpenFile("/sys/kernel/security/ima/policy", os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}

		defer f.Close() //nolint:errcheck

		for _, line := range rules {
			if _, err = f.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("rule %q is invalid", err)
			}
		}

		return nil
	}, "writeIMAPolicy"
}

const osReleaseTemplate = `
NAME="{{ .Name }}"
ID={{ .ID }}
VERSION_ID={{ .Version }}
PRETTY_NAME="{{ .Name }} ({{ .Version }})"
HOME_URL="https://docs.talos-systems.com/"
BUG_REPORT_URL="https://github.com/talos-systems/talos/issues"
`

// OSRelease renders a valid /etc/os-release file and writes it to disk. The
// node's OS Image field is reported by the node from /etc/os-release.
func OSRelease() (err error) {
	if err = createBindMount(filepath.Join(constants.SystemEtcPath, "os-release"), "/etc/os-release"); err != nil {
		return err
	}

	var (
		v    string
		tmpl *template.Template
	)

	switch version.Tag {
	case "none":
		v = version.SHA
	default:
		v = version.Tag
	}

	data := struct {
		Name    string
		ID      string
		Version string
	}{
		Name:    version.Name,
		ID:      strings.ToLower(version.Name),
		Version: v,
	}

	tmpl, err = template.New("").Parse(osReleaseTemplate)
	if err != nil {
		return err
	}

	var buf []byte

	writer := bytes.NewBuffer(buf)

	err = tmpl.Execute(writer, data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(constants.SystemEtcPath, "os-release"), writer.Bytes(), 0o644)
}

// createBindMount creates a common way to create a writable source file with a
// bind mounted destination. This is most commonly used for well known files
// under /etc that need to be adjusted during startup.
func createBindMount(src, dst string) (err error) {
	var f *os.File

	if f, err = os.OpenFile(src, os.O_WRONLY|os.O_CREATE, 0o644); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = unix.Mount(src, dst, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to create bind mount for %s: %w", dst, err)
	}

	return nil
}

// CreateOSReleaseFile represents the CreateOSReleaseFile task.
func CreateOSReleaseFile(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		// Create /etc/os-release.
		return OSRelease()
	}, "createOSReleaseFile"
}

// LoadConfig represents the LoadConfig task.
func LoadConfig(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		download := func() error {
			var b []byte

			fetchCtx, ctxCancel := context.WithTimeout(context.Background(), 70*time.Second)
			defer ctxCancel()

			b, e := fetchConfig(fetchCtx, r)
			if errors.Is(e, perrors.ErrNoConfigSource) {
				logger.Println("starting maintenance service")

				b, e = receiveConfigViaMaintenanceService(ctx, logger, r)
				if e != nil {
					return fmt.Errorf("failed to receive config via maintenance service: %w", e)
				}
			}

			if e != nil {
				return e
			}

			logger.Printf("storing config in memory")

			return r.SetConfig(b)
		}

		cfg, err := configloader.NewFromFile(constants.ConfigPath)
		if err != nil {
			logger.Printf("downloading config")

			return download()
		}

		if !cfg.Persist() {
			logger.Printf("found existing config, but persistence is disabled, downloading config")

			return download()
		}

		logger.Printf("persistence is enabled, using existing config on disk")

		b, err := cfg.Bytes()
		if err != nil {
			return err
		}

		return r.SetConfig(b)
	}, "loadConfig"
}

// SaveConfig represents the SaveConfig task.
func SaveConfig(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if err = r.Config().ApplyDynamicConfig(ctx, r.State().Platform()); err != nil {
			return err
		}

		err = r.State().V1Alpha2().SetConfig(r.Config())
		if err != nil {
			return err
		}

		var b []byte

		b, err = r.Config().Bytes()
		if err != nil {
			return err
		}

		return ioutil.WriteFile(constants.ConfigPath, b, 0o600)
	}, "saveConfig"
}

func fetchConfig(ctx context.Context, r runtime.Runtime) (out []byte, err error) {
	var b []byte

	if b, err = r.State().Platform().Configuration(ctx); err != nil {
		return nil, err
	}

	// Detect if config is a gzip archive and unzip it if so
	contentType := http.DetectContentType(b)
	if contentType == "application/x-gzip" {
		var gzipReader *gzip.Reader

		gzipReader, err = gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %w", err)
		}

		//nolint:errcheck
		defer gzipReader.Close()

		var unzippedData []byte

		unzippedData, err = ioutil.ReadAll(gzipReader)
		if err != nil {
			return nil, fmt.Errorf("error unzipping machine config: %w", err)
		}

		b = unzippedData
	}

	return b, nil
}

func receiveConfigViaMaintenanceService(ctx context.Context, logger *log.Logger, r runtime.Runtime) ([]byte, error) {
	cfgBytes, err := maintenance.Run(ctx, logger, r)
	if err != nil {
		return nil, fmt.Errorf("maintenance service failed: %w", err)
	}

	provider, err := configloader.NewFromBytes(cfgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	warnings, err := provider.Validate(r.State().Platform().Mode())
	for _, w := range warnings {
		logger.Printf("WARNING:\n%s", w)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	processedBytes, err := provider.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to export validated config: %w", err)
	}

	return processedBytes, nil
}

// ValidateConfig validates the config.
func ValidateConfig(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) error {
		warnings, err := r.Config().Validate(r.State().Platform().Mode())
		for _, w := range warnings {
			logger.Printf("WARNING:\n%s", w)
		}

		return err
	}, "validateConfig"
}

// SetUserEnvVars represents the SetUserEnvVars task.
func SetUserEnvVars(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		for key, val := range r.Config().Machine().Env() {
			if err = os.Setenv(key, val); err != nil {
				return fmt.Errorf("failed to set enivronment variable: %w", err)
			}
		}

		return nil
	}, "setUserEnvVars"
}

// StartContainerd represents the task to start containerd.
func StartContainerd(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		svc := &services.Containerd{}

		system.Services(r).LoadAndStart(svc)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		return system.WaitForService(system.StateEventUp, svc.ID(r)).Wait(ctx)
	}, "startContainerd"
}

// StartUdevd represents the task to start udevd.
func StartUdevd(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		svc := &services.Udevd{}

		system.Services(r).LoadAndStart(svc)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		return system.WaitForService(system.StateEventUp, svc.ID(r)).Wait(ctx)
	}, "startUdevd"
}

// StartAllServices represents the task to start the system services.
func StartAllServices(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		svcs := system.Services(r)

		svcs.Load(
			&services.APID{},
			&services.CRI{},
			&services.Kubelet{},
		)

		switch t := r.Config().Machine().Type(); t {
		case machine.TypeInit:
			svcs.Load(
				&services.Trustd{},
				&services.Etcd{Bootstrap: true},
			)
		case machine.TypeControlPlane:
			svcs.Load(
				&services.Trustd{},
				&services.Etcd{},
			)
		case machine.TypeWorker:
			// nothing
		case machine.TypeUnknown:
			fallthrough
		default:
			panic(fmt.Sprintf("unexpected machine type %v", t))
		}

		system.Services(r).StartAll()

		all := []conditions.Condition{}

		logger.Printf("waiting for %d services", len(svcs.List()))

		for _, svc := range svcs.List() {
			cond := system.WaitForService(system.StateEventUp, svc.AsProto().GetId())
			all = append(all, cond)
		}

		ctx, cancel := context.WithTimeout(ctx, constants.BootTimeout)
		defer cancel()

		return conditions.WaitForAll(all...).Wait(ctx)
	}, "startAllServices"
}

// StopServicesForUpgrade represents the StopServicesForUpgrade task.
func StopServicesForUpgrade(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return system.Services(nil).StopWithRevDepenencies(ctx, "cri", "etcd", "kubelet", "udevd")
	}, "stopServicesForUpgrade"
}

// StopAllServices represents the StopAllServices task.
func StopAllServices(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		system.Services(nil).Shutdown(ctx)

		return nil
	}, "stopAllServices"
}

// VerifyInstallation represents the VerifyInstallation task.
func VerifyInstallation(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var (
			current string
			next    string
			disk    string
		)

		disk, err = r.Config().Machine().Install().Disk()
		if err != nil {
			return err
		}

		grub := &grub.Grub{
			BootDisk: disk,
		}

		current, next, err = grub.Labels()
		if err != nil {
			return err
		}

		if current == "" && next == "" {
			return fmt.Errorf("bootloader is not configured")
		}

		return err
	}, "verifyInstallation"
}

// MountOverlayFilesystems represents the MountOverlayFilesystems task.
func MountOverlayFilesystems(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var mountpoints *mount.Points

		mountpoints, err = mount.OverlayMountPoints()
		if err != nil {
			return err
		}

		return mount.Mount(mountpoints)
	}, "mountOverlayFilesystems"
}

// SetupSharedFilesystems represents the SetupSharedFilesystems task.
func SetupSharedFilesystems(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		targets := []string{"/", "/var/lib/kubelet", "/etc/cni", "/run"}
		for _, t := range targets {
			if err = unix.Mount("", t, "", unix.MS_SHARED|unix.MS_REC, ""); err != nil {
				return err
			}
		}

		return nil
	}, "setupSharedFilesystems"
}

// SetupVarDirectory represents the SetupVarDirectory task.
func SetupVarDirectory(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		for _, p := range []string{"/var/log/pods", "/var/lib/kubelet", "/var/run/lock"} {
			if err = os.MkdirAll(p, 0o700); err != nil {
				return err
			}
		}

		return nil
	}, "setupVarDirectory"
}

// MountUserDisks represents the MountUserDisks task.
func MountUserDisks(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if err = partitionAndFormatDisks(logger, r); err != nil {
			return err
		}

		return mountDisks(r)
	}, "mountUserDisks"
}

// TODO(andrewrynhard): We shouldn't pull in the installer command package
// here.
func partitionAndFormatDisks(logger *log.Logger, r runtime.Runtime) error {
	m := &installer.Manifest{
		Devices: map[string]installer.Device{},
		Targets: map[string][]*installer.Target{},
	}

	for _, disk := range r.Config().Machine().Disks() {
		disk := disk

		if err := func() error {
			bd, err := blockdevice.Open(disk.Device())
			if err != nil {
				return err
			}

			//nolint:errcheck
			defer bd.Close()

			var pt *gpt.GPT

			pt, err = bd.PartitionTable()
			if err != nil {
				if !errors.Is(err, blockdevice.ErrMissingPartitionTable) {
					return err
				}
			}

			// Partitions will be created/recreated if either of the following
			//  conditions are true:
			// - a partition table exists AND there are no partitions
			// - a partition table does not exist

			if pt != nil {
				if len(pt.Partitions().Items()) > 0 {
					logger.Printf(("skipping setup of %q, found existing partitions"), disk.Device())

					return nil
				}
			}

			m.Devices[disk.Device()] = installer.Device{
				Device:                 disk.Device(),
				ResetPartitionTable:    true,
				SkipOverlayMountsCheck: true,
			}

			for _, part := range disk.Partitions() {
				extraTarget := &installer.Target{
					Device: disk.Device(),
					FormatOptions: &partition.FormatOptions{
						Size:           part.Size(),
						Force:          true,
						PartitionType:  partition.LinuxFilesystemData,
						FileSystemType: partition.FilesystemTypeXFS,
					},
				}

				m.Targets[disk.Device()] = append(m.Targets[disk.Device()], extraTarget)
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return m.Execute()
}

func mountDisks(r runtime.Runtime) (err error) {
	mountpoints := mount.NewMountPoints()

	for _, disk := range r.Config().Machine().Disks() {
		for i, part := range disk.Partitions() {
			var partname string

			partname, err = util.PartPath(disk.Device(), i+1)
			if err != nil {
				return err
			}

			if _, err = os.Stat(part.MountPoint()); errors.Is(err, os.ErrNotExist) {
				if err = os.MkdirAll(part.MountPoint(), 0o700); err != nil {
					return err
				}
			}

			mountpoints.Set(partname, mount.NewMountPoint(partname, part.MountPoint(), "xfs", unix.MS_NOATIME, ""))
		}
	}

	return mount.Mount(mountpoints)
}

func unmountDisks(r runtime.Runtime) (err error) {
	mountpoints := mount.NewMountPoints()

	for _, disk := range r.Config().Machine().Disks() {
		for i, part := range disk.Partitions() {
			var partname string

			partname, err = util.PartPath(disk.Device(), i+1)
			if err != nil {
				return err
			}

			mountpoints.Set(partname, mount.NewMountPoint(partname, part.MountPoint(), "xfs", unix.MS_NOATIME, ""))
		}
	}

	return mount.Unmount(mountpoints)
}

// WriteUserFiles represents the WriteUserFiles task.
//
//nolint:gocyclo,cyclop
func WriteUserFiles(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var result *multierror.Error

		files, err := r.Config().Machine().Files()
		if err != nil {
			return fmt.Errorf("error generating extra files: %w", err)
		}

		extra, err := containerd.GenerateRegistriesConfig(r.Config().Machine().Registries())
		if err != nil {
			return err
		}

		files = append(files, extra...)

		for _, f := range files {
			content := f.Content()

			switch f.Op() {
			case "create":
				// Allow create at all times.
			case "overwrite":
				if err = existsAndIsFile(f.Path()); err != nil {
					result = multierror.Append(result, err)

					continue
				}
			case "append":
				if err = existsAndIsFile(f.Path()); err != nil {
					result = multierror.Append(result, err)

					continue
				}

				var existingFileContents []byte

				existingFileContents, err = ioutil.ReadFile(f.Path())
				if err != nil {
					result = multierror.Append(result, err)

					continue
				}

				content = string(existingFileContents) + "\n" + f.Content()
			default:
				result = multierror.Append(result, fmt.Errorf("unknown operation for file %q: %q", f.Path(), f.Op()))

				continue
			}

			if filepath.Dir(f.Path()) == constants.ManifestsDirectory {
				if err = ioutil.WriteFile(f.Path(), []byte(content), f.Permissions()); err != nil {
					result = multierror.Append(result, err)

					continue
				}

				if err = os.Chmod(f.Path(), f.Permissions()); err != nil {
					result = multierror.Append(result, err)

					continue
				}

				continue
			}

			// Determine if supplied path is in /var or not.
			// If not, we'll write it to /var anyways and bind mount below
			p := f.Path()
			inVar := true
			parts := strings.Split(
				strings.TrimLeft(f.Path(), "/"),
				string(os.PathSeparator),
			)

			if parts[0] != "var" {
				p = filepath.Join("/var", f.Path())
				inVar = false
			}

			// We do not want to support creating new files anywhere outside of
			// /var. If a valid use case comes up, we can reconsider then.
			if !inVar && f.Op() == "create" {
				return fmt.Errorf("create operation not allowed outside of /var: %q", f.Path())
			}

			if err = os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				result = multierror.Append(result, err)

				continue
			}

			if err = ioutil.WriteFile(p, []byte(content), f.Permissions()); err != nil {
				result = multierror.Append(result, err)

				continue
			}

			if err = os.Chmod(p, f.Permissions()); err != nil {
				result = multierror.Append(result, err)

				continue
			}

			if !inVar {
				if err = unix.Mount(p, f.Path(), "", unix.MS_BIND|unix.MS_RDONLY, ""); err != nil {
					result = multierror.Append(result, fmt.Errorf("failed to create bind mount for %s: %w", p, err))
				}
			}
		}

		return result.ErrorOrNil()
	}, "writeUserFiles"
}

//nolint:deadcode,unused
func doesNotExists(p string) (err error) {
	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return fmt.Errorf("file exists")
}

func existsAndIsFile(p string) (err error) {
	var info os.FileInfo

	info, err = os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		return fmt.Errorf("file must exist: %q", p)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("invalid mode: %q", info.Mode().String())
	}

	return nil
}

// WriteUserSysctls represents the WriteUserSysctls task.
func WriteUserSysctls(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var result *multierror.Error

		for k, v := range r.Config().Machine().Sysctls() {
			if err = sysctl.WriteSystemProperty(&sysctl.SystemProperty{Key: k, Value: v}); err != nil {
				return err
			}
		}

		return result.ErrorOrNil()
	}, "writeUserSysctls"
}

// UnmountOverlayFilesystems represents the UnmountOverlayFilesystems task.
func UnmountOverlayFilesystems(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var mountpoints *mount.Points

		mountpoints, err = mount.OverlayMountPoints()
		if err != nil {
			return err
		}

		return mount.Unmount(mountpoints)
	}, "unmountOverlayFilesystems"
}

// UnmountUserDisks represents the UnmountUserDisks task.
func UnmountUserDisks(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return unmountDisks(r)
	}, "unmountUserDisks"
}

// UnmountPodMounts represents the UnmountPodMounts task.
func UnmountPodMounts(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var b []byte

		if b, err = ioutil.ReadFile("/proc/self/mounts"); err != nil {
			return err
		}

		rdr := bytes.NewReader(b)

		scanner := bufio.NewScanner(rdr)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())

			if len(fields) < 2 {
				continue
			}

			mountpoint := fields[1]
			if strings.HasPrefix(mountpoint, constants.EphemeralMountPoint+"/") {
				logger.Printf("unmounting %s\n", mountpoint)

				if err = unix.Unmount(mountpoint, 0); err != nil {
					if errors.Is(err, syscall.EINVAL) {
						log.Printf("ignoring unmount error %s: %v", mountpoint, err)
					} else {
						return fmt.Errorf("error unmounting %s: %w", mountpoint, err)
					}
				}
			}
		}

		return scanner.Err()
	}, "unmountPodMounts"
}

// UnmountSystemDiskBindMounts represents the UnmountSystemDiskBindMounts task.
func UnmountSystemDiskBindMounts(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		devname := r.State().Machine().Disk().BlockDevice.Device().Name()

		f, err := os.Open("/proc/mounts")
		if err != nil {
			return err
		}

		defer f.Close() //nolint:errcheck

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())

			if len(fields) < 2 {
				continue
			}

			device := strings.ReplaceAll(fields[0], "/dev/mapper", "/dev")
			mountpoint := fields[1]

			if strings.HasPrefix(device, devname) && device != devname {
				logger.Printf("unmounting %s\n", mountpoint)

				if err = unix.Unmount(mountpoint, 0); err != nil {
					if errors.Is(err, syscall.EINVAL) {
						log.Printf("ignoring unmount error %s: %v", mountpoint, err)
					} else {
						return fmt.Errorf("error unmounting %s: %w", mountpoint, err)
					}
				}
			}
		}

		return scanner.Err()
	}, "unmountSystemDiskBindMounts"
}

// CordonAndDrainNode represents the task for stop all containerd tasks in the
// k8s.io namespace.
func CordonAndDrainNode(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var nodename string

		if nodename, err = r.NodeName(); err != nil {
			return err
		}

		var kubeHelper *kubernetes.Client

		if kubeHelper, err = kubernetes.NewClientFromKubeletKubeconfig(); err != nil {
			return err
		}

		defer kubeHelper.Close() //nolint:errcheck

		return kubeHelper.CordonAndDrain(ctx, nodename)
	}, "cordonAndDrainNode"
}

// UncordonNode represents the task for mark node as scheduling enabled.
//
// This action undoes the CordonAndDrainNode task.
func UncordonNode(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		var nodename string

		if nodename, err = r.NodeName(); err != nil {
			return err
		}

		var kubeHelper *kubernetes.Client

		if err = retry.Constant(5*time.Minute, retry.WithUnits(time.Second), retry.WithErrorLogging(true)).RetryWithContext(ctx,
			func(ctx context.Context) error {
				kubeHelper, err = kubernetes.NewClientFromKubeletKubeconfig()

				return retry.ExpectedError(err)
			}); err != nil {
			return err
		}

		defer kubeHelper.Close() //nolint:errcheck

		if err = kubeHelper.WaitUntilReady(ctx, nodename); err != nil {
			return err
		}

		return kubeHelper.Uncordon(ctx, nodename, false)
	}, "uncordonNode"
}

// LeaveEtcd represents the task for removing a control plane node from etcd.
func LeaveEtcd(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		client, err := etcd.NewClientFromControlPlaneIPs(ctx, r.Config().Cluster().CA(), r.Config().Cluster().Endpoint())
		if err != nil {
			return fmt.Errorf("failed to create etcd client: %w", err)
		}

		//nolint:errcheck
		defer client.Close()

		ctx = clientv3.WithRequireLeader(ctx)

		if err = client.LeaveCluster(ctx); err != nil {
			return fmt.Errorf("failed to leave cluster: %w", err)
		}

		return nil
	}, "leaveEtcd"
}

// RemoveAllPods represents the task for stopping and removing all pods.
func RemoveAllPods(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return stopAndRemoveAllPods(cri.StopAndRemove), "removeAllPods"
}

// StopAllPods represents the task for stopping all pods.
func StopAllPods(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return stopAndRemoveAllPods(cri.StopOnly), "stopAllPods"
}

func stopAndRemoveAllPods(stopAction cri.StopAction) runtime.TaskExecutionFunc {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if err = system.Services(nil).Stop(ctx, "kubelet"); err != nil {
			return err
		}

		client, err := cri.NewClient("unix://"+constants.CRIContainerdAddress, 10*time.Second)
		if err != nil {
			return err
		}

		//nolint:errcheck
		defer client.Close()

		// We remove pods with POD network mode first so that the CNI can perform
		// any cleanup tasks. If we don't do this, we run the risk of killing the
		// CNI, preventing the CRI from cleaning up the pod's netwokring.

		if err = client.StopAndRemovePodSandboxes(ctx, stopAction, runtimeapi.NamespaceMode_POD, runtimeapi.NamespaceMode_CONTAINER); err != nil {
			return err
		}

		// With the POD network mode pods out of the way, we kill the remaining
		// pods.

		return client.StopAndRemovePodSandboxes(ctx, stopAction)
	}
}

// ResetSystemDisk represents the task to reset the system disk.
func ResetSystemDisk(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return r.State().Machine().Disk().BlockDevice.Reset()
	}, "resetSystemDisk"
}

// ResetSystemDiskSpec represents the task to reset the system disk by spec.
func ResetSystemDiskSpec(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		in, ok := data.(runtime.ResetOptions)
		if !ok {
			return fmt.Errorf("unexpected runtime data")
		}

		for _, target := range in.GetSystemDiskTargets() {
			if err = target.Format(); err != nil {
				return fmt.Errorf("failed wiping partition %s: %w", target, err)
			}
		}

		logger.Printf("successfully reset system disk by the spec")

		return nil
	}, "resetSystemDiskSpec"
}

// VerifyDiskAvailability represents the task for verifying that the system
// disk is not in use.
func VerifyDiskAvailability(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		devname := r.State().Machine().Disk().BlockDevice.Device().Name()

		// We MUST close this in order to avoid EBUSY.
		if err = r.State().Machine().Close(); err != nil {
			return err
		}

		// TODO(andrewrynhard): This should be more dynamic. If we ever change the
		// partition scheme there is the chance that 2 is not the correct parition to
		// check.
		partname, err := util.PartPath(devname, 2)
		if err != nil {
			return err
		}

		if _, err = os.Stat(partname); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("ephemeral partition not found: %w", err)
		}

		mountsReported := false

		return retry.Constant(3*time.Minute, retry.WithUnits(500*time.Millisecond)).Retry(func() error {
			if err = tryLock(partname); err != nil {
				if err == unix.EBUSY {
					if !mountsReported {
						// if disk is busy, report mounts for debugging purposes but just once
						// otherwise console might be flooded with messages
						dumpMounts(logger)
						mountsReported = true
					}

					return retry.ExpectedError(fmt.Errorf("ephemeral partition in use: %q", partname))
				}

				return fmt.Errorf("failed to verify ephemeral partition not in use: %w", err)
			}

			return nil
		})
	}, "verifyDiskAvailability"
}

func tryLock(path string) error {
	fd, errno := unix.Open(path, unix.O_RDONLY|unix.O_EXCL|unix.O_CLOEXEC, 0)

	//nolint:errcheck
	defer unix.Close(fd)

	return errno
}

func dumpMounts(logger *log.Logger) {
	mounts, err := os.Open("/proc/mounts")
	if err != nil {
		logger.Printf("failed to read mounts: %s", err)

		return
	}

	defer mounts.Close() //nolint:errcheck

	logger.Printf("contents of /proc/mounts:")

	_, _ = io.Copy(log.Writer(), mounts) //nolint:errcheck
}

// Upgrade represents the task for performing an upgrade.
func Upgrade(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		// This should be checked by the gRPC server, but we double check here just
		// to be safe.
		in, ok := data.(*machineapi.UpgradeRequest)
		if !ok {
			return runtime.ErrInvalidSequenceData
		}

		devname := r.State().Machine().Disk().BlockDevice.Device().Name()

		logger.Printf("performing upgrade via %q", in.GetImage())

		configBytes, err := r.Config().Bytes()
		if err != nil {
			return fmt.Errorf("error marshaling configuration: %w", err)
		}

		// We pull the installer image when we receive an upgrade request. No need
		// to pull it again.
		err = install.RunInstallerContainer(
			devname, r.State().Platform().Name(),
			in.GetImage(),
			configBytes,
			r.Config().Machine().Registries(),
			install.OptionsFromUpgradeRequest(r, in)...,
		)
		if err != nil {
			return err
		}

		logger.Println("upgrade successful")

		return nil
	}, "upgrade"
}

// LabelNodeAsMaster represents the LabelNodeAsMaster task.
func LabelNodeAsMaster(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		h, err := kubernetes.NewTemporaryClientFromPKI(r.Config().Cluster().CA(), r.Config().Cluster().Endpoint())
		if err != nil {
			return err
		}

		defer h.Close() //nolint:errcheck

		var nodename string

		if nodename, err = r.NodeName(); err != nil {
			return err
		}

		err = retry.Constant(constants.NodeReadyTimeout, retry.WithUnits(3*time.Second), retry.WithErrorLogging(true)).RetryWithContext(ctx, func(ctx context.Context) error {
			if err = h.LabelNodeAsMaster(ctx, nodename, !r.Config().Cluster().ScheduleOnMasters()); err != nil {
				return retry.ExpectedError(err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to label node as master: %w", err)
		}

		return nil
	}, "labelNodeAsMaster"
}

// UpdateBootloader represents the UpdateBootloader task.
func UpdateBootloader(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		meta, err := bootloader.NewMeta()
		if err != nil {
			return err
		}
		//nolint:errcheck
		defer meta.Close()

		if ok := meta.LegacyADV.DeleteTag(adv.Upgrade); ok {
			logger.Println("removing fallback")

			if err = meta.Write(); err != nil {
				return err
			}
		}

		return nil
	}, "updateBootloader"
}

// Reboot represents the Reboot task.
func Reboot(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		r.Events().Publish(&machineapi.RestartEvent{
			Cmd: unix.LINUX_REBOOT_CMD_RESTART,
		})

		return runtime.RebootError{Cmd: unix.LINUX_REBOOT_CMD_RESTART}
	}, "reboot"
}

// Shutdown represents the Shutdown task.
func Shutdown(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		cmd := unix.LINUX_REBOOT_CMD_POWER_OFF

		if p := procfs.ProcCmdline().Get(constants.KernelParamShutdown).First(); p != nil {
			if *p == "halt" {
				cmd = unix.LINUX_REBOOT_CMD_HALT
			}
		}

		r.Events().Publish(&machineapi.RestartEvent{
			Cmd: int64(cmd),
		})

		return runtime.RebootError{Cmd: cmd}
	}, "shutdown"
}

// SaveStateEncryptionConfig saves state partition encryption info in the meta partition.
func SaveStateEncryptionConfig(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		config := r.Config()
		if config == nil {
			return nil
		}

		encryption := config.Machine().SystemDiskEncryption().Get(constants.StatePartitionLabel)
		if encryption == nil {
			return nil
		}

		meta, err := bootloader.NewMeta()
		if err != nil {
			return err
		}
		//nolint:errcheck
		defer meta.Close()

		var data []byte

		if data, err = json.Marshal(encryption); err != nil {
			return err
		}

		if !meta.ADV.SetTagBytes(adv.StateEncryptionConfig, data) {
			return fmt.Errorf("failed to save state encryption config in the META partition")
		}

		return meta.Write()
	}, "SaveStateEncryptionConfig"
}

// MountBootPartition mounts the boot partition.
func MountBootPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return mount.SystemPartitionMount(r, constants.BootPartitionLabel)
	}, "mountBootPartition"
}

// UnmountBootPartition unmounts the boot partition.
func UnmountBootPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) error {
		return mount.SystemPartitionUnmount(r, constants.BootPartitionLabel)
	}, "unmountBootPartition"
}

// MountEFIPartition mounts the EFI partition.
func MountEFIPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return mount.SystemPartitionMount(r, constants.EFIPartitionLabel)
	}, "mountEFIPartition"
}

// UnmountEFIPartition unmounts the EFI partition.
func UnmountEFIPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) error {
		return mount.SystemPartitionUnmount(r, constants.EFIPartitionLabel)
	}, "unmountEFIPartition"
}

// MountStatePartition mounts the system partition.
func MountStatePartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		meta, err := bootloader.NewMeta()
		if err != nil {
			return err
		}
		//nolint:errcheck
		defer meta.Close()

		flags := mount.SkipIfMounted

		if seq == runtime.SequenceInitialize {
			flags |= mount.SkipIfNoFilesystem
		}

		opts := []mount.Option{mount.WithFlags(flags)}

		var encryption config.Encryption
		// first try reading encryption from the config
		// config always has the priority here
		if r.Config() != nil && r.Config().Machine() != nil {
			encryption = r.Config().Machine().SystemDiskEncryption().Get(constants.StatePartitionLabel)
		}

		// then try reading it from the META partition
		if encryption == nil {
			var encryptionFromMeta *v1alpha1.EncryptionConfig

			data, ok := meta.ADV.ReadTagBytes(adv.StateEncryptionConfig)
			if ok {
				if err = json.Unmarshal(data, &encryptionFromMeta); err != nil {
					return err
				}

				encryption = encryptionFromMeta
			}
		}

		if encryption != nil {
			opts = append(opts, mount.WithEncryptionConfig(encryption))
		}

		return mount.SystemPartitionMount(r, constants.StatePartitionLabel, opts...)
	}, "mountStatePartition"
}

// UnmountStatePartition unmounts the system partition.
func UnmountStatePartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) error {
		return mount.SystemPartitionUnmount(r, constants.StatePartitionLabel)
	}, "unmountStatePartition"
}

// MountEphemeralPartition mounts the ephemeral partition.
func MountEphemeralPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) error {
		return mount.SystemPartitionMount(r, constants.EphemeralPartitionLabel, mount.WithFlags(mount.Resize))
	}, "mountEphemeralPartition"
}

// UnmountEphemeralPartition unmounts the ephemeral partition.
func UnmountEphemeralPartition(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		return mount.SystemPartitionUnmount(r, constants.EphemeralPartitionLabel)
	}, "unmountEphemeralPartition"
}

// Install mounts or installs the system partitions.
func Install(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		configBytes, err := r.Config().Bytes()
		if err != nil {
			return fmt.Errorf("error marshaling configuration: %w", err)
		}

		switch {
		case !r.State().Machine().Installed():
			installerImage := r.Config().Machine().Install().Image()
			if installerImage == "" {
				installerImage = images.DefaultInstallerImage
			}

			var disk string

			disk, err = r.Config().Machine().Install().Disk()
			if err != nil {
				return err
			}

			err = install.RunInstallerContainer(
				disk,
				r.State().Platform().Name(),
				installerImage,
				configBytes,
				r.Config().Machine().Registries(),
				install.WithForce(true),
				install.WithZero(r.Config().Machine().Install().Zero()),
				install.WithExtraKernelArgs(r.Config().Machine().Install().ExtraKernelArgs()),
			)
			if err != nil {
				return err
			}

			logger.Println("install successful")

		case r.State().Machine().IsInstallStaged():
			devname := r.State().Machine().Disk().BlockDevice.Device().Name()

			var options install.Options

			if err = json.Unmarshal(r.State().Machine().StagedInstallOptions(), &options); err != nil {
				return fmt.Errorf("error unserializing install options: %w", err)
			}

			logger.Printf("performing staged upgrade via %q", r.State().Machine().StagedInstallImageRef())

			err = install.RunInstallerContainer(
				devname, r.State().Platform().Name(),
				r.State().Machine().StagedInstallImageRef(),
				configBytes,
				r.Config().Machine().Registries(),
				install.WithOptions(options),
			)
			if err != nil {
				return err
			}

			logger.Println("staged upgrade successful")

		default:
			return fmt.Errorf("unsupported configuration for install task")
		}

		return nil
	}, "install"
}

// BootstrapEtcd represents the task for bootstrapping etcd.
//
//nolint:gocyclo
func BootstrapEtcd(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		req, ok := data.(*machineapi.BootstrapRequest)
		if !ok {
			return fmt.Errorf("failed to typecast boostrap request")
		}

		if err = system.Services(r).Stop(ctx, "etcd"); err != nil {
			return fmt.Errorf("failed to stop etcd: %w", err)
		}

		// This is hack. We need to fake a finished state so that we can get the
		// wait in the boot sequence to unblock.
		for _, svc := range system.Services(r).List() {
			if svc.AsProto().GetId() == "etcd" {
				svc.UpdateState(events.StateFinished, "Bootstrap requested")

				break
			}
		}

		if err = func() error {
			// Since etcd has already attempted to start, we must delete the data. If
			// we don't, then an initial cluster state of "new" will fail.
			var dir *os.File

			dir, err = os.Open(constants.EtcdDataPath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}

				return err
			}

			//nolint:errcheck
			defer dir.Close()

			var files []os.FileInfo

			files, err = dir.Readdir(0)
			if err != nil {
				return err
			}

			for _, file := range files {
				fullPath := filepath.Join(constants.EtcdDataPath, file.Name())

				if err = os.RemoveAll(fullPath); err != nil {
					return fmt.Errorf("failed to remove %q: %w", file.Name(), err)
				}
			}

			return nil
		}(); err != nil {
			return err
		}

		svc := &services.Etcd{
			Bootstrap:            true,
			RecoverFromSnapshot:  req.RecoverEtcd,
			RecoverSkipHashCheck: req.RecoverSkipHashCheck,
		}

		if err = system.Services(r).Unload(ctx, svc.ID(r)); err != nil {
			return err
		}

		system.Services(r).Load(svc)

		if err = system.Services(r).Start(svc.ID(r)); err != nil {
			return fmt.Errorf("error starting etcd in bootstrap mode: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		return system.WaitForService(system.StateEventUp, svc.ID(r)).Wait(ctx)
	}, "bootstrapEtcd"
}

// ActivateLogicalVolumes represents the task for activating logical volumes.
func ActivateLogicalVolumes(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		if _, err = cmd.Run("/sbin/lvm", "vgchange", "-ay"); err != nil {
			return fmt.Errorf("failed to activate logical volumes: %w", err)
		}

		return nil
	}, "activateLogicalVolumes"
}

// CheckControlPlaneStatus represents the CheckControlPlaneStatus task.
func CheckControlPlaneStatus(seq runtime.Sequence, data interface{}) (runtime.TaskExecutionFunc, string) {
	return func(ctx context.Context, logger *log.Logger, r runtime.Runtime) (err error) {
		res, err := r.State().V1Alpha2().Resources().Get(ctx, resourcev1alpha1.NewBootstrapStatus().Metadata())
		if err != nil {
			logger.Printf("error getting bootstrap status: %s", err)

			return nil
		}

		if res.(*resourcev1alpha1.BootstrapStatus).TypedSpec().SelfHostedControlPlane {
			log.Printf("WARNING: Talos is running self-hosted control plane, convert to static pods using `talosctl convert-k8s`.")
		}

		return nil
	}, "checkControlPlaneStatus"
}
