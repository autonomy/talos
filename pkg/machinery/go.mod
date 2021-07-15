module github.com/talos-systems/talos/pkg/machinery

go 1.16

// forked go-yaml that introduces RawYAML interface, which can be used to populate YAML fields using bytes
// which are then encoded as a valid YAML blocks with proper indentiation
replace gopkg.in/yaml.v3 => github.com/unix4ever/yaml v0.0.0-20210315173758-8fb30b8e5a5b

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/containerd/go-cni v1.0.2
	github.com/containernetworking/cni v0.8.1 // indirect; security fix in 0.8.1
	github.com/cosi-project/runtime v0.0.0-20210707150857-25f235cd0682
	github.com/dustin/go-humanize v1.0.0
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jsimonetti/rtnetlink v0.0.0-20210614053835-9c52e516c709
	github.com/mdlayher/ethtool v0.0.0-20210210192532-2b88debcdd43
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20200929063507-e6143ca7d51d
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/talos-systems/crypto v0.3.2-0.20210707205149-deec8d47700e
	github.com/talos-systems/go-blockdevice v0.2.1
	github.com/talos-systems/net v0.3.0
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	google.golang.org/genproto v0.0.0-20210701191553-46259e63a0a9
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
