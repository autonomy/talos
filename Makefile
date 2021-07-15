REGISTRY ?= ghcr.io
USERNAME ?= talos-systems
SHA ?= $(shell git describe --match=none --always --abbrev=8 --dirty)
TAG ?= $(shell git describe --tag --always --dirty --match v[0-9]*)
IMAGE_REGISTRY ?= $(REGISTRY)
IMAGE_TAG ?= $(TAG)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REGISTRY_AND_USERNAME := $(IMAGE_REGISTRY)/$(USERNAME)
DOCKER_LOGIN_ENABLED ?= true
NAME = Talos

ARTIFACTS := _out
TOOLS ?= ghcr.io/talos-systems/tools:v0.7.0-alpha.0-2-g7172a5d
PKGS ?= v0.7.0-alpha.0-3-gb6fca88
EXTRAS ?= v0.5.0-alpha.0
GO_VERSION ?= 1.16
GOFUMPT_VERSION ?= v0.1.0
STRINGER_VERSION ?= v0.1.3
DEEPCOPY_GEN_VERSION ?= v0.21.2
VTPROTOBUF_VERSION ?= 81d623a9a700ede8ef765e5ab08b3aa1f5b4d5a8
IMPORTVET ?= autonomy/importvet:f6b07d9
OPERATING_SYSTEM := $(shell uname -s | tr "[:upper:]" "[:lower:]")
TALOSCTL_DEFAULT_TARGET := talosctl-$(OPERATING_SYSTEM)
INTEGRATION_TEST_DEFAULT_TARGET := integration-test-$(OPERATING_SYSTEM)
INTEGRATION_TEST_PROVISION_DEFAULT_TARGET := integration-test-provision-$(OPERATING_SYSTEM)
KUBECTL_URL ?= https://storage.googleapis.com/kubernetes-release/release/v1.21.2/bin/$(OPERATING_SYSTEM)/amd64/kubectl
CLUSTERCTL_VERSION ?= 0.3.19
CLUSTERCTL_URL ?= https://github.com/kubernetes-sigs/cluster-api/releases/download/v$(CLUSTERCTL_VERSION)/clusterctl-$(OPERATING_SYSTEM)-amd64
SONOBUOY_VERSION ?= 0.50.0
SONOBUOY_URL ?= https://github.com/vmware-tanzu/sonobuoy/releases/download/v$(SONOBUOY_VERSION)/sonobuoy_$(SONOBUOY_VERSION)_$(OPERATING_SYSTEM)_amd64.tar.gz
TESTPKGS ?= github.com/talos-systems/talos/...
RELEASES ?= v0.10.4 v0.11.0-beta.0
SHORT_INTEGRATION_TEST ?=
CUSTOM_CNI_URL ?=
INSTALLER_ARCH ?= all

VERSION_PKG = github.com/talos-systems/talos/pkg/version
IMAGES_PKGS = github.com/talos-systems/talos/pkg/images
MGMT_HELPERS_PKG = github.com/talos-systems/talos/cmd/talosctl/pkg/mgmt/helpers

CGO_ENABLED ?= 0
GO_BUILDFLAGS ?=
GO_LDFLAGS ?= \
	-X $(VERSION_PKG).Name=$(NAME) \
	-X $(VERSION_PKG).SHA=$(SHA) \
	-X $(VERSION_PKG).Tag=$(TAG) \
	-X $(VERSION_PKG).PkgsVersion=$(PKGS) \
	-X $(VERSION_PKG).ExtrasVersion=$(EXTRAS) \
	-X $(IMAGES_PKGS).Username=$(USERNAME) \
	-X $(IMAGES_PKGS).Registry=$(REGISTRY) \
	-X $(MGMT_HELPERS_PKG).ArtifactsPath=$(ARTIFACTS)

WITH_RACE ?= false
WITH_DEBUG ?= false

