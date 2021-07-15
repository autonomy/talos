---
title: Configuration
desription: Talos node configuration file reference.
---

<!-- markdownlint-disable -->




Package v1alpha1 configuration file contains all the options available for configuring a machine.

To generate a set of basic configuration files, run:

	talosctl gen config --version v1alpha1 <cluster name> <cluster endpoint>

This will generate a machine config for each node type, and a talosconfig for the CLI.

## Config
Config defines the v1alpha1 configuration file.



``` yaml
version: v1alpha1
persist: true
machine: # ...
cluster: # ...
```

<hr />

<div class="dd">

<code>version</code>  <i>string</i>

</div>
<div class="dt">

Indicates the schema used to decode the contents.


Valid values:


  - <code>v1alpha1</code>
</div>

<hr />

<div class="dd">

<code>debug</code>  <i>bool</i>

</div>
<div class="dt">

Enable verbose logging to the console.
All system containers logs will flow into serial console.

> Note: To avoid breaking Talos bootstrap flow enable this option only if serial console can handle high message throughput.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />

<div class="dd">

<code>persist</code>  <i>bool</i>

</div>
<div class="dt">

Indicates whether to pull the machine config upon every boot.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />

<div class="dd">

<code>machine</code>  <i><a href="#machineconfig">MachineConfig</a></i>

</div>
<div class="dt">

Provides machine specific configuration options.

</div>

<hr />

<div class="dd">

<code>cluster</code>  <i><a href="#clusterconfig">ClusterConfig</a></i>

</div>
<div class="dt">

Provides cluster specific configuration options.

</div>

<hr />





## MachineConfig
MachineConfig represents the machine-specific config values.

Appears in:


- <code><a href="#config">Config</a>.machine</code>


``` yaml
type: controlplane
# InstallConfig represents the installation options for preparing a node.
install:
    disk: /dev/sda # The disk used for installations.
    # Allows for supplying extra kernel args via the bootloader.
    extraKernelArgs:
        - console=ttyS1
        - panic=10
    image: ghcr.io/talos-systems/installer:latest # Allows for supplying the image used to perform the installation.
    bootloader: true # Indicates if a bootloader should be installed.
    wipe: false # Indicates if the installation disk should be wiped at installation time.

    # # Look up disk using disk characteristics like model, size, serial and others.
    # diskSelector:
    #     size: 4GB # Disk size.
    #     model: WDC* # Disk model `/sys/block/<dev>/device/model`.
```

<hr />

<div class="dd">

<code>type</code>  <i>string</i>

</div>
<div class="dt">

Defines the role of the machine within the cluster.

#### Init

Init node type designates the first control plane node to come up.
You can think of it like a bootstrap node.
This node will perform the initial steps to bootstrap the cluster -- generation of TLS assets, starting of the control plane, etc.

#### Control Plane

Control Plane node type designates the node as a control plane member.
This means it will host etcd along with the Kubernetes master components such as API Server, Controller Manager, Scheduler.

#### Worker

Worker node type designates the node as a worker node.
This means it will be an available compute node for scheduling workloads.

This node type was previously known as "join"; that value is still supported but deprecated.


Valid values:


  - <code>init</code>

  - <code>controlplane</code>

  - <code>worker</code>
</div>

<hr />

<div class="dd">

<code>token</code>  <i>string</i>

</div>
<div class="dt">

The `token` is used by a machine to join the PKI of the cluster.
Using this token, a machine will create a certificate signing request (CSR), and request a certificate that will be used as its' identity.


> Warning: It is important to ensure that this token is correct since a machine's certificate has a short TTL by default.



Examples:


``` yaml
token: 328hom.uqjzh6jnn2eie9oi
```


</div>

<hr />

<div class="dd">

<code>ca</code>  <i>PEMEncodedCertificateAndKey</i>

</div>
<div class="dt">

The root certificate authority of the PKI.
It is composed of a base64 encoded `crt` and `key`.



Examples:


``` yaml
ca:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>certSANs</code>  <i>[]string</i>

</div>
<div class="dt">

Extra certificate subject alternative names for the machine's certificate.
By default, all non-loopback interface IPs are automatically added to the certificate's SANs.



Examples:


``` yaml
certSANs:
    - 10.0.0.10
    - 172.16.0.10
    - 192.168.0.10
```


</div>

<hr />

<div class="dd">

<code>kubelet</code>  <i><a href="#kubeletconfig">KubeletConfig</a></i>

</div>
<div class="dt">

Used to provide additional options to the kubelet.



Examples:


``` yaml
kubelet:
    image: ghcr.io/talos-systems/kubelet:v1.22.0-beta.2 # The `image` field is an optional reference to an alternative kubelet image.
    # The `extraArgs` field is used to provide additional flags to the kubelet.
    extraArgs:
        feature-gates: ServerSideApply=true

    # # The `ClusterDNS` field is an optional reference to an alternative kubelet clusterDNS ip list.
    # clusterDNS:
    #     - 10.96.0.10
    #     - 169.254.2.53

    # # The `extraMounts` field is used to add additional mounts to the kubelet container.
    # extraMounts:
    #     - destination: /var/lib/example
    #       type: bind
    #       source: /var/lib/example
    #       options:
    #         - rshared
    #         - rw
```


</div>

<hr />

<div class="dd">

<code>network</code>  <i><a href="#networkconfig">NetworkConfig</a></i>

</div>
<div class="dt">

Provides machine specific network configuration options.



Examples:


``` yaml
network:
    hostname: worker-1 # Used to statically set the hostname for the machine.
    # `interfaces` is used to define the network interface configuration.
    interfaces:
        - interface: eth0 # The interface name.
          cidr: 192.168.2.0/24 # Assigns a static IP address to the interface.
          # A list of routes associated with the interface.
          routes:
            - network: 0.0.0.0/0 # The route's network.
              gateway: 192.168.2.1 # The route's gateway.
              metric: 1024 # The optional metric for the route.
          mtu: 1500 # The interface's MTU.

          # # Bond specific options.
          # bond:
          #     # The interfaces that make up the bond.
          #     interfaces:
          #         - eth0
          #         - eth1
          #     mode: 802.3ad # A bond option.
          #     lacpRate: fast # A bond option.

          # # Indicates if DHCP should be used to configure the interface.
          # dhcp: true

          # # DHCP specific options.
          # dhcpOptions:
          #     routeMetric: 1024 # The priority of all routes received via DHCP.

          # # Wireguard specific configuration.

          # # wireguard server example
          # wireguard:
          #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
          #     listenPort: 51111 # Specifies a device's listening port.
          #     # Specifies a list of peer configurations to apply to a device.
          #     peers:
          #         - publicKey: ABCDEF... # Specifies the public key of this peer.
          #           endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
          #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
          #           allowedIPs:
          #             - 192.168.1.0/24
          # # wireguard peer example
          # wireguard:
          #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
          #     # Specifies a list of peer configurations to apply to a device.
          #     peers:
          #         - publicKey: ABCDEF... # Specifies the public key of this peer.
          #           endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
          #           persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
          #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
          #           allowedIPs:
          #             - 192.168.1.0/24

          # # Virtual (shared) IP address configuration.
          # vip:
          #     ip: 172.16.199.55 # Specifies the IP address to be used.
    # Used to statically set the nameservers for the machine.
    nameservers:
        - 9.8.7.6
        - 8.7.6.5

    # # Allows for extra entries to be added to the `/etc/hosts` file
    # extraHostEntries:
    #     - ip: 192.168.1.100 # The IP of the host.
    #       # The host alias.
    #       aliases:
    #         - example
    #         - example.domain.tld
```


</div>

<hr />

<div class="dd">

<code>disks</code>  <i>[]<a href="#machinedisk">MachineDisk</a></i>

</div>
<div class="dt">

Used to partition, format and mount additional disks.
Since the rootfs is read only with the exception of `/var`, mounts are only valid if they are under `/var`.
Note that the partitioning and formating is done only once, if and only if no existing partitions are found.
If `size:` is omitted, the partition is sized to occupy the full disk.


> Note: `size` is in units of bytes.



Examples:


``` yaml
disks:
    - device: /dev/sdb # The name of the disk to use.
      # A list of partitions to create on the disk.
      partitions:
        - mountpoint: /var/mnt/extra # Where to mount the partition.

          # # The size of partition: either bytes or human readable representation. If `size:` is omitted, the partition is sized to occupy the full disk.

          # # Human readable representation.
          # size: 100 MB
          # # Precise value in bytes.
          # size: 1073741824
```


</div>

<hr />

<div class="dd">

<code>install</code>  <i><a href="#installconfig">InstallConfig</a></i>

</div>
<div class="dt">

Used to provide instructions for installations.



Examples:


``` yaml
install:
    disk: /dev/sda # The disk used for installations.
    # Allows for supplying extra kernel args via the bootloader.
    extraKernelArgs:
        - console=ttyS1
        - panic=10
    image: ghcr.io/talos-systems/installer:latest # Allows for supplying the image used to perform the installation.
    bootloader: true # Indicates if a bootloader should be installed.
    wipe: false # Indicates if the installation disk should be wiped at installation time.

    # # Look up disk using disk characteristics like model, size, serial and others.
    # diskSelector:
    #     size: 4GB # Disk size.
    #     model: WDC* # Disk model `/sys/block/<dev>/device/model`.
