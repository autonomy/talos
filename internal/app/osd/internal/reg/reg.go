/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package reg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	criconstants "github.com/containerd/cri/pkg/constants"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/go-multierror"
	"github.com/jsimonetti/rtnetlink"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"

	initproto "github.com/talos-systems/talos/internal/app/init/proto"
	"github.com/talos-systems/talos/internal/app/osd/proto"
	"github.com/talos-systems/talos/internal/pkg/chunker"
	filechunker "github.com/talos-systems/talos/internal/pkg/chunker/file"
	"github.com/talos-systems/talos/internal/pkg/constants"
	"github.com/talos-systems/talos/internal/pkg/containers"
	"github.com/talos-systems/talos/internal/pkg/containers/containerd"
	"github.com/talos-systems/talos/internal/pkg/containers/cri"
	"github.com/talos-systems/talos/internal/pkg/proc"
	"github.com/talos-systems/talos/internal/pkg/version"
	"github.com/talos-systems/talos/pkg/userdata"
)

// Registrator is the concrete type that implements the factory.Registrator and
// proto.OSDServer interfaces.
type Registrator struct {
	// every Init service API is proxied via OSD
	*InitServiceClient

	Data *userdata.UserData
}

// Register implements the factory.Registrator interface.
func (r *Registrator) Register(s *grpc.Server) {
	proto.RegisterOSDServer(s, r)
	initproto.RegisterInitServer(s, r)
}

// Kubeconfig implements the proto.OSDServer interface. The admin kubeconfig is
// generated by kubeadm and placed at /etc/kubernetes/admin.conf. This method
// returns the contents of the generated admin.conf in the response.
func (r *Registrator) Kubeconfig(ctx context.Context, in *empty.Empty) (data *proto.Data, err error) {
	fileBytes, err := ioutil.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		return
	}
	data = &proto.Data{
		Bytes: fileBytes,
	}

	return data, err
}

// Processes implements the proto.OSDServer interface.
func (r *Registrator) Processes(ctx context.Context, in *proto.ProcessesRequest) (reply *proto.ProcessesReply, err error) {
	inspector, err := getContainerInspector(ctx, in.Namespace, in.Driver)
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer inspector.Close()

	pods, err := inspector.Pods()
	if err != nil {
		// fatal error
		if pods == nil {
			return nil, err
		}
		// TODO: only some failed, need to handle it better via client
		log.Println(err.Error())
	}

	processes := []*proto.Process{}

	for _, pod := range pods {
		for _, container := range pod.Containers {
			process := &proto.Process{
				Namespace: in.Namespace,
				Id:        container.Display,
				PodId:     pod.Name,
				Name:      container.Name,
				Image:     container.Image,
				Pid:       container.Pid,
				Status:    container.Status,
			}
			processes = append(processes, process)
		}
	}

	return &proto.ProcessesReply{Processes: processes}, nil

}

// Stats implements the proto.OSDServer interface.
// nolint: gocyclo
func (r *Registrator) Stats(ctx context.Context, in *proto.StatsRequest) (reply *proto.StatsReply, err error) {
	inspector, err := getContainerInspector(ctx, in.Namespace, in.Driver)
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer inspector.Close()

	pods, err := inspector.Pods()
	if err != nil {
		// fatal error
		if pods == nil {
			return nil, err
		}
		// TODO: only some failed, need to handle it better via client
		log.Println(err.Error())
	}

	stats := []*proto.Stat{}

	for _, pod := range pods {
		for _, container := range pod.Containers {
			if container.Metrics == nil {
				continue
			}

			stat := &proto.Stat{
				Namespace:   in.Namespace,
				Id:          container.Display,
				PodId:       pod.Name,
				Name:        container.Name,
				MemoryUsage: container.Metrics.MemoryUsage,
				CpuUsage:    container.Metrics.CPUUsage,
			}

			stats = append(stats, stat)
		}

	}

	reply = &proto.StatsReply{Stats: stats}

	return reply, nil
}