ifneq (, $(filter $(WITH_RACE), t true TRUE y yes 1))
CGO_ENABLED = 1
GO_BUILDFLAGS += -race
GO_LDFLAGS += -linkmode=external -extldflags '-static'
INSTALLER_ARCH = targetarch
endif

ifneq (, $(filter $(WITH_DEBUG), t true TRUE y yes 1))
GO_BUILDFLAGS += -tags sidero.debug
else
GO_LDFLAGS += -s -w
endif

, := ,
space := $(subst ,, )
BUILD := docker buildx build
PLATFORM ?= linux/amd64
PROGRESS ?= auto
PUSH ?= false
COMMON_ARGS := --file=Dockerfile
COMMON_ARGS += --progress=$(PROGRESS)
COMMON_ARGS += --platform=$(PLATFORM)
COMMON_ARGS += --push=$(PUSH)
COMMON_ARGS += --build-arg=TOOLS=$(TOOLS)
COMMON_ARGS += --build-arg=PKGS=$(PKGS)
COMMON_ARGS += --build-arg=EXTRAS=$(EXTRAS)
COMMON_ARGS += --build-arg=GOFUMPT_VERSION=$(GOFUMPT_VERSION)
COMMON_ARGS += --build-arg=STRINGER_VERSION=$(STRINGER_VERSION)
COMMON_ARGS += --build-arg=DEEPCOPY_GEN_VERSION=$(DEEPCOPY_GEN_VERSION)
COMMON_ARGS += --build-arg=VTPROTOBUF_VERSION=$(VTPROTOBUF_VERSION)
COMMON_ARGS += --build-arg=TAG=$(TAG)
COMMON_ARGS += --build-arg=ARTIFACTS=$(ARTIFACTS)
COMMON_ARGS += --build-arg=IMPORTVET=$(IMPORTVET)
COMMON_ARGS += --build-arg=TESTPKGS=$(TESTPKGS)
COMMON_ARGS += --build-arg=INSTALLER_ARCH=$(INSTALLER_ARCH)
COMMON_ARGS += --build-arg=CGO_ENABLED=$(CGO_ENABLED)
COMMON_ARGS += --build-arg=GO_BUILDFLAGS="$(GO_BUILDFLAGS)"
COMMON_ARGS += --build-arg=GO_LDFLAGS="$(GO_LDFLAGS)"
COMMON_ARGS += --build-arg=http_proxy=$(http_proxy)
COMMON_ARGS += --build-arg=https_proxy=$(https_proxy)

CI_ARGS ?=

all: initramfs kernel installer talosctl talosctl-image talos

# Help Menu

define HELP_MENU_HEADER
# Getting Started

To build this project, you must have the following installed:

