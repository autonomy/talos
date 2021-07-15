---
title: CLI
desription: Talosctl CLI tool reference.
---

<!-- markdownlint-disable -->

## talosctl apply-config

Apply a new configuration to a node

```
talosctl apply-config [flags]
```

### Options

```
      --cert-fingerprint strings   list of server certificate fingeprints to accept (defaults to no check)
  -f, --file string                the filename of the updated configuration
  -h, --help                       help for apply-config
      --immediate                  apply the config immediately (without a reboot)
  -i, --insecure                   apply the config using the insecure (encrypted with no auth) maintenance service
      --interactive                apply the config using text based interactive mode
      --on-reboot                  apply the config on reboot
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl bootstrap

Bootstrap the etcd cluster on the specified node.

### Synopsis

When Talos cluster is created etcd service on control plane nodes enter the join loop waiting
to join etcd peers from other control plane nodes. One node should be picked as the boostrap node.
When boostrap command is issued, the node aborts join process and bootstraps etcd cluster as a single node cluster.
Other control plane nodes will join etcd cluster once Kubernetes is boostrapped on the bootstrap node.

This command should not be used when "init" type node are used.

Talos etcd cluster can be recovered from a known snapshot with '--recover-from=' flag.

```
talosctl bootstrap [flags]
```

### Options

```
  -h, --help                      help for bootstrap
      --recover-from string       recover etcd cluster from the snapshot
      --recover-skip-hash-check   skip integrity check when recovering etcd (use when recovering from data directory copy)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl cluster create

Creates a local docker-based or QEMU-based kubernetes cluster

```
talosctl cluster create [flags]
```

### Options