```


</div>

<hr />

<div class="dd">

<code>files</code>  <i>[]<a href="#machinefile">MachineFile</a></i>

</div>
<div class="dt">

Allows the addition of user specified files.
The value of `op` can be `create`, `overwrite`, or `append`.
In the case of `create`, `path` must not exist.
In the case of `overwrite`, and `append`, `path` must be a valid file.
If an `op` value of `append` is used, the existing file will be appended.
Note that the file contents are not required to be base64 encoded.


> Note: The specified `path` is relative to `/var`.



Examples:


``` yaml
files:
    - content: '...' # The contents of the file.
      permissions: 0o666 # The file's permissions in octal.
      path: /tmp/file.txt # The path of the file.
      op: append # The operation to use
```


</div>

<hr />

<div class="dd">

<code>env</code>  <i>Env</i>

</div>
<div class="dt">

The `env` field allows for the addition of environment variables.
All environment variables are set on PID 1 in addition to every service.


Valid values:


  - <code>`GRPC_GO_LOG_VERBOSITY_LEVEL`</code>

  - <code>`GRPC_GO_LOG_SEVERITY_LEVEL`</code>

  - <code>`http_proxy`</code>

  - <code>`https_proxy`</code>

  - <code>`no_proxy`</code>


Examples:


``` yaml
env:
    GRPC_GO_LOG_SEVERITY_LEVEL: info
    GRPC_GO_LOG_VERBOSITY_LEVEL: "99"
    https_proxy: http://SERVER:PORT/
```

``` yaml
env:
    GRPC_GO_LOG_SEVERITY_LEVEL: error
    https_proxy: https://USERNAME:PASSWORD@SERVER:PORT/
```

``` yaml
env:
    https_proxy: http://DOMAIN\USERNAME:PASSWORD@SERVER:PORT/
```


</div>

<hr />

<div class="dd">

<code>time</code>  <i><a href="#timeconfig">TimeConfig</a></i>

</div>
<div class="dt">

Used to configure the machine's time settings.



Examples:


``` yaml
time:
    disabled: false # Indicates if the time service is disabled for the machine.
    # Specifies time (NTP) servers to use for setting the system time.
    servers:
        - time.cloudflare.com
```


</div>

<hr />

<div class="dd">

<code>sysctls</code>  <i>map[string]string</i>

</div>
<div class="dt">

Used to configure the machine's sysctls.



Examples:


``` yaml
sysctls:
    kernel.domainname: talos.dev
    net.ipv4.ip_forward: "0"
```


</div>

<hr />

<div class="dd">

<code>registries</code>  <i><a href="#registriesconfig">RegistriesConfig</a></i>

</div>
<div class="dt">

Used to configure the machine's container image registry mirrors.

Automatically generates matching CRI configuration for registry mirrors.

The `mirrors` section allows to redirect requests for images to non-default registry,
which might be local registry or caching mirror.

The `config` section provides a way to authenticate to the registry with TLS client
identity, provide registry CA, or authentication information.
Authentication information has same meaning with the corresponding field in `.docker/config.json`.

See also matching configuration for [CRI containerd plugin](https://github.com/containerd/cri/blob/master/docs/registry.md).



Examples:


``` yaml
registries:
    # Specifies mirror configuration for each registry.
    mirrors:
        docker.io:
            # List of endpoints (URLs) for registry mirrors to use.
            endpoints:
                - https://registry.local
    # Specifies TLS & auth configuration for HTTPS image registries.
    config:
        registry.local:
            # The TLS configuration for the registry.
            tls:
                # Enable mutual TLS authentication with the registry.
                clientIdentity:
                    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
                    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
            # The auth configuration for this registry.
            auth:
                username: username # Optional registry authentication.
                password: password # Optional registry authentication.
```


</div>

<hr />

<div class="dd">

<code>systemDiskEncryption</code>  <i><a href="#systemdiskencryptionconfig">SystemDiskEncryptionConfig</a></i>

</div>
<div class="dt">

Machine system disk encryption configuration.
Defines each system partition encryption parameters.



Examples:


``` yaml
systemDiskEncryption:
    # Ephemeral partition encryption.
    ephemeral:
        provider: luks2 # Encryption provider to use for the encryption.
        # Defines the encryption keys generation and storage method.
        keys:
            - # Deterministically generated key from the node UUID and PartitionLabel.
              nodeID: {}
              slot: 0 # Key slot number for luks2 encryption.
```


</div>

<hr />

<div class="dd">

<code>features</code>  <i><a href="#featuresconfig">FeaturesConfig</a></i>

</div>
<div class="dt">

Features describe individual Talos features that can be switched on or off.



Examples:


``` yaml
features:
    rbac: true # Enable role-based access control (RBAC).
```


</div>

<hr />





## ClusterConfig
ClusterConfig represents the cluster-wide config values.

Appears in:


- <code><a href="#config">Config</a>.cluster</code>


``` yaml
# ControlPlaneConfig represents the control plane configuration options.
controlPlane:
    endpoint: https://1.2.3.4 # Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
    localAPIServerPort: 443 # The port that the API server listens on internally.
clusterName: talos.local
# ClusterNetworkConfig represents kube networking configuration options.
network:
    # The CNI used.
    cni:
        name: flannel # Name of CNI to use.
    dnsDomain: cluster.local # The domain used by Kubernetes DNS.
    # The pod subnet CIDR.
    podSubnets:
        - 10.244.0.0/16
    # The service subnet CIDR.
    serviceSubnets:
        - 10.96.0.0/12
```

<hr />

<div class="dd">

<code>controlPlane</code>  <i><a href="#controlplaneconfig">ControlPlaneConfig</a></i>

</div>
<div class="dt">

Provides control plane specific configuration options.



Examples:


``` yaml
controlPlane:
    endpoint: https://1.2.3.4 # Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
    localAPIServerPort: 443 # The port that the API server listens on internally.
```


</div>

<hr />

<div class="dd">

<code>clusterName</code>  <i>string</i>

</div>
<div class="dt">

Configures the cluster's name.

</div>

<hr />

<div class="dd">

<code>network</code>  <i><a href="#clusternetworkconfig">ClusterNetworkConfig</a></i>

</div>
<div class="dt">

Provides cluster specific network configuration options.



Examples:


``` yaml
network:
    # The CNI used.
    cni:
        name: flannel # Name of CNI to use.
    dnsDomain: cluster.local # The domain used by Kubernetes DNS.
    # The pod subnet CIDR.
    podSubnets:
        - 10.244.0.0/16
    # The service subnet CIDR.
    serviceSubnets:
        - 10.96.0.0/12
```


</div>

<hr />

<div class="dd">

<code>token</code>  <i>string</i>

</div>
<div class="dt">

The [bootstrap token](https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/) used to join the cluster.



Examples:


``` yaml
token: wlzjyw.bei2zfylhs2by0wd
```


</div>

<hr />

<div class="dd">

<code>aescbcEncryptionSecret</code>  <i>string</i>

</div>
<div class="dt">

The key used for the [encryption of secret data at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/).



Examples:


``` yaml
aescbcEncryptionSecret: z01mye6j16bspJYtTB/5SFX8j7Ph4JXxM2Xuu4vsBPM=
```


</div>

<hr />

<div class="dd">

<code>ca</code>  <i>PEMEncodedCertificateAndKey</i>

</div>
<div class="dt">

The base64 encoded root certificate authority used by Kubernetes.



Examples:


``` yaml
ca:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>aggregatorCA</code>  <i>PEMEncodedCertificateAndKey</i>

</div>
<div class="dt">

The base64 encoded aggregator certificate authority used by Kubernetes for front-proxy certificate generation.

This CA can be self-signed.



Examples:


``` yaml
aggregatorCA:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>serviceAccount</code>  <i>PEMEncodedKey</i>

</div>
<div class="dt">

The base64 encoded private key for service account token generation.



Examples:


``` yaml
serviceAccount:
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>apiServer</code>  <i><a href="#apiserverconfig">APIServerConfig</a></i>

</div>
<div class="dt">

API server specific configuration options.



Examples:


``` yaml
apiServer:
    image: k8s.gcr.io/kube-apiserver:v1.22.0-beta.2 # The container image used in the API server manifest.
    # Extra arguments to supply to the API server.
    extraArgs:
        feature-gates: ServerSideApply=true
        http2-max-streams-per-connection: "32"
    # Extra certificate subject alternative names for the API server's certificate.
    certSANs:
        - 1.2.3.4
        - 4.5.6.7
```


</div>

<hr />

<div class="dd">

<code>controllerManager</code>  <i><a href="#controllermanagerconfig">ControllerManagerConfig</a></i>

</div>
<div class="dt">

Controller manager server specific configuration options.



Examples:


``` yaml
controllerManager:
    image: k8s.gcr.io/kube-controller-manager:v1.22.0-beta.2 # The container image used in the controller manager manifest.
    # Extra arguments to supply to the controller manager.
    extraArgs:
        feature-gates: ServerSideApply=true
```


</div>

<hr />

<div class="dd">

<code>proxy</code>  <i><a href="#proxyconfig">ProxyConfig</a></i>