- git
- make
- docker (19.03 or higher)
- buildx (https://github.com/docker/buildx)

## Creating a Builder Instance

The build process makes use of features not currently supported by the default
builder instance (docker driver). To create a compatible builder instance, run:

```
docker buildx create --driver docker-container --name local --buildkitd-flags '--allow-insecure-entitlement security.insecure' --use
```

If you already have a compatible builder instance, you may use that instead.

> Note: The security.insecure entitlement is only required, and used by the unit-tests target and targets which build container images
for applications using `img` tool.

## Artifacts

All artifacts will be output to ./$(ARTIFACTS). Images will be tagged with the
registry "$(IMAGE_REGISTRY)", username "$(USERNAME)", and a dynamic tag (e.g. $(REGISTRY_AND_USERNAME)/image:$(IMAGE_TAG)).
The registry and username can be overriden by exporting REGISTRY, and USERNAME
respectively.

## Race Detector

Building with `WITH_RACE=1` enables race detector in the Talos executables. Integration tests are always built with the race detector
enabled.

endef

export HELP_MENU_HEADER

help: ## This help menu.
	@echo "$$HELP_MENU_HEADER"
	@grep -E '^[a-zA-Z0-9%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build Abstractions

.PHONY: base
target-%: ## Builds the specified target defined in the Dockerfile. The build result will only remain in the build cache.
	@$(BUILD) \
		--target=$* \
		$(COMMON_ARGS) \
		$(TARGET_ARGS) \
		$(CI_ARGS) .

local-%: ## Builds the specified target defined in the Dockerfile using the local output type. The build result will be output to the specified local destination.
	@$(MAKE) target-$* TARGET_ARGS="--output=type=local,dest=$(DEST) $(TARGET_ARGS)"
	@PLATFORM=$(PLATFORM) \
		ARTIFACTS=$(ARTIFACTS) \
		./hack/fix-artifacts.sh

docker-%: ## Builds the specified target defined in the Dockerfile using the docker output type. The build result will be output to the specified local destination.
	@mkdir -p $(DEST)
	@$(MAKE) target-$* TARGET_ARGS="--output type=docker,dest=$(DEST)/$*.tar,name=$(REGISTRY_AND_USERNAME)/$*:$(IMAGE_TAG) $(TARGET_ARGS)"

registry-%: ## Builds the specified target defined in the Dockerfile using the image/registry output type. The build result will be pushed to the registry if PUSH=true.
	@$(MAKE) target-$* TARGET_ARGS="--output type=image,name=$(REGISTRY_AND_USERNAME)/$*:$(IMAGE_TAG) $(TARGET_ARGS)"

hack-test-%: ## Runs the specied script in ./hack/test with well known environment variables.
	@./hack/test/$*.sh

# Generators

.PHONY: generate
generate: ## Generates code from protobuf service definitions and machinery config.
	@$(MAKE) local-$@ DEST=./ PLATFORM=linux/amd64

.PHONY: docs
docs: ## Generates the documentation for machine config, and talosctl.
	@rm -rf docs/configuration/*
	@rm -rf docs/talosctl/*
	@$(MAKE) local-$@ DEST=./ PLATFORM=linux/amd64

# Local Artifacts

.PHONY: kernel
kernel: ## Outputs the kernel package contents (vmlinuz) to the artifact directory.
	@$(MAKE) local-$@ DEST=$(ARTIFACTS) PUSH=false
	@-rm -rf $(ARTIFACTS)/modules

.PHONY: initramfs
initramfs: ## Builds the compressed initramfs and outputs it to the artifact directory.
	@$(MAKE) local-$@ DEST=$(ARTIFACTS) PUSH=false TARGET_ARGS="--allow security.insecure"

.PHONY: installer
installer: ## Builds the container image for the installer and outputs it to the artifact directory.
	@$(MAKE) registry-$@ TARGET_ARGS="--allow security.insecure"

.PHONY: talos
talos: ## Builds the Talos container image and outputs it to the artifact directory.
	@$(MAKE) registry-$@ TARGET_ARGS="--allow security.insecure"

.PHONY: talosctl-image
talosctl-image: ## Builds the talosctl container image and outputs it to the artifact directory.
	@$(MAKE) registry-talosctl TARGET_ARGS="--allow security.insecure"

talosctl-%:
	@$(MAKE) local-$@ DEST=$(ARTIFACTS) PLATFORM=linux/amd64 PUSH=false NAME=Client

talosctl: $(TALOSCTL_DEFAULT_TARGET) ## Builds the talosctl binary for the local machine.

image-%: ## Builds the specified image. Valid options are aws, azure, digital-ocean, gcp, and vmware (e.g. image-aws)
	@docker pull $(REGISTRY_AND_USERNAME)/installer:$(TAG)
	@for platform in $(subst $(,),$(space),$(PLATFORM)); do \
		arch=`basename "$${platform}"` ; \
		docker run --rm -v /dev:/dev --privileged $(REGISTRY_AND_USERNAME)/installer:$(TAG) image --platform $* --arch $$arch --tar-to-stdout | tar xz -C $(ARTIFACTS) ; \
	done

images: image-aws image-azure image-digital-ocean image-gcp image-metal image-openstack image-vmware ## Builds all known images (AWS, Azure, DigitalOcean, GCP, Metal, Openstack, and VMware).

sbc-%: ## Builds the specified SBC image. Valid options are rpi_4, rock64, bananapi_m64, libretech_all_h3_cc_h5, rockpi_4 and pine64 (e.g. sbc-rpi_4)
	@docker pull $(REGISTRY_AND_USERNAME)/installer:$(TAG)
	@docker run --rm -v /dev:/dev --privileged $(REGISTRY_AND_USERNAME)/installer:$(TAG) image --platform metal --arch arm64 --board $* --tar-to-stdout | tar xz -C $(ARTIFACTS)

sbcs: sbc-rpi_4 sbc-rock64 sbc-bananapi_m64 sbc-libretech_all_h3_cc_h5 sbc-rockpi_4 sbc-pine64 ## Builds all known SBC images (Raspberry Pi 4 Model B, Rock64, Banana Pi M64, Radxa ROCK Pi 4, pine64, and Libre Computer Board ALL-H3-CC).

.PHONY: iso
iso: ## Builds the ISO and outputs it to the artifact directory.
	@docker pull $(REGISTRY_AND_USERNAME)/installer:$(TAG)
	@for platform in $(subst $(,),$(space),$(PLATFORM)); do \
		arch=`basename "$${platform}"` ; \
		docker run --rm -i $(REGISTRY_AND_USERNAME)/installer:$(TAG) iso --arch $$arch --tar-to-stdout | tar xz -C $(ARTIFACTS)  ; \
	done

.PHONY: boot
boot: ## Creates a compressed tarball that includes vmlinuz-{amd64,arm64} and initramfs-{amd64,arm64}.xz. Note that these files must already be present in the artifacts directory.
	@for platform in $(subst $(,),$(space),$(PLATFORM)); do \
		arch=`basename "$${platform}"` ; \
		tar  -C $(ARTIFACTS) --transform=s/-$${arch}// -czf $(ARTIFACTS)/boot-$${arch}.tar.gz vmlinuz-$${arch} initramfs-$${arch}.xz ; \
	done

.PHONY: talosctl-cni-bundle
talosctl-cni-bundle: ## Creates a compressed tarball that includes CNI bundle for talosctl.
	@$(MAKE) local-$@ DEST=$(ARTIFACTS)
	@for platform in $(subst $(,),$(space),$(PLATFORM)); do \
		arch=`basename "$${platform}"` ; \
		tar  -C $(ARTIFACTS)/talosctl-cni-bundle-$${arch} -czf $(ARTIFACTS)/talosctl-cni-bundle-$${arch}.tar.gz . ; \
	done
	@rm -rf $(ARTIFACTS)/talosctl-cni-bundle-*/

.PHONY: cloud-images
cloud-images: ## Uploads cloud images (AMIs, etc.) to the cloud registry.
	@docker run --rm -v $(PWD):/src -w /src \
		-e TAG=$(TAG) -e ARTIFACTS=$(ARTIFACTS) \
		-e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SVC_ACCT \
		-e AZURE_SVC_ACCT -e GCE_SVC_ACCT -e PACKET_AUTH_TOKEN \
		golang:$(GO_VERSION) \
		./hack/cloud-image-uploader.sh

# Code Quality

.PHONY: fmt
fmt: ## Formats the source code.
	@docker run --rm -it -v $(PWD):/src -w /src golang:$(GO_VERSION) bash -c "go install mvdan.cc/gofumpt/gofumports@$(GOFUMPT_VERSION) && gofumports -w -local github.com/talos-systems/talos ."

lint-%: ## Runs the specified linter. Valid options are go, protobuf, and markdown (e.g. lint-go).
	@$(MAKE) target-lint-$* PLATFORM=linux/amd64

lint: ## Runs linters on go, protobuf, and markdown file types.
	@$(MAKE) lint-go lint-protobuf lint-markdown

check-dirty: ## Verifies that source tree is not dirty
	@if test -n "`git status --porcelain`"; then echo "Source tree is dirty"; git status; exit 1 ; fi

# Tests

.PHONY: unit-tests
unit-tests: ## Performs unit tests.
	@$(MAKE) local-$@ DEST=$(ARTIFACTS) TARGET_ARGS="--allow security.insecure" PLATFORM=linux/amd64

.PHONY: unit-tests-race
unit-tests-race: ## Performs unit tests with race detection enabled.
	@$(MAKE) target-$@ TARGET_ARGS="--allow security.insecure" PLATFORM=linux/amd64

$(ARTIFACTS)/$(INTEGRATION_TEST_DEFAULT_TARGET)-amd64:
	@$(MAKE) local-$(INTEGRATION_TEST_DEFAULT_TARGET) DEST=$(ARTIFACTS) PLATFORM=linux/amd64 WITH_RACE=true NAME=Client

$(ARTIFACTS)/$(INTEGRATION_TEST_PROVISION_DEFAULT_TARGET)-amd64:
	@$(MAKE) local-$(INTEGRATION_TEST_PROVISION_DEFAULT_TARGET) DEST=$(ARTIFACTS) PLATFORM=linux/amd64 WITH_RACE=true NAME=Client

$(ARTIFACTS)/sonobuoy:
	@mkdir -p $(ARTIFACTS)
	@curl -L -o /tmp/sonobuoy.tar.gz ${SONOBUOY_URL}
	@tar -xf /tmp/sonobuoy.tar.gz -C $(ARTIFACTS)

$(ARTIFACTS)/kubectl:
	@mkdir -p $(ARTIFACTS)
	@curl -L -o $(ARTIFACTS)/kubectl "$(KUBECTL_URL)"
	@chmod +x $(ARTIFACTS)/kubectl

$(ARTIFACTS)/clusterctl:
	@mkdir -p $(ARTIFACTS)
	@curl -L -o $(ARTIFACTS)/clusterctl "$(CLUSTERCTL_URL)"
	@chmod +x $(ARTIFACTS)/clusterctl

e2e-%: $(ARTIFACTS)/$(INTEGRATION_TEST_DEFAULT_TARGET)-amd64 $(ARTIFACTS)/sonobuoy $(ARTIFACTS)/kubectl $(ARTIFACTS)/clusterctl ## Runs the E2E test for the specified platform (e.g. e2e-docker).
	@$(MAKE) hack-test-$@ \
		PLATFORM=$* \
		TAG=$(TAG) \
		SHA=$(SHA) \
		REGISTRY=$(IMAGE_REGISTRY) \
		IMAGE=$(REGISTRY_AND_USERNAME)/talos:$(TAG) \
		INSTALLER_IMAGE=$(REGISTRY_AND_USERNAME)/installer:$(TAG) \
		ARTIFACTS=$(ARTIFACTS) \
		TALOSCTL=$(PWD)/$(ARTIFACTS)/$(TALOSCTL_DEFAULT_TARGET)-amd64 \
		INTEGRATION_TEST=$(PWD)/$(ARTIFACTS)/$(INTEGRATION_TEST_DEFAULT_TARGET)-amd64 \
		SHORT_INTEGRATION_TEST=$(SHORT_INTEGRATION_TEST) \
		CUSTOM_CNI_URL=$(CUSTOM_CNI_URL) \
		KUBECTL=$(PWD)/$(ARTIFACTS)/kubectl \
		SONOBUOY=$(PWD)/$(ARTIFACTS)/sonobuoy \
		CLUSTERCTL=$(PWD)/$(ARTIFACTS)/clusterctl

provision-tests-prepare: release-artifacts $(ARTIFACTS)/$(INTEGRATION_TEST_PROVISION_DEFAULT_TARGET)-amd64

provision-tests: provision-tests-prepare
	@$(MAKE) hack-test-$@ \
		TAG=$(TAG) \
		TALOSCTL=$(PWD)/$(ARTIFACTS)/$(TALOSCTL_DEFAULT_TARGET)-amd64 \
		INTEGRATION_TEST=$(PWD)/$(ARTIFACTS)/$(INTEGRATION_TEST_PROVISION_DEFAULT_TARGET)-amd64

provision-tests-track-%:
	@$(MAKE) hack-test-provision-tests \
		TAG=$(TAG) \
		TALOSCTL=$(PWD)/$(ARTIFACTS)/$(TALOSCTL_DEFAULT_TARGET)-amd64 \
		INTEGRATION_TEST=$(PWD)/$(ARTIFACTS)/$(INTEGRATION_TEST_PROVISION_DEFAULT_TARGET)-amd64 \
		INTEGRATION_TEST_RUN="TestIntegration/.+-TR$*" \
		INTEGRATION_TEST_TRACK="$*" \
		CUSTOM_CNI_URL=$(CUSTOM_CNI_URL) \
		REGISTRY=$(IMAGE_REGISTRY) \
		ARTIFACTS=$(ARTIFACTS)

# Assets for releases

.PHONY: $(ARTIFACTS)/$(TALOS_RELEASE)
$(ARTIFACTS)/$(TALOS_RELEASE): $(ARTIFACTS)/$(TALOS_RELEASE)/vmlinuz $(ARTIFACTS)/$(TALOS_RELEASE)/initramfs.xz

# download release artifacts for specific version
$(ARTIFACTS)/$(TALOS_RELEASE)/%:
	@mkdir -p $(ARTIFACTS)/$(TALOS_RELEASE)/
	@case "$(TALOS_RELEASE)" in \
		v0.6*) \
			curl -L -o "$(ARTIFACTS)/$(TALOS_RELEASE)/$*" "https://github.com/talos-systems/talos/releases/download/$(TALOS_RELEASE)/$*" \
			;; \
		v0.7.0-alpha.[0-3]) \
			curl -L -o "$(ARTIFACTS)/$(TALOS_RELEASE)/$*" "https://github.com/talos-systems/talos/releases/download/$(TALOS_RELEASE)/$*" \
			;; \
		*) \
			case "$*" in \
				vmlinuz) \
					curl -L -o "$(ARTIFACTS)/$(TALOS_RELEASE)/$*" "https://github.com/talos-systems/talos/releases/download/$(TALOS_RELEASE)/vmlinuz-amd64" \
					;; \
				initramfs.xz) \
					curl -L -o "$(ARTIFACTS)/$(TALOS_RELEASE)/$*" "https://github.com/talos-systems/talos/releases/download/$(TALOS_RELEASE)/initramfs-amd64.xz" \
					;; \
			esac \
			;; \
	esac