```
      --arch string                             cluster architecture (default "amd64")
      --bad-rtc                                 launch VM with bad RTC state (QEMU only)
      --cidr string                             CIDR of the cluster network (IPv4, ULA network for IPv6 is derived in automated way) (default "10.5.0.0/24")
      --cni-bin-path strings                    search path for CNI binaries (VM only) (default [/home/user/.talos/cni/bin])
      --cni-bundle-url string                   URL to download CNI bundle from (VM only) (default "https://github.com/talos-systems/talos/releases/download/v0.11.0-alpha.2/talosctl-cni-bundle-${ARCH}.tar.gz")
      --cni-cache-dir string                    CNI cache directory path (VM only) (default "/home/user/.talos/cni/cache")
      --cni-conf-dir string                     CNI config directory path (VM only) (default "/home/user/.talos/cni/conf.d")
      --config-patch string                     patch generated machineconfigs (applied to all node types)
      --config-patch-control-plane string       patch generated machineconfigs (applied to 'init' and 'controlplane' types)
      --config-patch-worker string              patch generated machineconfigs (applied to 'worker' type)
      --cpus string                             the share of CPUs as fraction (each container/VM) (default "2.0")
      --crashdump                               print debug crashdump to stderr when cluster startup fails
      --custom-cni-url string                   install custom CNI from the URL (Talos cluster)
      --disk int                                default limit on disk size in MB (each VM) (default 6144)
      --disk-image-path string                  disk image to use
      --dns-domain string                       the dns domain to use for cluster (default "cluster.local")
      --docker-host-ip string                   Host IP to forward exposed ports to (Docker provisioner only) (default "0.0.0.0")
      --encrypt-ephemeral                       enable ephemeral partition encryption
      --encrypt-state                           enable state partition encryption
      --endpoint string                         use endpoint instead of provider defaults
  -p, --exposed-ports string                    Comma-separated list of ports/protocols to expose on init node. Ex -p <hostPort>:<containerPort>/<protocol (tcp or udp)> (Docker provisioner only)
  -h, --help                                    help for create
      --image string                            the image to use (default "ghcr.io/talos-systems/talos:latest")
      --init-node-as-endpoint                   use init node as endpoint instead of any load balancer endpoint
      --initrd-path string                      initramfs image to use (default "_out/initramfs-${ARCH}.xz")
  -i, --input-dir string                        location of pre-generated config files
      --install-image string                    the installer image to use (default "ghcr.io/talos-systems/installer:latest")
      --ipv4                                    enable IPv4 network in the cluster (default true)
      --ipv6                                    enable IPv6 network in the cluster (QEMU provisioner only)
      --iso-path string                         the ISO path to use for the initial boot (VM only)
      --kubernetes-version string               desired kubernetes version to run (default "1.22.0-beta.2")
      --masters int                             the number of masters to create (default 1)
      --memory int                              the limit on memory usage in MB (each container/VM) (default 2048)
      --mtu int                                 MTU of the cluster network (default 1500)
      --nameservers strings                     list of nameservers to use (default [8.8.8.8,1.1.1.1,2001:4860:4860::8888,2606:4700:4700::1111])
      --registry-insecure-skip-verify strings   list of registry hostnames to skip TLS verification for
      --registry-mirror strings                 list of registry mirrors to use in format: <registry host>=<mirror URL>
      --skip-injecting-config                   skip injecting config from embedded metadata server, write config files to current directory
      --skip-kubeconfig                         skip merging kubeconfig from the created cluster
      --talos-version string                    the desired Talos version to generate config for (if not set, defaults to image version)
      --use-vip                                 use a virtual IP for the controlplane endpoint instead of the loadbalancer
      --user-disk strings                       list of disks to create for each VM in format: <mount_point1>:<size1>:<mount_point2>:<size2>
      --vmlinuz-path string                     the compressed kernel image to use (default "_out/vmlinuz-${ARCH}")
      --wait                                    wait for the cluster to be ready before returning (default true)
      --wait-timeout duration                   timeout to wait for the cluster to be ready (default 20m0s)
      --wireguard-cidr string                   CIDR of the wireguard network
      --with-apply-config                       enable apply config when the VM is starting in maintenance mode
      --with-bootloader                         enable bootloader to load kernel and initramfs from disk image after install (default true)
      --with-debug                              enable debug in Talos config to send service logs to the console
      --with-init-node                          create the cluster with an init node
      --with-uefi                               enable UEFI on x86_64 architecture (always enabled for arm64)
      --workers int                             the number of workers to create (default 1)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
      --name string          the name of the cluster (default "talos-default")
  -n, --nodes strings        target the specified nodes
      --provisioner string   Talos cluster provisioner to use (default "docker")
      --state string         directory path to store cluster state (default "/home/user/.talos/clusters")
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl cluster](#talosctl-cluster)	 - A collection of commands for managing local docker-based or firecracker-based clusters

## talosctl cluster destroy

Destroys a local docker-based or firecracker-based kubernetes cluster

```
talosctl cluster destroy [flags]
```

### Options

```
  -h, --help   help for destroy
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
      --name string          the name of the cluster (default "talos-default")
  -n, --nodes strings        target the specified nodes
      --provisioner string   Talos cluster provisioner to use (default "docker")
      --state string         directory path to store cluster state (default "/home/user/.talos/clusters")
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl cluster](#talosctl-cluster)	 - A collection of commands for managing local docker-based or firecracker-based clusters

## talosctl cluster show

Shows info about a local provisioned kubernetes cluster

```
talosctl cluster show [flags]
```

### Options

```
  -h, --help   help for show
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
      --name string          the name of the cluster (default "talos-default")
  -n, --nodes strings        target the specified nodes
      --provisioner string   Talos cluster provisioner to use (default "docker")
      --state string         directory path to store cluster state (default "/home/user/.talos/clusters")
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl cluster](#talosctl-cluster)	 - A collection of commands for managing local docker-based or firecracker-based clusters

## talosctl cluster

A collection of commands for managing local docker-based or firecracker-based clusters

### Options

```
  -h, --help                 help for cluster
      --name string          the name of the cluster (default "talos-default")
      --provisioner string   Talos cluster provisioner to use (default "docker")
      --state string         directory path to store cluster state (default "/home/user/.talos/clusters")
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl cluster create](#talosctl-cluster-create)	 - Creates a local docker-based or QEMU-based kubernetes cluster
* [talosctl cluster destroy](#talosctl-cluster-destroy)	 - Destroys a local docker-based or firecracker-based kubernetes cluster
* [talosctl cluster show](#talosctl-cluster-show)	 - Shows info about a local provisioned kubernetes cluster

## talosctl completion

Output shell completion code for the specified shell (bash or zsh)

### Synopsis

Output shell completion code for the specified shell (bash or zsh).
The shell code must be evaluated to provide interactive
completion of talosctl commands.  This can be done by sourcing it from
the .bash_profile.

Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

```
talosctl completion SHELL [flags]
```

### Examples

```
# Installing bash completion on macOS using homebrew
## If running Bash 3.2 included with macOS
	brew install bash-completion
## or, if running Bash 4.1+
	brew install bash-completion@2
## If talosctl is installed via homebrew, this should start working immediately.
## If you've installed via other means, you may need add the completion to your completion directory
	talosctl completion bash > $(brew --prefix)/etc/bash_completion.d/talosctl

# Installing bash completion on Linux
## If bash-completion is not installed on Linux, please install the 'bash-completion' package
## via your distribution's package manager.
## Load the talosctl completion code for bash into the current shell
	source <(talosctl completion bash)
## Write bash completion code to a file and source if from .bash_profile
	talosctl completion bash > ~/.talos/completion.bash.inc
	printf "
		# talosctl shell completion
		source '$HOME/.talos/completion.bash.inc'
		" >> $HOME/.bash_profile
	source $HOME/.bash_profile
# Load the talosctl completion code for zsh[1] into the current shell
	source <(talosctl completion zsh)
# Set the talosctl completion code for zsh[1] to autoload on startup
talosctl completion zsh > "${fpath[1]}/_talosctl"
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl config add

Add a new context

```
talosctl config add <context> [flags]
```

### Options

```
      --ca string    the path to the CA certificate
      --crt string   the path to the certificate
  -h, --help         help for add
      --key string   the path to the key
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config context

Set the current context

```
talosctl config context <context> [flags]
```

### Options

```
  -h, --help   help for context
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config contexts

List defined contexts

```
talosctl config contexts [flags]
```

### Options

```
  -h, --help   help for contexts
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config endpoint

Set the endpoint(s) for the current context

```
talosctl config endpoint <endpoint>... [flags]
```

### Options

```
  -h, --help   help for endpoint
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config info

Show information about the current context

```
talosctl config info [flags]
```

### Options

```
  -h, --help   help for info
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config merge

Merge additional contexts from another client configuration file

### Synopsis

Contexts with the same name are renamed while merging configs.

```
talosctl config merge <from> [flags]
```

### Options

```
  -h, --help   help for merge
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config new

Generate a new client configuration file

```
talosctl config new [<path>] [flags]
```

### Options

```
      --crt-ttl duration   certificate TTL (default 87600h0m0s)
  -h, --help               help for new
      --roles strings      roles (default [os:admin])
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config node

Set the node(s) for the current context

```
talosctl config node <endpoint>... [flags]
```

### Options

```
  -h, --help   help for node
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)

## talosctl config

Manage the client configuration file (talosconfig)

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl config add](#talosctl-config-add)	 - Add a new context
* [talosctl config context](#talosctl-config-context)	 - Set the current context
* [talosctl config contexts](#talosctl-config-contexts)	 - List defined contexts
* [talosctl config endpoint](#talosctl-config-endpoint)	 - Set the endpoint(s) for the current context
* [talosctl config info](#talosctl-config-info)	 - Show information about the current context
* [talosctl config merge](#talosctl-config-merge)	 - Merge additional contexts from another client configuration file
* [talosctl config new](#talosctl-config-new)	 - Generate a new client configuration file
* [talosctl config node](#talosctl-config-node)	 - Set the node(s) for the current context

## talosctl conformance kubernetes

Run Kubernetes conformance tests

```
talosctl conformance kubernetes [flags]
```

### Options

```
  -h, --help          help for kubernetes
      --mode string   conformance test mode: [fast, certified] (default "fast")
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl conformance](#talosctl-conformance)	 - Run conformance tests

## talosctl conformance

Run conformance tests

### Options

```
  -h, --help   help for conformance
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl conformance kubernetes](#talosctl-conformance-kubernetes)	 - Run Kubernetes conformance tests

## talosctl containers

List containers

```
talosctl containers [flags]
```

### Options

```
  -h, --help         help for containers
  -k, --kubernetes   use the k8s.io containerd namespace
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl convert-k8s

Convert Kubernetes control plane from self-hosted (bootkube) to Talos-managed (static pods).

### Synopsis

Command converts control plane bootstrapped on Talos <= 0.8 to Talos-managed control plane (Talos >= 0.9).
As part of the conversion process tool reads existing configuration of the control plane, updates
Talos node configuration to reflect changes made since the boostrap time. Once config is updated,
tool releases static pods and deletes self-hosted DaemonSets.

```
talosctl convert-k8s [flags]
```

### Options

```
      --endpoint string   the cluster control plane endpoint
      --force             skip prompts, assume yes
  -h, --help              help for convert-k8s
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl copy

Copy data out from the node

### Synopsis

Creates an .tar.gz archive at the node starting at <src-path> and
streams it back to the client.

If '-' is given for <local-path>, archive is written to stdout.
Otherwise archive is extracted to <local-path> which should be an empty directory or
talosctl creates a directory if <local-path> doesn't exist. Command doesn't preserve
ownership and access mode for the files in extract mode, while  streamed .tar archive
captures ownership and permission bits.

```
talosctl copy <src-path> -|<local-path> [flags]
```

### Options

```
  -h, --help   help for copy
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl crashdump

Dump debug information about the cluster

```
talosctl crashdump [flags]
```

### Options

```
      --control-plane-nodes strings   specify IPs of control plane nodes
  -h, --help                          help for crashdump
      --init-node string              specify IPs of init node
      --worker-nodes strings          specify IPs of worker nodes
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl dashboard

Cluster dashboard with real-time metrics

### Synopsis

Provide quick UI to navigate through node real-time metrics.

Keyboard shortcuts:

 - h, <Left>: switch one node to the left
 - l, <Right>: switch one node to the right
 - j, <Down>: scroll process list down
 - k, <Up>: scroll process list up
 - <C-d>: scroll process list half page down
 - <C-u>: scroll process list half page up
 - <C-f>: scroll process list one page down
 - <C-b>: scroll process list one page up


```
talosctl dashboard [flags]
```

### Options

```
  -h, --help                       help for dashboard
  -d, --update-interval duration   interval between updates (default 3s)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl disks

Get the list of disks from /sys/block on the machine

```
talosctl disks [flags]
```

### Options

```
  -h, --help       help for disks
  -i, --insecure   get disks using the insecure (encrypted with no auth) maintenance service
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl dmesg

Retrieve kernel logs

```
talosctl dmesg [flags]
```

### Options

```
  -f, --follow   specify if the kernel log should be streamed
  -h, --help     help for dmesg
      --tail     specify if only new messages should be sent (makes sense only when combined with --follow)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl edit

Edit a resource from the default editor.

### Synopsis

The edit command allows you to directly edit any API resource
you can retrieve via the command line tools.

It will open the editor defined by your TALOS_EDITOR,
or EDITOR environment variables, or fall back to 'vi' for Linux
or 'notepad' for Windows.

```
talosctl edit <type> [<id>] [flags]
```

### Options

```
  -h, --help               help for edit
      --immediate          apply the change immediately (without a reboot)
      --namespace string   resource namespace (default is to use default namespace per resource)
      --on-reboot          apply the change on next reboot
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl etcd forfeit-leadership

Tell node to forfeit etcd cluster leadership

```
talosctl etcd forfeit-leadership [flags]
```

### Options

```
  -h, --help   help for forfeit-leadership
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl etcd](#talosctl-etcd)	 - Manage etcd

## talosctl etcd leave

Tell nodes to leave etcd cluster

```
talosctl etcd leave [flags]
```

### Options

```
  -h, --help   help for leave
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl etcd](#talosctl-etcd)	 - Manage etcd

## talosctl etcd members

Get the list of etcd cluster members

```
talosctl etcd members [flags]
```

### Options

```
  -h, --help   help for members
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl etcd](#talosctl-etcd)	 - Manage etcd

## talosctl etcd remove-member

Remove the node from etcd cluster

### Synopsis

Use this command only if you want to remove a member which is in broken state.
If there is no access to the node, or the node can't access etcd to call etcd leave.
Always prefer etcd leave over this command.

```
talosctl etcd remove-member <hostname> [flags]
```

### Options

```
  -h, --help   help for remove-member
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl etcd](#talosctl-etcd)	 - Manage etcd

## talosctl etcd snapshot

Stream snapshot of the etcd node to the path.

```
talosctl etcd snapshot <path> [flags]
```

### Options

```
  -h, --help   help for snapshot
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl etcd](#talosctl-etcd)	 - Manage etcd

## talosctl etcd

Manage etcd

### Options

```
  -h, --help   help for etcd
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl etcd forfeit-leadership](#talosctl-etcd-forfeit-leadership)	 - Tell node to forfeit etcd cluster leadership
* [talosctl etcd leave](#talosctl-etcd-leave)	 - Tell nodes to leave etcd cluster
* [talosctl etcd members](#talosctl-etcd-members)	 - Get the list of etcd cluster members
* [talosctl etcd remove-member](#talosctl-etcd-remove-member)	 - Remove the node from etcd cluster
* [talosctl etcd snapshot](#talosctl-etcd-snapshot)	 - Stream snapshot of the etcd node to the path.

## talosctl events

Stream runtime events

```
talosctl events [flags]
```

### Options

```
      --duration duration   show events for the past duration interval (one second resolution, default is to show no history)
  -h, --help                help for events
      --since string        show events after the specified event ID (default is to show no history)
      --tail int32          show specified number of past events (use -1 to show full history, default is to show no history)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl gen ca

Generates a self-signed X.509 certificate authority

```
talosctl gen ca [flags]
```

### Options

```
  -h, --help                  help for ca
      --hours int             the hours from now on which the certificate validity period ends (default 87600)
      --organization string   X.509 distinguished name for the Organization
      --rsa                   generate in RSA format
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen config

Generates a set of configuration files for Talos cluster

### Synopsis

The cluster endpoint is the URL for the Kubernetes API. If you decide to use
a control plane node, common in a single node control plane setup, use port 6443 as
this is the port that the API server binds to on every control plane node. For an HA
setup, usually involving a load balancer, use the IP and port of the load balancer.

```
talosctl gen config <cluster name> <cluster endpoint> [flags]
```

### Options

```
      --additional-sans strings             additional Subject-Alt-Names for the APIServer certificate
      --config-patch string                 patch generated machineconfigs (applied to all node types)
      --config-patch-control-plane string   patch generated machineconfigs (applied to 'init' and 'controlplane' types)
      --config-patch-worker string          patch generated machineconfigs (applied to 'worker' type)
      --dns-domain string                   the dns domain to use for cluster (default "cluster.local")
  -h, --help                                help for config
      --install-disk string                 the disk to install to (default "/dev/sda")
      --install-image string                the image used to perform an installation (default "ghcr.io/talos-systems/installer:latest")
      --kubernetes-version string           desired kubernetes version to run
  -o, --output-dir string                   destination to output generated files
  -p, --persist                             the desired persist value for configs (default true)
      --registry-mirror strings             list of registry mirrors to use in format: <registry host>=<mirror URL>
      --talos-version string                the desired Talos version to generate config for (backwards compatibility, e.g. v0.8)
      --version string                      the desired machine config version to generate (default "v1alpha1")
      --with-docs                           renders all machine configs adding the documentation for each field (default true)
      --with-examples                       renders all machine configs with the commented examples (default true)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen crt

Generates an X.509 Ed25519 certificate

```
talosctl gen crt [flags]
```

### Options

```
      --ca string     path to the PEM encoded CERTIFICATE
      --csr string    path to the PEM encoded CERTIFICATE REQUEST
  -h, --help          help for crt
      --hours int     the hours from now on which the certificate validity period ends (default 24)
      --name string   the basename of the generated file
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen csr

Generates a CSR using an Ed25519 private key

```
talosctl gen csr [flags]
```

### Options

```
  -h, --help            help for csr
      --ip string       generate the certificate for this IP address
      --key string      path to the PEM encoded EC or RSA PRIVATE KEY
      --roles strings   roles (default [os:admin])
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen key

Generates an Ed25519 private key

```
talosctl gen key [flags]
```

### Options

```
  -h, --help          help for key
      --name string   the basename of the generated file
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen keypair

Generates an X.509 Ed25519 key pair

```
talosctl gen keypair [flags]
```

### Options

```
  -h, --help                  help for keypair
      --ip string             generate the certificate for this IP address
      --organization string   X.509 distinguished name for the Organization
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys

## talosctl gen

Generate CAs, certificates, and private keys

### Options

```
  -h, --help   help for gen
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl gen ca](#talosctl-gen-ca)	 - Generates a self-signed X.509 certificate authority
* [talosctl gen config](#talosctl-gen-config)	 - Generates a set of configuration files for Talos cluster
* [talosctl gen crt](#talosctl-gen-crt)	 - Generates an X.509 Ed25519 certificate
* [talosctl gen csr](#talosctl-gen-csr)	 - Generates a CSR using an Ed25519 private key
* [talosctl gen key](#talosctl-gen-key)	 - Generates an Ed25519 private key
* [talosctl gen keypair](#talosctl-gen-keypair)	 - Generates an X.509 Ed25519 key pair

## talosctl get

Get a specific resource or list of resources.

```
talosctl get <type> [<id>] [flags]
```

### Options

```
  -h, --help               help for get
      --namespace string   resource namespace (default is to use default namespace per resource)
  -o, --output string      output mode (table, yaml) (default "table")
  -w, --watch              watch resource changes
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl health

Check cluster health

```
talosctl health [flags]
```

### Options

```
      --control-plane-nodes strings   specify IPs of control plane nodes
  -h, --help                          help for health
      --init-node string              specify IPs of init node
      --k8s-endpoint string           use endpoint instead of kubeconfig default
      --run-e2e                       run Kubernetes e2e test
      --server                        run server-side check (default true)
      --wait-timeout duration         timeout to wait for the cluster to be ready (default 20m0s)
      --worker-nodes strings          specify IPs of worker nodes
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl images

List the default images used by Talos

```
talosctl images [flags]
```

### Options

```
  -h, --help   help for images
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl inspect dependencies

Inspect controller-resource dependencies as graphviz graph.

### Synopsis

Inspect controller-resource dependencies as graphviz graph.

Pipe the output of the command through the "dot" program (part of graphviz package)
to render the graph:

  talosctl inspect dependencies | dot -Tpng > graph.png


```
talosctl inspect dependencies [flags]
```

### Options

```
  -h, --help             help for dependencies
      --with-resources   display live resource information with dependencies
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl inspect](#talosctl-inspect)	 - Inspect internals of Talos

## talosctl inspect

Inspect internals of Talos

### Options

```
  -h, --help   help for inspect
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos
* [talosctl inspect dependencies](#talosctl-inspect-dependencies)	 - Inspect controller-resource dependencies as graphviz graph.

## talosctl interfaces

List network interfaces

```
talosctl interfaces [flags]
```

### Options

```
  -h, --help   help for interfaces
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl kubeconfig

Download the admin kubeconfig from the node

### Synopsis

Download the admin kubeconfig from the node.
If merge flag is defined, config will be merged with ~/.kube/config or [local-path] if specified.
Otherwise kubeconfig will be written to PWD or [local-path] if specified.

```
talosctl kubeconfig [local-path] [flags]
```

### Options

```
  -f, --force                       Force overwrite of kubeconfig if already present, force overwrite on kubeconfig merge
      --force-context-name string   Force context name for kubeconfig merge
  -h, --help                        help for kubeconfig
  -m, --merge                       Merge with existing kubeconfig (default true)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl list

Retrieve a directory listing

```
talosctl list [path] [flags]
```

### Options

```
  -d, --depth int32    maximum recursion depth
  -h, --help           help for list
  -H, --humanize       humanize size and time in the output
  -l, --long           display additional file details
  -r, --recurse        recurse into subdirectories
  -t, --type strings   filter by specified types:
                       f	regular file
                       d	directory
                       l, L	symbolic link
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl logs

Retrieve logs for a service

```
talosctl logs <service name> [flags]
```

### Options

```
  -f, --follow       specify if the logs should be streamed
  -h, --help         help for logs
  -k, --kubernetes   use the k8s.io containerd namespace
      --tail int32   lines of log file to display (default is to show from the beginning) (default -1)
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl memory

Show memory usage

```
talosctl memory [flags]
```

### Options

```
  -h, --help      help for memory
  -v, --verbose   display extended memory statistics
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl mounts

List mounts

```
talosctl mounts [flags]
```

### Options

```
  -h, --help   help for mounts
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl patch

Update field(s) of a resource using a JSON patch.

```
talosctl patch <type> [<id>] [flags]
```

### Options

```
  -h, --help                help for patch
      --immediate           apply the change immediately (without a reboot)
      --namespace string    resource namespace (default is to use default namespace per resource)
      --on-reboot           apply the change on next reboot
  -p, --patch string        the patch to be applied to the resource file.
      --patch-file string   a file containing a patch to be applied to the resource.
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl processes

List running processes

```
talosctl processes [flags]
```

### Options

```
  -h, --help          help for processes
  -s, --sort string   Column to sort output by. [rss|cpu] (default "rss")
  -w, --watch         Stream running processes
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl read

Read a file on the machine

```
talosctl read <path> [flags]
```

### Options

```
  -h, --help   help for read
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl reboot

Reboot a node

```
talosctl reboot [flags]
```

### Options

```
  -h, --help   help for reboot
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl reset

Reset a node

```
talosctl reset [flags]
```

### Options

```
      --graceful                        if true, attempt to cordon/drain node and leave etcd (if applicable) (default true)
  -h, --help                            help for reset
      --reboot                          if true, reboot the node after resetting instead of shutting down
      --system-labels-to-wipe strings   if set, just wipe selected system disk partitions by label but keep other partitions intact
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl restart

Restart a process

```
talosctl restart <id> [flags]
```

### Options

```
  -h, --help         help for restart
  -k, --kubernetes   use the k8s.io containerd namespace
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl rollback

Rollback a node to the previous installation

```
talosctl rollback [flags]
```

### Options

```
  -h, --help   help for rollback
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl routes

List network routes

```
talosctl routes [flags]
```

### Options

```
  -h, --help   help for routes
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl service

Retrieve the state of a service (or all services), control service state

### Synopsis

Service control command. If run without arguments, lists all the services and their state.
If service ID is specified, default action 'status' is executed which shows status of a single list service.
With actions 'start', 'stop', 'restart', service state is updated respectively.

```
talosctl service [<id> [start|stop|restart|status]] [flags]
```

### Options

```
  -h, --help   help for service
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl shutdown

Shutdown a node

```
talosctl shutdown [flags]
```

### Options

```
  -h, --help   help for shutdown
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl stats

Get container stats

```
talosctl stats [flags]
```

### Options

```
  -h, --help         help for stats
  -k, --kubernetes   use the k8s.io containerd namespace
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl time

Gets current server time

```
talosctl time [--check server] [flags]
```

### Options

```
  -c, --check string   checks server time against specified ntp server
  -h, --help           help for time
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl upgrade

Upgrade Talos on the target node

```
talosctl upgrade [flags]
```

### Options

```
  -f, --force          force the upgrade (skip checks on etcd health and members, might lead to data loss)
  -h, --help           help for upgrade
  -i, --image string   the container image to use for performing the install
  -p, --preserve       preserve data
  -s, --stage          stage the upgrade to perform it after a reboot
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl upgrade-k8s

Upgrade Kubernetes control plane in the Talos cluster.

### Synopsis

Command runs upgrade of Kubernetes control plane components between specified versions. Pod-checkpointer is handled in a special way to speed up kube-apisever upgrades.

```
talosctl upgrade-k8s [flags]
```

### Options

```
      --endpoint string   the cluster control plane endpoint
      --from string       the Kubernetes control plane version to upgrade from
  -h, --help              help for upgrade-k8s
      --to string         the Kubernetes control plane version to upgrade to (default "1.22.0-beta.2")
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl usage

Retrieve a disk usage

```
talosctl usage [path1] [path2] ... [pathN] [flags]
```

### Options

```
  -a, --all             write counts for all files, not just directories
  -d, --depth int32     maximum recursion depth
  -h, --help            help for usage
  -H, --humanize        humanize size and time in the output
  -t, --threshold int   threshold exclude entries smaller than SIZE if positive, or entries greater than SIZE if negative
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl validate

Validate config

```
talosctl validate [flags]
```

### Options

```
  -c, --config string   the path of the config file
  -h, --help            help for validate
  -m, --mode string     the mode to validate the config for (valid values are metal, cloud, and container)
      --strict          treat validation warnings as errors
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl version

Prints the version

```
talosctl version [flags]
```

### Options

```
      --client   Print client version only
  -h, --help     help for version
      --short    Print the short version
```

### Options inherited from parent commands

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl](#talosctl)	 - A CLI for out-of-band management of Kubernetes nodes created by Talos

## talosctl

A CLI for out-of-band management of Kubernetes nodes created by Talos

### Options

```
      --context string       Context to be used in command
  -e, --endpoints strings    override default endpoints in Talos configuration
  -h, --help                 help for talosctl
  -n, --nodes strings        target the specified nodes
      --talosconfig string   The path to the Talos configuration file (default "/home/user/.talos/config")
```

### SEE ALSO

* [talosctl apply-config](#talosctl-apply-config)	 - Apply a new configuration to a node
* [talosctl bootstrap](#talosctl-bootstrap)	 - Bootstrap the etcd cluster on the specified node.
* [talosctl cluster](#talosctl-cluster)	 - A collection of commands for managing local docker-based or firecracker-based clusters
* [talosctl completion](#talosctl-completion)	 - Output shell completion code for the specified shell (bash or zsh)
* [talosctl config](#talosctl-config)	 - Manage the client configuration file (talosconfig)
* [talosctl conformance](#talosctl-conformance)	 - Run conformance tests
* [talosctl containers](#talosctl-containers)	 - List containers
* [talosctl convert-k8s](#talosctl-convert-k8s)	 - Convert Kubernetes control plane from self-hosted (bootkube) to Talos-managed (static pods).
* [talosctl copy](#talosctl-copy)	 - Copy data out from the node
* [talosctl crashdump](#talosctl-crashdump)	 - Dump debug information about the cluster
* [talosctl dashboard](#talosctl-dashboard)	 - Cluster dashboard with real-time metrics
* [talosctl disks](#talosctl-disks)	 - Get the list of disks from /sys/block on the machine
* [talosctl dmesg](#talosctl-dmesg)	 - Retrieve kernel logs
* [talosctl edit](#talosctl-edit)	 - Edit a resource from the default editor.
* [talosctl etcd](#talosctl-etcd)	 - Manage etcd
* [talosctl events](#talosctl-events)	 - Stream runtime events
* [talosctl gen](#talosctl-gen)	 - Generate CAs, certificates, and private keys
* [talosctl get](#talosctl-get)	 - Get a specific resource or list of resources.
* [talosctl health](#talosctl-health)	 - Check cluster health
* [talosctl images](#talosctl-images)	 - List the default images used by Talos
* [talosctl inspect](#talosctl-inspect)	 - Inspect internals of Talos
* [talosctl interfaces](#talosctl-interfaces)	 - List network interfaces
* [talosctl kubeconfig](#talosctl-kubeconfig)	 - Download the admin kubeconfig from the node
* [talosctl list](#talosctl-list)	 - Retrieve a directory listing
* [talosctl logs](#talosctl-logs)	 - Retrieve logs for a service
* [talosctl memory](#talosctl-memory)	 - Show memory usage
* [talosctl mounts](#talosctl-mounts)	 - List mounts
* [talosctl patch](#talosctl-patch)	 - Update field(s) of a resource using a JSON patch.
* [talosctl processes](#talosctl-processes)	 - List running processes
* [talosctl read](#talosctl-read)	 - Read a file on the machine
* [talosctl reboot](#talosctl-reboot)	 - Reboot a node
* [talosctl reset](#talosctl-reset)	 - Reset a node
* [talosctl restart](#talosctl-restart)	 - Restart a process
* [talosctl rollback](#talosctl-rollback)	 - Rollback a node to the previous installation
* [talosctl routes](#talosctl-routes)	 - List network routes
* [talosctl service](#talosctl-service)	 - Retrieve the state of a service (or all services), control service state
* [talosctl shutdown](#talosctl-shutdown)	 - Shutdown a node
* [talosctl stats](#talosctl-stats)	 - Get container stats
* [talosctl time](#talosctl-time)	 - Gets current server time
* [talosctl upgrade](#talosctl-upgrade)	 - Upgrade Talos on the target node
* [talosctl upgrade-k8s](#talosctl-upgrade-k8s)	 - Upgrade Kubernetes control plane in the Talos cluster.
* [talosctl usage](#talosctl-usage)	 - Retrieve a disk usage
* [talosctl validate](#talosctl-validate)	 - Validate config
* [talosctl version](#talosctl-version)	 - Prints the version