// Restart implements the proto.OSDServer interface.
func (r *Registrator) Restart(ctx context.Context, in *proto.RestartRequest) (*proto.RestartReply, error) {
	inspector, err := getContainerInspector(ctx, in.Namespace, in.Driver)
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer inspector.Close()

	container, err := inspector.Container(in.Id)
	if err != nil {
		return nil, err
	}

	if container == nil {
		return nil, fmt.Errorf("container %q not found", in.Id)
	}

	err = container.Kill(syscall.SIGTERM)
	if err != nil {
		return nil, err
	}

	return &proto.RestartReply{}, nil
}

// Dmesg implements the proto.OSDServer interface. The klogctl syscall is used
// to read from the ring buffer at /proc/kmsg by taking the
// SYSLOG_ACTION_READ_ALL action. This action reads all messages remaining in
// the ring buffer non-destructively.
func (r *Registrator) Dmesg(ctx context.Context, in *empty.Empty) (data *proto.Data, err error) {
	// Return the size of the kernel ring buffer
	size, err := unix.Klogctl(constants.SYSLOG_ACTION_SIZE_BUFFER, nil)
	if err != nil {
		return
	}
	// Read all messages from the log (non-destructively)
	buf := make([]byte, size)
	n, err := unix.Klogctl(constants.SYSLOG_ACTION_READ_ALL, buf)
	if err != nil {
		return
	}

	data = &proto.Data{Bytes: buf[:n]}

	return data, err
}

// Logs implements the proto.OSDServer interface. Service or container logs can
// be requested and the contents of the log file are streamed in chunks.
// nolint: gocyclo
func (r *Registrator) Logs(req *proto.LogsRequest, l proto.OSD_LogsServer) (err error) {
	var chunk chunker.Chunker

	switch {
	case req.Namespace == "system" || req.Id == "kubelet" || req.Id == "kubeadm":
		filename := filepath.Join("/var/log", filepath.Base(req.Id)+".log")
		var file *os.File
		file, err = os.OpenFile(filename, os.O_RDONLY, 0)
		if err != nil {
			return
		}
		// nolint: errcheck
		defer file.Close()

		chunk = filechunker.NewChunker(file)
	default:
		var file io.Closer
		if chunk, file, err = k8slogs(l.Context(), req); err != nil {
			return err
		}
		// nolint: errcheck
		defer file.Close()
	}

	for data := range chunk.Read(l.Context()) {
		if err = l.Send(&proto.Data{Bytes: data}); err != nil {
			return
		}
	}

	return nil
}

// Routes implements the proto.OSDServer interface.
func (r *Registrator) Routes(ctx context.Context, in *empty.Empty) (*proto.RoutesReply, error) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return nil, errors.Errorf("failed to open socket to rtnetlink: %v", err)
	}
	// nolint: errcheck
	defer conn.Close()

	list, err := conn.Route.List()
	if err != nil {
		return nil, errors.Errorf("failed to get route list: %v", err)
	}

	routes := []*proto.Route{}

	for _, rMesg := range list {

		var ifaceName string
		ifaceData, err := conn.Link.Get(rMesg.Attributes.OutIface)
		if err != nil {
			log.Printf("failed to get interface details for interface index %d: %v", rMesg.Attributes.OutIface, err)
			// TODO: Remove once we get this sorted on why there's a
			// failure here
			log.Printf("%+v", rMesg)
			continue
		}
		if ifaceData.Attributes != nil {
			ifaceName = ifaceData.Attributes.Name
		}

		routes = append(routes, &proto.Route{
			Interface:   ifaceName,
			Destination: toCIDR(rMesg.Family, rMesg.Attributes.Dst, int(rMesg.DstLength)),
			Gateway:     rMesg.Attributes.Gateway.String(),
			Metric:      rMesg.Attributes.Priority,
			Scope:       uint32(rMesg.Scope),
			Source:      toCIDR(rMesg.Family, rMesg.Attributes.Src, int(rMesg.SrcLength)),
			Family:      proto.AddressFamily(rMesg.Family),
			Protocol:    proto.RouteProtocol(rMesg.Protocol),
			Flags:       rMesg.Flags,
		})

	}
	return &proto.RoutesReply{
		Routes: routes,
	}, nil
}