</div>
<div class="dt">

Kube-proxy server-specific configuration options



Examples:


``` yaml
proxy:
    image: k8s.gcr.io/kube-proxy:v1.22.0-beta.2 # The container image used in the kube-proxy manifest.
    mode: ipvs # proxy mode of kube-proxy.
    # Extra arguments to supply to kube-proxy.
    extraArgs:
        proxy-mode: iptables
```


</div>

<hr />

<div class="dd">

<code>scheduler</code>  <i><a href="#schedulerconfig">SchedulerConfig</a></i>

</div>
<div class="dt">

Scheduler server specific configuration options.



Examples:


``` yaml
scheduler:
    image: k8s.gcr.io/kube-scheduler:v1.22.0-beta.2 # The container image used in the scheduler manifest.
    # Extra arguments to supply to the scheduler.
    extraArgs:
        feature-gates: AllBeta=true
```


</div>

<hr />

<div class="dd">

<code>etcd</code>  <i><a href="#etcdconfig">EtcdConfig</a></i>

</div>
<div class="dt">

Etcd specific configuration options.



Examples:


``` yaml
etcd:
    image: gcr.io/etcd-development/etcd:v3.4.16 # The container image used to create the etcd service.
    # The `ca` is the root certificate authority of the PKI.
    ca:
        crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
        key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
    # Extra arguments to supply to etcd.
    extraArgs:
        election-timeout: "5000"
```


</div>

<hr />

<div class="dd">

<code>coreDNS</code>  <i><a href="#coredns">CoreDNS</a></i>

</div>
<div class="dt">

Core DNS specific configuration options.



Examples:


``` yaml
coreDNS:
    image: docker.io/coredns/coredns:1.8.4 # The `image` field is an override to the default coredns image.
```


</div>

<hr />

<div class="dd">

<code>externalCloudProvider</code>  <i><a href="#externalcloudproviderconfig">ExternalCloudProviderConfig</a></i>

</div>
<div class="dt">

External cloud provider configuration.



Examples:


``` yaml
externalCloudProvider:
    enabled: true # Enable external cloud provider.
    # A list of urls that point to additional manifests for an external cloud provider.
    manifests:
        - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml
        - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml
```


</div>

<hr />

<div class="dd">

<code>extraManifests</code>  <i>[]string</i>

</div>
<div class="dt">

A list of urls that point to additional manifests.
These will get automatically deployed as part of the bootstrap.



Examples:


``` yaml
extraManifests:
    - https://www.example.com/manifest1.yaml
    - https://www.example.com/manifest2.yaml
```


</div>

<hr />

<div class="dd">

<code>extraManifestHeaders</code>  <i>map[string]string</i>

</div>
<div class="dt">

A map of key value pairs that will be added while fetching the extraManifests.



Examples:


``` yaml
extraManifestHeaders:
    Token: "1234567"
    X-ExtraInfo: info
```


</div>

<hr />

<div class="dd">

<code>inlineManifests</code>  <i>ClusterInlineManifests</i>

</div>
<div class="dt">

A list of inline Kubernetes manifests.
These will get automatically deployed as part of the bootstrap.



Examples:


``` yaml
inlineManifests:
    - name: namespace-ci # Name of the manifest.
      contents: |- # Manifest contents as a string.
        apiVersion: v1
        kind: Namespace
        metadata:
        	name: ci
```


</div>

<hr />

<div class="dd">

<code>adminKubeconfig</code>  <i><a href="#adminkubeconfigconfig">AdminKubeconfigConfig</a></i>

</div>
<div class="dt">

Settings for admin kubeconfig generation.
Certificate lifetime can be configured.



Examples:


``` yaml
adminKubeconfig:
    certLifetime: 1h0m0s # Admin kubeconfig certificate lifetime (default is 1 year).
```


</div>

<hr />

<div class="dd">

<code>allowSchedulingOnMasters</code>  <i>bool</i>

</div>
<div class="dt">

Allows running workload on master nodes.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />





## ExtraMount
ExtraMount wraps OCI Mount specification.

Appears in:


- <code><a href="#kubeletconfig">KubeletConfig</a>.extraMounts</code>


``` yaml
- destination: /var/lib/example
  type: bind
  source: /var/lib/example
  options:
    - rshared
    - rw
```



## KubeletConfig
KubeletConfig represents the kubelet config values.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.kubelet</code>


``` yaml
image: ghcr.io/talos-systems/kubelet:v1.22.0-beta.2 # The `image` field is an optional reference to an alternative kubelet image.
# The `extraArgs` field is used to provide additional flags to the kubelet.
extraArgs:
    feature-gates: ServerSideApply=true

# # The `ClusterDNS` field is an optional reference to an alternative kubelet clusterDNS ip list.
# clusterDNS:
#     - 10.96.0.10
#     - 169.254.2.53

# # The `extraMounts` field is used to add additional mounts to the kubelet container.
# extraMounts:
#     - destination: /var/lib/example
#       type: bind
#       source: /var/lib/example
#       options:
#         - rshared
#         - rw
```

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The `image` field is an optional reference to an alternative kubelet image.



Examples:


``` yaml
image: ghcr.io/talos-systems/kubelet:v1.22.0-beta.2
```


</div>

<hr />

<div class="dd">

<code>clusterDNS</code>  <i>[]string</i>

</div>
<div class="dt">

The `ClusterDNS` field is an optional reference to an alternative kubelet clusterDNS ip list.



Examples:


``` yaml
clusterDNS:
    - 10.96.0.10
    - 169.254.2.53
```


</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

The `extraArgs` field is used to provide additional flags to the kubelet.



Examples:


``` yaml
extraArgs:
    key: value
```


</div>

<hr />

<div class="dd">

<code>extraMounts</code>  <i>[]<a href="#extramount">ExtraMount</a></i>

</div>
<div class="dt">

The `extraMounts` field is used to add additional mounts to the kubelet container.



Examples:


``` yaml
extraMounts:
    - destination: /var/lib/example
      type: bind
      source: /var/lib/example
      options:
        - rshared
        - rw
```


</div>

<hr />

<div class="dd">

<code>registerWithFQDN</code>  <i>bool</i>

</div>
<div class="dt">

The `registerWithFQDN` field is used to force kubelet to use the node FQDN for registration.
This is required in clouds like AWS.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />





## NetworkConfig
NetworkConfig represents the machine's networking config values.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.network</code>


``` yaml
hostname: worker-1 # Used to statically set the hostname for the machine.
# `interfaces` is used to define the network interface configuration.
interfaces:
    - interface: eth0 # The interface name.
      cidr: 192.168.2.0/24 # Assigns a static IP address to the interface.
      # A list of routes associated with the interface.
      routes:
        - network: 0.0.0.0/0 # The route's network.
          gateway: 192.168.2.1 # The route's gateway.
          metric: 1024 # The optional metric for the route.
      mtu: 1500 # The interface's MTU.

      # # Bond specific options.
      # bond:
      #     # The interfaces that make up the bond.
      #     interfaces:
      #         - eth0
      #         - eth1
      #     mode: 802.3ad # A bond option.
      #     lacpRate: fast # A bond option.

      # # Indicates if DHCP should be used to configure the interface.
      # dhcp: true

      # # DHCP specific options.
      # dhcpOptions:
      #     routeMetric: 1024 # The priority of all routes received via DHCP.

      # # Wireguard specific configuration.

      # # wireguard server example
      # wireguard:
      #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
      #     listenPort: 51111 # Specifies a device's listening port.
      #     # Specifies a list of peer configurations to apply to a device.
      #     peers:
      #         - publicKey: ABCDEF... # Specifies the public key of this peer.
      #           endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
      #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      #           allowedIPs:
      #             - 192.168.1.0/24
      # # wireguard peer example
      # wireguard:
      #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
      #     # Specifies a list of peer configurations to apply to a device.
      #     peers:
      #         - publicKey: ABCDEF... # Specifies the public key of this peer.
      #           endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
      #           persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
      #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      #           allowedIPs:
      #             - 192.168.1.0/24

      # # Virtual (shared) IP address configuration.
      # vip:
      #     ip: 172.16.199.55 # Specifies the IP address to be used.
# Used to statically set the nameservers for the machine.
nameservers:
    - 9.8.7.6
    - 8.7.6.5

# # Allows for extra entries to be added to the `/etc/hosts` file
# extraHostEntries:
#     - ip: 192.168.1.100 # The IP of the host.
#       # The host alias.
#       aliases:
#         - example
#         - example.domain.tld
```

<hr />

<div class="dd">

<code>hostname</code>  <i>string</i>

</div>
<div class="dt">

Used to statically set the hostname for the machine.

</div>

<hr />

<div class="dd">

<code>interfaces</code>  <i>[]<a href="#device">Device</a></i>

</div>
<div class="dt">

`interfaces` is used to define the network interface configuration.
By default all network interfaces will attempt a DHCP discovery.
This can be further tuned through this configuration parameter.



Examples:


``` yaml
interfaces:
    - interface: eth0 # The interface name.
      cidr: 192.168.2.0/24 # Assigns a static IP address to the interface.
      # A list of routes associated with the interface.
      routes:
        - network: 0.0.0.0/0 # The route's network.
          gateway: 192.168.2.1 # The route's gateway.
          metric: 1024 # The optional metric for the route.
      mtu: 1500 # The interface's MTU.

      # # Bond specific options.
      # bond:
      #     # The interfaces that make up the bond.
      #     interfaces:
      #         - eth0
      #         - eth1
      #     mode: 802.3ad # A bond option.
      #     lacpRate: fast # A bond option.

      # # Indicates if DHCP should be used to configure the interface.
      # dhcp: true

      # # DHCP specific options.
      # dhcpOptions:
      #     routeMetric: 1024 # The priority of all routes received via DHCP.

      # # Wireguard specific configuration.

      # # wireguard server example
      # wireguard:
      #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
      #     listenPort: 51111 # Specifies a device's listening port.
      #     # Specifies a list of peer configurations to apply to a device.
      #     peers:
      #         - publicKey: ABCDEF... # Specifies the public key of this peer.
      #           endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
      #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      #           allowedIPs:
      #             - 192.168.1.0/24
      # # wireguard peer example
      # wireguard:
      #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
      #     # Specifies a list of peer configurations to apply to a device.
      #     peers:
      #         - publicKey: ABCDEF... # Specifies the public key of this peer.
      #           endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
      #           persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
      #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      #           allowedIPs:
      #             - 192.168.1.0/24

      # # Virtual (shared) IP address configuration.
      # vip:
      #     ip: 172.16.199.55 # Specifies the IP address to be used.
```


</div>

<hr />

<div class="dd">

<code>nameservers</code>  <i>[]string</i>

</div>
<div class="dt">

Used to statically set the nameservers for the machine.
Defaults to `1.1.1.1` and `8.8.8.8`



Examples:


``` yaml
nameservers:
    - 8.8.8.8
    - 1.1.1.1
```


</div>

<hr />

<div class="dd">

<code>extraHostEntries</code>  <i>[]<a href="#extrahost">ExtraHost</a></i>

</div>
<div class="dt">

Allows for extra entries to be added to the `/etc/hosts` file



Examples:


``` yaml
extraHostEntries:
    - ip: 192.168.1.100 # The IP of the host.
      # The host alias.
      aliases:
        - example
        - example.domain.tld
```


</div>

<hr />





## InstallConfig
InstallConfig represents the installation options for preparing a node.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.install</code>


``` yaml
disk: /dev/sda # The disk used for installations.
# Allows for supplying extra kernel args via the bootloader.
extraKernelArgs:
    - console=ttyS1
    - panic=10
image: ghcr.io/talos-systems/installer:latest # Allows for supplying the image used to perform the installation.
bootloader: true # Indicates if a bootloader should be installed.
wipe: false # Indicates if the installation disk should be wiped at installation time.

# # Look up disk using disk characteristics like model, size, serial and others.
# diskSelector:
#     size: 4GB # Disk size.
#     model: WDC* # Disk model `/sys/block/<dev>/device/model`.
```

<hr />

<div class="dd">

<code>disk</code>  <i>string</i>

</div>
<div class="dt">

The disk used for installations.



Examples:


``` yaml
disk: /dev/sda
```

``` yaml
disk: /dev/nvme0
```


</div>

<hr />

<div class="dd">

<code>diskSelector</code>  <i><a href="#installdiskselector">InstallDiskSelector</a></i>

</div>
<div class="dt">

Look up disk using disk characteristics like model, size, serial and others.
Always has priority over `disk`.



Examples:


``` yaml
diskSelector:
    size: 4GB # Disk size.
    model: WDC* # Disk model `/sys/block/<dev>/device/model`.
```


</div>

<hr />

<div class="dd">

<code>extraKernelArgs</code>  <i>[]string</i>

</div>
<div class="dt">

Allows for supplying extra kernel args via the bootloader.



Examples:


``` yaml
extraKernelArgs:
    - talos.platform=metal
    - reboot=k
```


</div>

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