.PHONY: release-artifacts
release-artifacts:
	@for release in $(RELEASES); do \
		$(MAKE) $(ARTIFACTS)/$$release TALOS_RELEASE=$$release; \
	done

# Utilities

.PHONY: conformance
conformance: ## Performs policy checks against the commit and source code.
	docker run --rm -it -v $(PWD):/src -w /src docker.io/autonomy/conform:v0.1.0-alpha.20

.PHONY: release-notes
release-notes:
	ARTIFACTS=$(ARTIFACTS) ./hack/release.sh $@ $(ARTIFACTS)/RELEASE_NOTES.md $(TAG)

.PHONY: login
login: ## Logs in to the configured container registry.
ifeq ($(DOCKER_LOGIN_ENABLED), true)
	@docker login --username "$(GHCR_USERNAME)" --password "$(GHCR_PASSWORD)" $(IMAGE_REGISTRY)
endif

push: login ## Pushes the installer, and talos images to the configured container registry with the generated tag.
	@$(MAKE) installer PUSH=true
	@$(MAKE) talos PUSH=true
	@$(MAKE) talosctl-image PUSH=true

push-%: login ## Pushes the installer, and talos images to the configured container registry with the specified tag (e.g. push-latest).
	@$(MAKE) push IMAGE_TAG=$*

.PHONY: clean
clean: ## Cleans up all artifacts.
	@-rm -rf $(ARTIFACTS)