// Version implements the proto.OSDServer interface.
func (r *Registrator) Version(ctx context.Context, in *empty.Empty) (data *proto.Data, err error) {
	v, err := version.NewVersion()
	if err != nil {
		return
	}

	data = &proto.Data{Bytes: []byte(v)}

	return data, err
}

// Top implements the proto.OSDServer interface
func (r *Registrator) Top(ctx context.Context, in *empty.Empty) (reply *proto.TopReply, err error) {
	var procs []proc.ProcessList
	procs, err = proc.List()
	if err != nil {
		return
	}

	var plist bytes.Buffer
	enc := gob.NewEncoder(&plist)
	err = enc.Encode(procs)
	if err != nil {
		return
	}

	p := &proto.ProcessList{Bytes: plist.Bytes()}
	reply = &proto.TopReply{ProcessList: p}
	return
}

// DF implements the proto.OSDServer interface.
func (r *Registrator) DF(ctx context.Context, in *empty.Empty) (reply *proto.DFReply, err error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	// nolint: errcheck
	defer file.Close()

	var (
		stat     unix.Statfs_t
		multiErr *multierror.Error
	)

	stats := []*proto.DFStat{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 2 {
			continue
		}

		filesystem := fields[0]
		mountpoint := fields[1]

		f, err := os.Stat(mountpoint)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}

		if mode := f.Mode(); !mode.IsDir() {
			continue
		}

		if err := unix.Statfs(mountpoint, &stat); err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}

		totalSize := uint64(stat.Bsize) * stat.Blocks
		totalAvail := uint64(stat.Bsize) * stat.Bavail

		stat := &proto.DFStat{
			Filesystem: filesystem,
			Size:       totalSize,
			Available:  totalAvail,
			MountedOn:  mountpoint,
		}

		stats = append(stats, stat)
	}

	if err := scanner.Err(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	reply = &proto.DFReply{
		Stats: stats,
	}

	return reply, multiErr.ErrorOrNil()
}

func getContainerInspector(ctx context.Context, namespace string, driver proto.ContainerDriver) (containers.Inspector, error) {
	switch driver {
	case proto.ContainerDriver_CRI:
		if namespace != criconstants.K8sContainerdNamespace {
			return nil, errors.New("CRI inspector is supported only for K8s namespace")
		}
		return cri.NewInspector(ctx)
	case proto.ContainerDriver_CONTAINERD:
		return containerd.NewInspector(ctx, namespace)
	default:
		return nil, errors.Errorf("unsupported driver %q", driver)
	}
}

func k8slogs(ctx context.Context, req *proto.LogsRequest) (chunker.Chunker, io.Closer, error) {
	inspector, err := getContainerInspector(ctx, req.Namespace, req.Driver)
	if err != nil {
		return nil, nil, err
	}
	// nolint: errcheck
	defer inspector.Close()

	container, err := inspector.Container(req.Id)
	if err != nil {
		return nil, nil, err
	}
	if container == nil {
		return nil, nil, fmt.Errorf("container %q not found", req.Id)
	}

	return container.GetLogChunker()
}

func toCIDR(family uint8, prefix net.IP, prefixLen int) string {
	var netLen = 32
	if family == unix.AF_INET6 {
		netLen = 128
	}
	ipNet := &net.IPNet{
		IP:   prefix,
		Mask: net.CIDRMask(prefixLen, netLen),
	}
	return ipNet.String()
}