Allows for supplying the image used to perform the installation.
Image reference for each Talos release can be found on
[GitHub releases page](https://github.com/talos-systems/talos/releases).



Examples:


``` yaml
image: ghcr.io/talos-systems/installer:latest
```


</div>

<hr />

<div class="dd">

<code>bootloader</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if a bootloader should be installed.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />

<div class="dd">

<code>wipe</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if the installation disk should be wiped at installation time.
Defaults to `true`.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />

<div class="dd">

<code>legacyBIOSSupport</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if MBR partition should be marked as bootable (active).
Should be enabled only for the systems with legacy BIOS that doesn't support GPT partitioning scheme.

</div>

<hr />





## InstallDiskSizeMatcher
InstallDiskSizeMatcher disk size condition parser.

Appears in:


- <code><a href="#installdiskselector">InstallDiskSelector</a>.size</code>


``` yaml
4GB
```
``` yaml
'> 1TB'
```
``` yaml
<= 2TB
```



## InstallDiskSelector
InstallDiskSelector represents a disk query parameters for the install disk lookup.

Appears in:


- <code><a href="#installconfig">InstallConfig</a>.diskSelector</code>


``` yaml
size: 4GB # Disk size.
model: WDC* # Disk model `/sys/block/<dev>/device/model`.
```

<hr />

<div class="dd">

<code>size</code>  <i><a href="#installdisksizematcher">InstallDiskSizeMatcher</a></i>

</div>
<div class="dt">

Disk size.



Examples:


``` yaml
size: 4GB
```

``` yaml
size: '> 1TB'
```

``` yaml
size: <= 2TB
```


</div>

<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Disk name `/sys/block/<dev>/device/name`.

</div>

<hr />

<div class="dd">

<code>model</code>  <i>string</i>

</div>
<div class="dt">

Disk model `/sys/block/<dev>/device/model`.

</div>

<hr />

<div class="dd">

<code>serial</code>  <i>string</i>

</div>
<div class="dt">

Disk serial number `/sys/block/<dev>/serial`.

</div>

<hr />

<div class="dd">

<code>modalias</code>  <i>string</i>

</div>
<div class="dt">

Disk modalias `/sys/block/<dev>/device/modalias`.

</div>

<hr />

<div class="dd">

<code>uuid</code>  <i>string</i>

</div>
<div class="dt">

Disk UUID `/sys/block/<dev>/uuid`.

</div>

<hr />

<div class="dd">

<code>wwid</code>  <i>string</i>

</div>
<div class="dt">

Disk WWID `/sys/block/<dev>/wwid`.

</div>

<hr />

<div class="dd">

<code>type</code>  <i>InstallDiskType</i>

</div>
<div class="dt">

Disk Type.


Valid values:


  - <code>ssd</code>

  - <code>hdd</code>

  - <code>nvme</code>

  - <code>sd</code>
</div>

<hr />





## TimeConfig
TimeConfig represents the options for configuring time on a machine.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.time</code>


``` yaml
disabled: false # Indicates if the time service is disabled for the machine.
# Specifies time (NTP) servers to use for setting the system time.
servers:
    - time.cloudflare.com
```

<hr />

<div class="dd">

<code>disabled</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if the time service is disabled for the machine.
Defaults to `false`.

</div>

<hr />

<div class="dd">

<code>servers</code>  <i>[]string</i>

</div>
<div class="dt">

Specifies time (NTP) servers to use for setting the system time.
Defaults to `pool.ntp.org`


> This parameter only supports a single time server.

</div>

<hr />





## RegistriesConfig
RegistriesConfig represents the image pull options.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.registries</code>


``` yaml
# Specifies mirror configuration for each registry.
mirrors:
    docker.io:
        # List of endpoints (URLs) for registry mirrors to use.
        endpoints:
            - https://registry.local
# Specifies TLS & auth configuration for HTTPS image registries.
config:
    registry.local:
        # The TLS configuration for the registry.
        tls:
            # Enable mutual TLS authentication with the registry.
            clientIdentity:
                crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
                key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
        # The auth configuration for this registry.
        auth:
            username: username # Optional registry authentication.
            password: password # Optional registry authentication.
```

<hr />

<div class="dd">

<code>mirrors</code>  <i>map[string]<a href="#registrymirrorconfig">RegistryMirrorConfig</a></i>

</div>
<div class="dt">

Specifies mirror configuration for each registry.
This setting allows to use local pull-through caching registires,
air-gapped installations, etc.

Registry name is the first segment of image identifier, with 'docker.io'
being default one.
To catch any registry names not specified explicitly, use '*'.



Examples:


``` yaml
mirrors:
    ghcr.io:
        # List of endpoints (URLs) for registry mirrors to use.
        endpoints:
            - https://registry.insecure
            - https://ghcr.io/v2/
```


</div>

<hr />

<div class="dd">

<code>config</code>  <i>map[string]<a href="#registryconfig">RegistryConfig</a></i>

</div>
<div class="dt">

Specifies TLS & auth configuration for HTTPS image registries.
Mutual TLS can be enabled with 'clientIdentity' option.

TLS configuration can be skipped if registry has trusted
server certificate.



Examples:


``` yaml
config:
    registry.insecure:
        # The TLS configuration for the registry.
        tls:
            insecureSkipVerify: true # Skip TLS server certificate verification (not recommended).

            # # Enable mutual TLS authentication with the registry.
            # clientIdentity:
            #     crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
            #     key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u

        # # The auth configuration for this registry.
        # auth:
        #     username: username # Optional registry authentication.
        #     password: password # Optional registry authentication.
```


</div>

<hr />





## PodCheckpointer
PodCheckpointer represents the pod-checkpointer config values.




<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The `image` field is an override to the default pod-checkpointer image.

</div>

<hr />





## CoreDNS
CoreDNS represents the CoreDNS config values.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.coreDNS</code>


``` yaml
image: docker.io/coredns/coredns:1.8.4 # The `image` field is an override to the default coredns image.
```

<hr />

<div class="dd">

<code>disabled</code>  <i>bool</i>

</div>
<div class="dt">

Disable coredns deployment on cluster bootstrap.

</div>

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The `image` field is an override to the default coredns image.

</div>

<hr />





## Endpoint
Endpoint represents the endpoint URL parsed out of the machine config.

Appears in:


- <code><a href="#controlplaneconfig">ControlPlaneConfig</a>.endpoint</code>


``` yaml
https://1.2.3.4:6443
```
``` yaml
https://cluster1.internal:6443
```



## ControlPlaneConfig
ControlPlaneConfig represents the control plane configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.controlPlane</code>


``` yaml
endpoint: https://1.2.3.4 # Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
localAPIServerPort: 443 # The port that the API server listens on internally.
```

<hr />

<div class="dd">

<code>endpoint</code>  <i><a href="#endpoint">Endpoint</a></i>

</div>
<div class="dt">

Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
It is single-valued, and may optionally include a port number.



Examples:


``` yaml
endpoint: https://1.2.3.4:6443
```

``` yaml
endpoint: https://cluster1.internal:6443
```


</div>

<hr />

<div class="dd">

<code>localAPIServerPort</code>  <i>int</i>

</div>
<div class="dt">

The port that the API server listens on internally.
This may be different than the port portion listed in the endpoint field above.
The default is `6443`.

</div>

<hr />





## APIServerConfig
APIServerConfig represents the kube apiserver configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.apiServer</code>


``` yaml
image: k8s.gcr.io/kube-apiserver:v1.22.0-beta.2 # The container image used in the API server manifest.
# Extra arguments to supply to the API server.
extraArgs:
    feature-gates: ServerSideApply=true
    http2-max-streams-per-connection: "32"
# Extra certificate subject alternative names for the API server's certificate.
certSANs:
    - 1.2.3.4
    - 4.5.6.7
```

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The container image used in the API server manifest.



Examples:


``` yaml
image: k8s.gcr.io/kube-apiserver:v1.22.0-beta.2
```


</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

Extra arguments to supply to the API server.

</div>

<hr />

<div class="dd">

<code>extraVolumes</code>  <i>[]<a href="#volumemountconfig">VolumeMountConfig</a></i>

</div>
<div class="dt">

Extra volumes to mount to the API server static pod.

</div>

<hr />

<div class="dd">

<code>certSANs</code>  <i>[]string</i>

</div>
<div class="dt">

Extra certificate subject alternative names for the API server's certificate.

</div>

<hr />





## ControllerManagerConfig
ControllerManagerConfig represents the kube controller manager configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.controllerManager</code>


``` yaml
image: k8s.gcr.io/kube-controller-manager:v1.22.0-beta.2 # The container image used in the controller manager manifest.
# Extra arguments to supply to the controller manager.
extraArgs:
    feature-gates: ServerSideApply=true
```

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The container image used in the controller manager manifest.



Examples:


``` yaml
image: k8s.gcr.io/kube-controller-manager:v1.22.0-beta.2
```


</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

Extra arguments to supply to the controller manager.

</div>

<hr />

<div class="dd">

<code>extraVolumes</code>  <i>[]<a href="#volumemountconfig">VolumeMountConfig</a></i>

</div>
<div class="dt">

Extra volumes to mount to the controller manager static pod.

</div>

<hr />





## ProxyConfig
ProxyConfig represents the kube proxy configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.proxy</code>


``` yaml
image: k8s.gcr.io/kube-proxy:v1.22.0-beta.2 # The container image used in the kube-proxy manifest.
mode: ipvs # proxy mode of kube-proxy.
# Extra arguments to supply to kube-proxy.
extraArgs:
    proxy-mode: iptables
```

<hr />

<div class="dd">

<code>disabled</code>  <i>bool</i>

</div>
<div class="dt">

Disable kube-proxy deployment on cluster bootstrap.



Examples:


``` yaml
disabled: false
```


</div>

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The container image used in the kube-proxy manifest.



Examples:


``` yaml
image: k8s.gcr.io/kube-proxy:v1.22.0-beta.2
```


</div>

<hr />

<div class="dd">

<code>mode</code>  <i>string</i>

</div>
<div class="dt">

proxy mode of kube-proxy.
The default is 'iptables'.

</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

Extra arguments to supply to kube-proxy.

</div>

<hr />





## SchedulerConfig
SchedulerConfig represents the kube scheduler configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.scheduler</code>


``` yaml
image: k8s.gcr.io/kube-scheduler:v1.22.0-beta.2 # The container image used in the scheduler manifest.
# Extra arguments to supply to the scheduler.
extraArgs:
    feature-gates: AllBeta=true
```

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The container image used in the scheduler manifest.



Examples:


``` yaml
image: k8s.gcr.io/kube-scheduler:v1.22.0-beta.2
```


</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

Extra arguments to supply to the scheduler.

</div>

<hr />

<div class="dd">

<code>extraVolumes</code>  <i>[]<a href="#volumemountconfig">VolumeMountConfig</a></i>

</div>
<div class="dt">

Extra volumes to mount to the scheduler static pod.

</div>

<hr />





## EtcdConfig
EtcdConfig represents the etcd configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.etcd</code>


``` yaml
image: gcr.io/etcd-development/etcd:v3.4.16 # The container image used to create the etcd service.
# The `ca` is the root certificate authority of the PKI.
ca:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
# Extra arguments to supply to etcd.
extraArgs:
    election-timeout: "5000"
```

<hr />

<div class="dd">

<code>image</code>  <i>string</i>

</div>
<div class="dt">

The container image used to create the etcd service.



Examples:


``` yaml
image: gcr.io/etcd-development/etcd:v3.4.16
```


</div>

<hr />

<div class="dd">

<code>ca</code>  <i>PEMEncodedCertificateAndKey</i>

</div>
<div class="dt">

The `ca` is the root certificate authority of the PKI.
It is composed of a base64 encoded `crt` and `key`.



Examples:


``` yaml
ca:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>extraArgs</code>  <i>map[string]string</i>

</div>
<div class="dt">

Extra arguments to supply to etcd.
Note that the following args are not allowed:

- `name`
- `data-dir`
- `initial-cluster-state`
- `listen-peer-urls`
- `listen-client-urls`
- `cert-file`
- `key-file`
- `trusted-ca-file`
- `peer-client-cert-auth`
- `peer-cert-file`
- `peer-trusted-ca-file`
- `peer-key-file`

</div>

<hr />





## ClusterNetworkConfig
ClusterNetworkConfig represents kube networking configuration options.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.network</code>


``` yaml
# The CNI used.
cni:
    name: flannel # Name of CNI to use.
dnsDomain: cluster.local # The domain used by Kubernetes DNS.
# The pod subnet CIDR.
podSubnets:
    - 10.244.0.0/16
# The service subnet CIDR.
serviceSubnets:
    - 10.96.0.0/12
```

<hr />

<div class="dd">

<code>cni</code>  <i><a href="#cniconfig">CNIConfig</a></i>

</div>
<div class="dt">

The CNI used.
Composed of "name" and "urls".
The "name" key supports the following options: "flannel", "custom", and "none".
"flannel" uses Talos-managed Flannel CNI, and that's the default option.
"custom" uses custom manifests that should be provided in "urls".
"none" indicates that Talos will not manage any CNI installation.



Examples:


``` yaml
cni:
    name: custom # Name of CNI to use.
    # URLs containing manifests to apply for the CNI.
    urls:
        - https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml
```


</div>

<hr />

<div class="dd">

<code>dnsDomain</code>  <i>string</i>

</div>
<div class="dt">

The domain used by Kubernetes DNS.
The default is `cluster.local`



Examples:


``` yaml
dnsDomain: cluser.local
```


</div>

<hr />

<div class="dd">

<code>podSubnets</code>  <i>[]string</i>

</div>
<div class="dt">

The pod subnet CIDR.



Examples:


``` yaml
podSubnets:
    - 10.244.0.0/16
```


</div>

<hr />

<div class="dd">

<code>serviceSubnets</code>  <i>[]string</i>

</div>
<div class="dt">

The service subnet CIDR.



Examples:


``` yaml
serviceSubnets:
    - 10.96.0.0/12
```


</div>

<hr />





## CNIConfig
CNIConfig represents the CNI configuration options.

Appears in:


- <code><a href="#clusternetworkconfig">ClusterNetworkConfig</a>.cni</code>


``` yaml
name: custom # Name of CNI to use.
# URLs containing manifests to apply for the CNI.
urls:
    - https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml
```

<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of CNI to use.


Valid values:


  - <code>flannel</code>

  - <code>custom</code>

  - <code>none</code>
</div>

<hr />

<div class="dd">

<code>urls</code>  <i>[]string</i>

</div>
<div class="dt">

URLs containing manifests to apply for the CNI.
Should be present for "custom", must be empty for "flannel" and "none".

</div>

<hr />





## ExternalCloudProviderConfig
ExternalCloudProviderConfig contains external cloud provider configuration.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.externalCloudProvider</code>


``` yaml
enabled: true # Enable external cloud provider.
# A list of urls that point to additional manifests for an external cloud provider.
manifests:
    - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml
    - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml
```

<hr />

<div class="dd">

<code>enabled</code>  <i>bool</i>

</div>
<div class="dt">

Enable external cloud provider.


Valid values:


  - <code>true</code>

  - <code>yes</code>

  - <code>false</code>

  - <code>no</code>
</div>

<hr />

<div class="dd">

<code>manifests</code>  <i>[]string</i>

</div>
<div class="dt">

A list of urls that point to additional manifests for an external cloud provider.
These will get automatically deployed as part of the bootstrap.



Examples:


``` yaml
manifests:
    - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml
    - https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml
```


</div>

<hr />





## AdminKubeconfigConfig
AdminKubeconfigConfig contains admin kubeconfig settings.

Appears in:


- <code><a href="#clusterconfig">ClusterConfig</a>.adminKubeconfig</code>


``` yaml
certLifetime: 1h0m0s # Admin kubeconfig certificate lifetime (default is 1 year).
```

<hr />

<div class="dd">

<code>certLifetime</code>  <i>Duration</i>

</div>
<div class="dt">

Admin kubeconfig certificate lifetime (default is 1 year).
Field format accepts any Go time.Duration format ('1h' for one hour, '10m' for ten minutes).

</div>

<hr />





## MachineDisk
MachineDisk represents the options available for partitioning, formatting, and
mounting extra disks.


Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.disks</code>


``` yaml
- device: /dev/sdb # The name of the disk to use.
  # A list of partitions to create on the disk.
  partitions:
    - mountpoint: /var/mnt/extra # Where to mount the partition.

      # # The size of partition: either bytes or human readable representation. If `size:` is omitted, the partition is sized to occupy the full disk.

      # # Human readable representation.
      # size: 100 MB
      # # Precise value in bytes.
      # size: 1073741824
```

<hr />

<div class="dd">

<code>device</code>  <i>string</i>

</div>
<div class="dt">

The name of the disk to use.

</div>

<hr />

<div class="dd">

<code>partitions</code>  <i>[]<a href="#diskpartition">DiskPartition</a></i>

</div>
<div class="dt">

A list of partitions to create on the disk.

</div>

<hr />





## DiskPartition
DiskPartition represents the options for a disk partition.

Appears in:


- <code><a href="#machinedisk">MachineDisk</a>.partitions</code>



<hr />

<div class="dd">

<code>size</code>  <i>DiskSize</i>

</div>
<div class="dt">

The size of partition: either bytes or human readable representation. If `size:` is omitted, the partition is sized to occupy the full disk.



Examples:


``` yaml
size: 100 MB
```

``` yaml
size: 1073741824
```


</div>

<hr />

<div class="dd">

<code>mountpoint</code>  <i>string</i>

</div>
<div class="dt">

Where to mount the partition.

</div>

<hr />





## EncryptionConfig
EncryptionConfig represents partition encryption settings.

Appears in:


- <code><a href="#systemdiskencryptionconfig">SystemDiskEncryptionConfig</a>.state</code>

- <code><a href="#systemdiskencryptionconfig">SystemDiskEncryptionConfig</a>.ephemeral</code>



<hr />

<div class="dd">

<code>provider</code>  <i>string</i>

</div>
<div class="dt">

Encryption provider to use for the encryption.



Examples:


``` yaml
provider: luks2
```


</div>

<hr />

<div class="dd">

<code>keys</code>  <i>[]<a href="#encryptionkey">EncryptionKey</a></i>

</div>
<div class="dt">

Defines the encryption keys generation and storage method.

</div>

<hr />

<div class="dd">

<code>cipher</code>  <i>string</i>

</div>
<div class="dt">

Cipher kind to use for the encryption. Depends on the encryption provider.

</div>

<hr />





## EncryptionKey
EncryptionKey represents configuration for disk encryption key.

Appears in:


- <code><a href="#encryptionconfig">EncryptionConfig</a>.keys</code>



<hr />

<div class="dd">

<code>static</code>  <i><a href="#encryptionkeystatic">EncryptionKeyStatic</a></i>

</div>
<div class="dt">

Key which value is stored in the configuration file.

</div>

<hr />

<div class="dd">

<code>nodeID</code>  <i><a href="#encryptionkeynodeid">EncryptionKeyNodeID</a></i>

</div>
<div class="dt">

Deterministically generated key from the node UUID and PartitionLabel.

</div>

<hr />

<div class="dd">

<code>slot</code>  <i>int</i>

</div>
<div class="dt">

Key slot number for luks2 encryption.

</div>

<hr />





## EncryptionKeyStatic
EncryptionKeyStatic represents throw away key type.

Appears in:


- <code><a href="#encryptionkey">EncryptionKey</a>.static</code>



<hr />

<div class="dd">

<code>passphrase</code>  <i>string</i>

</div>
<div class="dt">

Defines the static passphrase value.

</div>

<hr />





## EncryptionKeyNodeID
EncryptionKeyNodeID represents deterministically generated key from the node UUID and PartitionLabel.

Appears in:


- <code><a href="#encryptionkey">EncryptionKey</a>.nodeID</code>





## MachineFile
MachineFile represents a file to write to disk.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.files</code>


``` yaml
- content: '...' # The contents of the file.
  permissions: 0o666 # The file's permissions in octal.
  path: /tmp/file.txt # The path of the file.
  op: append # The operation to use
```

<hr />

<div class="dd">

<code>content</code>  <i>string</i>

</div>
<div class="dt">

The contents of the file.

</div>

<hr />

<div class="dd">

<code>permissions</code>  <i>FileMode</i>

</div>
<div class="dt">

The file's permissions in octal.

</div>

<hr />

<div class="dd">

<code>path</code>  <i>string</i>

</div>
<div class="dt">

The path of the file.

</div>

<hr />

<div class="dd">

<code>op</code>  <i>string</i>

</div>
<div class="dt">

The operation to use


Valid values:


  - <code>create</code>

  - <code>append</code>

  - <code>overwrite</code>
</div>

<hr />





## ExtraHost
ExtraHost represents a host entry in /etc/hosts.

Appears in:


- <code><a href="#networkconfig">NetworkConfig</a>.extraHostEntries</code>


``` yaml
- ip: 192.168.1.100 # The IP of the host.
  # The host alias.
  aliases:
    - example
    - example.domain.tld
```

<hr />

<div class="dd">

<code>ip</code>  <i>string</i>

</div>
<div class="dt">

The IP of the host.

</div>

<hr />

<div class="dd">

<code>aliases</code>  <i>[]string</i>

</div>
<div class="dt">

The host alias.

</div>

<hr />





## Device
Device represents a network interface.

Appears in:


- <code><a href="#networkconfig">NetworkConfig</a>.interfaces</code>


``` yaml
- interface: eth0 # The interface name.
  cidr: 192.168.2.0/24 # Assigns a static IP address to the interface.
  # A list of routes associated with the interface.
  routes:
    - network: 0.0.0.0/0 # The route's network.
      gateway: 192.168.2.1 # The route's gateway.
      metric: 1024 # The optional metric for the route.
  mtu: 1500 # The interface's MTU.

  # # Bond specific options.
  # bond:
  #     # The interfaces that make up the bond.
  #     interfaces:
  #         - eth0
  #         - eth1
  #     mode: 802.3ad # A bond option.
  #     lacpRate: fast # A bond option.

  # # Indicates if DHCP should be used to configure the interface.
  # dhcp: true

  # # DHCP specific options.
  # dhcpOptions:
  #     routeMetric: 1024 # The priority of all routes received via DHCP.

  # # Wireguard specific configuration.

  # # wireguard server example
  # wireguard:
  #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
  #     listenPort: 51111 # Specifies a device's listening port.
  #     # Specifies a list of peer configurations to apply to a device.
  #     peers:
  #         - publicKey: ABCDEF... # Specifies the public key of this peer.
  #           endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
  #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
  #           allowedIPs:
  #             - 192.168.1.0/24
  # # wireguard peer example
  # wireguard:
  #     privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
  #     # Specifies a list of peer configurations to apply to a device.
  #     peers:
  #         - publicKey: ABCDEF... # Specifies the public key of this peer.
  #           endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
  #           persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
  #           # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
  #           allowedIPs:
  #             - 192.168.1.0/24

  # # Virtual (shared) IP address configuration.
  # vip:
  #     ip: 172.16.199.55 # Specifies the IP address to be used.
```

<hr />

<div class="dd">

<code>interface</code>  <i>string</i>

</div>
<div class="dt">

The interface name.



Examples:


``` yaml
interface: eth0
```


</div>

<hr />

<div class="dd">

<code>cidr</code>  <i>string</i>

</div>
<div class="dt">

Assigns a static IP address to the interface.
This should be in proper CIDR notation.

> Note: This option is mutually exclusive with DHCP option.



Examples:


``` yaml
cidr: 10.5.0.0/16
```


</div>

<hr />

<div class="dd">

<code>routes</code>  <i>[]<a href="#route">Route</a></i>

</div>
<div class="dt">

A list of routes associated with the interface.
If used in combination with DHCP, these routes will be appended to routes returned by DHCP server.



Examples:


``` yaml
routes:
    - network: 0.0.0.0/0 # The route's network.
      gateway: 10.5.0.1 # The route's gateway.
    - network: 10.2.0.0/16 # The route's network.
      gateway: 10.2.0.1 # The route's gateway.
```


</div>

<hr />

<div class="dd">

<code>bond</code>  <i><a href="#bond">Bond</a></i>

</div>
<div class="dt">

Bond specific options.



Examples:


``` yaml
bond:
    # The interfaces that make up the bond.
    interfaces:
        - eth0
        - eth1
    mode: 802.3ad # A bond option.
    lacpRate: fast # A bond option.
```


</div>

<hr />

<div class="dd">

<code>vlans</code>  <i>[]<a href="#vlan">Vlan</a></i>

</div>
<div class="dt">

VLAN specific options.

</div>

<hr />

<div class="dd">

<code>mtu</code>  <i>int</i>

</div>
<div class="dt">

The interface's MTU.
If used in combination with DHCP, this will override any MTU settings returned from DHCP server.

</div>

<hr />

<div class="dd">

<code>dhcp</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if DHCP should be used to configure the interface.
The following DHCP options are supported:

- `OptionClasslessStaticRoute`
- `OptionDomainNameServer`
- `OptionDNSDomainSearchList`
- `OptionHostName`

> Note: This option is mutually exclusive with CIDR.
>
> Note: To configure an interface with *only* IPv6 SLAAC addressing, CIDR should be set to "" and DHCP to false
> in order for Talos to skip configuration of addresses.
> All other options will still apply.



Examples:


``` yaml
dhcp: true
```


</div>

<hr />

<div class="dd">

<code>ignore</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if the interface should be ignored (skips configuration).

</div>

<hr />

<div class="dd">

<code>dummy</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if the interface is a dummy interface.
`dummy` is used to specify that this interface should be a virtual-only, dummy interface.

</div>

<hr />

<div class="dd">

<code>dhcpOptions</code>  <i><a href="#dhcpoptions">DHCPOptions</a></i>

</div>
<div class="dt">

DHCP specific options.
`dhcp` *must* be set to true for these to take effect.



Examples:


``` yaml
dhcpOptions:
    routeMetric: 1024 # The priority of all routes received via DHCP.
```


</div>

<hr />

<div class="dd">

<code>wireguard</code>  <i><a href="#devicewireguardconfig">DeviceWireguardConfig</a></i>

</div>
<div class="dt">

Wireguard specific configuration.
Includes things like private key, listen port, peers.



Examples:


``` yaml
wireguard:
    privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
    listenPort: 51111 # Specifies a device's listening port.
    # Specifies a list of peer configurations to apply to a device.
    peers:
        - publicKey: ABCDEF... # Specifies the public key of this peer.
          endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
          # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
          allowedIPs:
            - 192.168.1.0/24
```

``` yaml
wireguard:
    privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
    # Specifies a list of peer configurations to apply to a device.
    peers:
        - publicKey: ABCDEF... # Specifies the public key of this peer.
          endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
          persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
          # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
          allowedIPs:
            - 192.168.1.0/24
```


</div>

<hr />

<div class="dd">

<code>vip</code>  <i><a href="#devicevipconfig">DeviceVIPConfig</a></i>

</div>
<div class="dt">

Virtual (shared) IP address configuration.



Examples:


``` yaml
vip:
    ip: 172.16.199.55 # Specifies the IP address to be used.
```


</div>

<hr />





## DHCPOptions
DHCPOptions contains options for configuring the DHCP settings for a given interface.

Appears in:


- <code><a href="#device">Device</a>.dhcpOptions</code>


``` yaml
routeMetric: 1024 # The priority of all routes received via DHCP.
```

<hr />

<div class="dd">

<code>routeMetric</code>  <i>uint32</i>

</div>
<div class="dt">

The priority of all routes received via DHCP.

</div>

<hr />

<div class="dd">

<code>ipv4</code>  <i>bool</i>

</div>
<div class="dt">

Enables DHCPv4 protocol for the interface (default is enabled).

</div>

<hr />

<div class="dd">

<code>ipv6</code>  <i>bool</i>

</div>
<div class="dt">

Enables DHCPv6 protocol for the interface (default is disabled).

</div>

<hr />





## DeviceWireguardConfig
DeviceWireguardConfig contains settings for configuring Wireguard network interface.

Appears in:


- <code><a href="#device">Device</a>.wireguard</code>


``` yaml
privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
listenPort: 51111 # Specifies a device's listening port.
# Specifies a list of peer configurations to apply to a device.
peers:
    - publicKey: ABCDEF... # Specifies the public key of this peer.
      endpoint: 192.168.1.3 # Specifies the endpoint of this peer entry.
      # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      allowedIPs:
        - 192.168.1.0/24
```
``` yaml
privateKey: ABCDEF... # Specifies a private key configuration (base64 encoded).
# Specifies a list of peer configurations to apply to a device.
peers:
    - publicKey: ABCDEF... # Specifies the public key of this peer.
      endpoint: 192.168.1.2 # Specifies the endpoint of this peer entry.
      persistentKeepaliveInterval: 10s # Specifies the persistent keepalive interval for this peer.
      # AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
      allowedIPs:
        - 192.168.1.0/24
```

<hr />

<div class="dd">

<code>privateKey</code>  <i>string</i>

</div>
<div class="dt">

Specifies a private key configuration (base64 encoded).
Can be generated by `wg genkey`.

</div>

<hr />

<div class="dd">

<code>listenPort</code>  <i>int</i>

</div>
<div class="dt">

Specifies a device's listening port.

</div>

<hr />

<div class="dd">

<code>firewallMark</code>  <i>int</i>

</div>
<div class="dt">

Specifies a device's firewall mark.

</div>

<hr />

<div class="dd">

<code>peers</code>  <i>[]<a href="#devicewireguardpeer">DeviceWireguardPeer</a></i>

</div>
<div class="dt">

Specifies a list of peer configurations to apply to a device.

</div>

<hr />





## DeviceWireguardPeer
DeviceWireguardPeer a WireGuard device peer configuration.

Appears in:


- <code><a href="#devicewireguardconfig">DeviceWireguardConfig</a>.peers</code>



<hr />

<div class="dd">

<code>publicKey</code>  <i>string</i>

</div>
<div class="dt">

Specifies the public key of this peer.
Can be extracted from private key by running `wg pubkey < private.key > public.key && cat public.key`.

</div>

<hr />

<div class="dd">

<code>endpoint</code>  <i>string</i>

</div>
<div class="dt">

Specifies the endpoint of this peer entry.

</div>

<hr />

<div class="dd">

<code>persistentKeepaliveInterval</code>  <i>Duration</i>

</div>
<div class="dt">

Specifies the persistent keepalive interval for this peer.
Field format accepts any Go time.Duration format ('1h' for one hour, '10m' for ten minutes).

</div>

<hr />

<div class="dd">

<code>allowedIPs</code>  <i>[]string</i>

</div>
<div class="dt">

AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.

</div>

<hr />





## DeviceVIPConfig
DeviceVIPConfig contains settings for configuring a Virtual Shared IP on an interface.

Appears in:


- <code><a href="#device">Device</a>.vip</code>


``` yaml
ip: 172.16.199.55 # Specifies the IP address to be used.
```

<hr />

<div class="dd">

<code>ip</code>  <i>string</i>

</div>
<div class="dt">

Specifies the IP address to be used.

</div>

<hr />





## Bond
Bond contains the various options for configuring a bonded interface.

Appears in:


- <code><a href="#device">Device</a>.bond</code>


``` yaml
# The interfaces that make up the bond.
interfaces:
    - eth0
    - eth1
mode: 802.3ad # A bond option.
lacpRate: fast # A bond option.
```

<hr />

<div class="dd">

<code>interfaces</code>  <i>[]string</i>

</div>
<div class="dt">

The interfaces that make up the bond.

</div>

<hr />

<div class="dd">

<code>arpIPTarget</code>  <i>[]string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.
Not supported at the moment.

</div>

<hr />

<div class="dd">

<code>mode</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>xmitHashPolicy</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>lacpRate</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>adActorSystem</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.
Not supported at the moment.

</div>

<hr />

<div class="dd">

<code>arpValidate</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>arpAllTargets</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>primary</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>primaryReselect</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>failOverMac</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>adSelect</code>  <i>string</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>miimon</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>updelay</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>downdelay</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>arpInterval</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>resendIgmp</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>minLinks</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>lpInterval</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>packetsPerSlave</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>numPeerNotif</code>  <i>uint8</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>tlbDynamicLb</code>  <i>uint8</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>allSlavesActive</code>  <i>uint8</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>useCarrier</code>  <i>bool</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>adActorSysPrio</code>  <i>uint16</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>adUserPortKey</code>  <i>uint16</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />

<div class="dd">

<code>peerNotifyDelay</code>  <i>uint32</i>

</div>
<div class="dt">

A bond option.
Please see the official kernel documentation.

</div>

<hr />





## Vlan
Vlan represents vlan settings for a device.

Appears in:


- <code><a href="#device">Device</a>.vlans</code>



<hr />

<div class="dd">

<code>cidr</code>  <i>string</i>

</div>
<div class="dt">

The CIDR to use.

</div>

<hr />

<div class="dd">

<code>routes</code>  <i>[]<a href="#route">Route</a></i>

</div>
<div class="dt">

A list of routes associated with the VLAN.

</div>

<hr />

<div class="dd">

<code>dhcp</code>  <i>bool</i>

</div>
<div class="dt">

Indicates if DHCP should be used.

</div>

<hr />

<div class="dd">

<code>vlanId</code>  <i>uint16</i>

</div>
<div class="dt">

The VLAN's ID.

</div>

<hr />





## Route
Route represents a network route.

Appears in:


- <code><a href="#device">Device</a>.routes</code>

- <code><a href="#vlan">Vlan</a>.routes</code>


``` yaml
- network: 0.0.0.0/0 # The route's network.
  gateway: 10.5.0.1 # The route's gateway.
- network: 10.2.0.0/16 # The route's network.
  gateway: 10.2.0.1 # The route's gateway.
```

<hr />

<div class="dd">

<code>network</code>  <i>string</i>

</div>
<div class="dt">

The route's network.

</div>

<hr />

<div class="dd">

<code>gateway</code>  <i>string</i>

</div>
<div class="dt">

The route's gateway.

</div>

<hr />

<div class="dd">

<code>metric</code>  <i>uint32</i>

</div>
<div class="dt">

The optional metric for the route.

</div>

<hr />





## RegistryMirrorConfig
RegistryMirrorConfig represents mirror configuration for a registry.

Appears in:


- <code><a href="#registriesconfig">RegistriesConfig</a>.mirrors</code>


``` yaml
ghcr.io:
    # List of endpoints (URLs) for registry mirrors to use.
    endpoints:
        - https://registry.insecure
        - https://ghcr.io/v2/
```

<hr />

<div class="dd">

<code>endpoints</code>  <i>[]string</i>

</div>
<div class="dt">

List of endpoints (URLs) for registry mirrors to use.
Endpoint configures HTTP/HTTPS access mode, host name,
port and path (if path is not set, it defaults to `/v2`).

</div>

<hr />





## RegistryConfig
RegistryConfig specifies auth & TLS config per registry.

Appears in:


- <code><a href="#registriesconfig">RegistriesConfig</a>.config</code>


``` yaml
registry.insecure:
    # The TLS configuration for the registry.
    tls:
        insecureSkipVerify: true # Skip TLS server certificate verification (not recommended).

        # # Enable mutual TLS authentication with the registry.
        # clientIdentity:
        #     crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
        #     key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u

    # # The auth configuration for this registry.
    # auth:
    #     username: username # Optional registry authentication.
    #     password: password # Optional registry authentication.
```

<hr />

<div class="dd">

<code>tls</code>  <i><a href="#registrytlsconfig">RegistryTLSConfig</a></i>

</div>
<div class="dt">

The TLS configuration for the registry.



Examples:


``` yaml
tls:
    # Enable mutual TLS authentication with the registry.
    clientIdentity:
        crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
        key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```

``` yaml
tls:
    insecureSkipVerify: true # Skip TLS server certificate verification (not recommended).

    # # Enable mutual TLS authentication with the registry.
    # clientIdentity:
    #     crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    #     key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>auth</code>  <i><a href="#registryauthconfig">RegistryAuthConfig</a></i>

</div>
<div class="dt">

The auth configuration for this registry.



Examples:


``` yaml
auth:
    username: username # Optional registry authentication.
    password: password # Optional registry authentication.
```


</div>

<hr />





## RegistryAuthConfig
RegistryAuthConfig specifies authentication configuration for a registry.

Appears in:


- <code><a href="#registryconfig">RegistryConfig</a>.auth</code>


``` yaml
username: username # Optional registry authentication.
password: password # Optional registry authentication.
```

<hr />

<div class="dd">

<code>username</code>  <i>string</i>

</div>
<div class="dt">

Optional registry authentication.
The meaning of each field is the same with the corresponding field in .docker/config.json.

</div>

<hr />

<div class="dd">

<code>password</code>  <i>string</i>

</div>
<div class="dt">

Optional registry authentication.
The meaning of each field is the same with the corresponding field in .docker/config.json.

</div>

<hr />

<div class="dd">

<code>auth</code>  <i>string</i>

</div>
<div class="dt">

Optional registry authentication.
The meaning of each field is the same with the corresponding field in .docker/config.json.

</div>

<hr />

<div class="dd">

<code>identityToken</code>  <i>string</i>

</div>
<div class="dt">

Optional registry authentication.
The meaning of each field is the same with the corresponding field in .docker/config.json.

</div>

<hr />





## RegistryTLSConfig
RegistryTLSConfig specifies TLS config for HTTPS registries.

Appears in:


- <code><a href="#registryconfig">RegistryConfig</a>.tls</code>


``` yaml
# Enable mutual TLS authentication with the registry.
clientIdentity:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```
``` yaml
insecureSkipVerify: true # Skip TLS server certificate verification (not recommended).

# # Enable mutual TLS authentication with the registry.
# clientIdentity:
#     crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
#     key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```

<hr />

<div class="dd">

<code>clientIdentity</code>  <i>PEMEncodedCertificateAndKey</i>

</div>
<div class="dt">

Enable mutual TLS authentication with the registry.
Client certificate and key should be base64-encoded.



Examples:


``` yaml
clientIdentity:
    crt: TFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVSklla05DTUhGLi4u
    key: TFMwdExTMUNSVWRKVGlCRlJESTFOVEU1SUZCU1NWWkJWRVVnUzBWWkxTMHRMUzBLVFVNLi4u
```


</div>

<hr />

<div class="dd">

<code>ca</code>  <i>Base64Bytes</i>

</div>
<div class="dt">

CA registry certificate to add the list of trusted certificates.
Certificate should be base64-encoded.

</div>

<hr />

<div class="dd">

<code>insecureSkipVerify</code>  <i>bool</i>

</div>
<div class="dt">

Skip TLS server certificate verification (not recommended).

</div>

<hr />





## SystemDiskEncryptionConfig
SystemDiskEncryptionConfig specifies system disk partitions encryption settings.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.systemDiskEncryption</code>


``` yaml
# Ephemeral partition encryption.
ephemeral:
    provider: luks2 # Encryption provider to use for the encryption.
    # Defines the encryption keys generation and storage method.
    keys:
        - # Deterministically generated key from the node UUID and PartitionLabel.
          nodeID: {}
          slot: 0 # Key slot number for luks2 encryption.
```

<hr />

<div class="dd">

<code>state</code>  <i><a href="#encryptionconfig">EncryptionConfig</a></i>

</div>
<div class="dt">

State partition encryption.

</div>

<hr />

<div class="dd">

<code>ephemeral</code>  <i><a href="#encryptionconfig">EncryptionConfig</a></i>

</div>
<div class="dt">

Ephemeral partition encryption.

</div>

<hr />





## FeaturesConfig
FeaturesConfig describe individual Talos features that can be switched on or off.

Appears in:


- <code><a href="#machineconfig">MachineConfig</a>.features</code>


``` yaml
rbac: true # Enable role-based access control (RBAC).
```

<hr />

<div class="dd">

<code>rbac</code>  <i>bool</i>

</div>
<div class="dt">

Enable role-based access control (RBAC).

</div>

<hr />





## VolumeMountConfig
VolumeMountConfig struct describes extra volume mount for the static pods.

Appears in:


- <code><a href="#apiserverconfig">APIServerConfig</a>.extraVolumes</code>

- <code><a href="#controllermanagerconfig">ControllerManagerConfig</a>.extraVolumes</code>

- <code><a href="#schedulerconfig">SchedulerConfig</a>.extraVolumes</code>



<hr />

<div class="dd">

<code>hostPath</code>  <i>string</i>

</div>
<div class="dt">

Path on the host.



Examples:


``` yaml
hostPath: /var/lib/auth
```


</div>

<hr />

<div class="dd">

<code>mountPath</code>  <i>string</i>

</div>
<div class="dt">

Path in the container.



Examples:


``` yaml
mountPath: /etc/kubernetes/auth
```


</div>

<hr />

<div class="dd">

<code>readonly</code>  <i>bool</i>

</div>
<div class="dt">

Mount the volume read only.



Examples:


``` yaml
readonly: true
```


</div>

<hr />





## ClusterInlineManifest
ClusterInlineManifest struct describes inline bootstrap manifests for the user.




<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the manifest.
Name should be unique.



Examples:


``` yaml
name: csi
```


</div>

<hr />

<div class="dd">

<code>contents</code>  <i>string</i>

</div>
<div class="dt">

Manifest contents as a string.



Examples:


``` yaml
contents: /etc/kubernetes/auth
```


</div>

<hr />




