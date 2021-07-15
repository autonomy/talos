# syntax = docker/dockerfile-upstream:1.2.0-labs

# Meta args applied to stage base names.

ARG TOOLS
ARG IMPORTVET
ARG PKGS
ARG EXTRAS
ARG INSTALLER_ARCH

# Resolve package images using ${PKGS} to be used later in COPY --from=.

FROM ghcr.io/talos-systems/fhs:${PKGS} AS pkg-fhs
FROM ghcr.io/talos-systems/ca-certificates:${PKGS} AS pkg-ca-certificates

FROM --platform=amd64 ghcr.io/talos-systems/cryptsetup:${PKGS} AS pkg-cryptsetup-amd64
FROM --platform=arm64 ghcr.io/talos-systems/cryptsetup:${PKGS} AS pkg-cryptsetup-arm64

FROM --platform=amd64 ghcr.io/talos-systems/containerd:${PKGS} AS pkg-containerd-amd64
FROM --platform=arm64 ghcr.io/talos-systems/containerd:${PKGS} AS pkg-containerd-arm64

FROM --platform=amd64 ghcr.io/talos-systems/dosfstools:${PKGS} AS pkg-dosfstools-amd64
FROM --platform=arm64 ghcr.io/talos-systems/dosfstools:${PKGS} AS pkg-dosfstools-arm64

FROM --platform=amd64 ghcr.io/talos-systems/eudev:${PKGS} AS pkg-eudev-amd64
FROM --platform=arm64 ghcr.io/talos-systems/eudev:${PKGS} AS pkg-eudev-arm64

FROM ghcr.io/talos-systems/grub:${PKGS} AS pkg-grub
FROM --platform=amd64 ghcr.io/talos-systems/grub:${PKGS} AS pkg-grub-amd64
FROM --platform=arm64 ghcr.io/talos-systems/grub:${PKGS} AS pkg-grub-arm64

FROM --platform=amd64 ghcr.io/talos-systems/iptables:${PKGS} AS pkg-iptables-amd64
FROM --platform=arm64 ghcr.io/talos-systems/iptables:${PKGS} AS pkg-iptables-arm64

FROM --platform=amd64 ghcr.io/talos-systems/libjson-c:${PKGS} AS pkg-libjson-c-amd64
FROM --platform=arm64 ghcr.io/talos-systems/libjson-c:${PKGS} AS pkg-libjson-c-arm64

FROM --platform=amd64 ghcr.io/talos-systems/libpopt:${PKGS} AS pkg-libpopt-amd64
FROM --platform=arm64 ghcr.io/talos-systems/libpopt:${PKGS} AS pkg-libpopt-arm64

FROM --platform=amd64 ghcr.io/talos-systems/libressl:${PKGS} AS pkg-libressl-amd64
FROM --platform=arm64 ghcr.io/talos-systems/libressl:${PKGS} AS pkg-libressl-arm64

FROM --platform=amd64 ghcr.io/talos-systems/libseccomp:${PKGS} AS pkg-libseccomp-amd64
FROM --platform=arm64 ghcr.io/talos-systems/libseccomp:${PKGS} AS pkg-libseccomp-arm64

FROM --platform=amd64 ghcr.io/talos-systems/linux-firmware:${PKGS} AS pkg-linux-firmware-amd64
FROM --platform=arm64 ghcr.io/talos-systems/linux-firmware:${PKGS} AS pkg-linux-firmware-arm64

FROM --platform=amd64 ghcr.io/talos-systems/lvm2:${PKGS} AS pkg-lvm2-amd64
FROM --platform=arm64 ghcr.io/talos-systems/lvm2:${PKGS} AS pkg-lvm2-arm64

FROM --platform=amd64 ghcr.io/talos-systems/libaio:${PKGS} AS pkg-libaio-amd64
FROM --platform=arm64 ghcr.io/talos-systems/libaio:${PKGS} AS pkg-libaio-arm64

FROM --platform=amd64 ghcr.io/talos-systems/musl:${PKGS} AS pkg-musl-amd64
FROM --platform=arm64 ghcr.io/talos-systems/musl:${PKGS} AS pkg-musl-arm64

FROM --platform=amd64 ghcr.io/talos-systems/open-iscsi:${PKGS} AS pkg-open-iscsi-amd64
FROM --platform=arm64 ghcr.io/talos-systems/open-iscsi:${PKGS} AS pkg-open-iscsi-arm64

FROM --platform=amd64 ghcr.io/talos-systems/open-isns:${PKGS} AS pkg-open-isns-amd64
FROM --platform=arm64 ghcr.io/talos-systems/open-isns:${PKGS} AS pkg-open-isns-arm64

FROM --platform=amd64 ghcr.io/talos-systems/runc:${PKGS} AS pkg-runc-amd64
FROM --platform=arm64 ghcr.io/talos-systems/runc:${PKGS} AS pkg-runc-arm64

FROM --platform=amd64 ghcr.io/talos-systems/xfsprogs:${PKGS} AS pkg-xfsprogs-amd64
FROM --platform=arm64 ghcr.io/talos-systems/xfsprogs:${PKGS} AS pkg-xfsprogs-arm64

FROM --platform=amd64 ghcr.io/talos-systems/util-linux:${PKGS} AS pkg-util-linux-amd64
FROM --platform=arm64 ghcr.io/talos-systems/util-linux:${PKGS} AS pkg-util-linux-arm64

FROM --platform=amd64 ghcr.io/talos-systems/kmod:${PKGS} AS pkg-kmod-amd64
FROM --platform=arm64 ghcr.io/talos-systems/kmod:${PKGS} AS pkg-kmod-arm64

FROM ghcr.io/talos-systems/kernel:${PKGS} AS pkg-kernel
FROM --platform=amd64 ghcr.io/talos-systems/kernel:${PKGS} AS pkg-kernel-amd64
FROM --platform=arm64 ghcr.io/talos-systems/kernel:${PKGS} AS pkg-kernel-arm64

FROM --platform=arm64 ghcr.io/talos-systems/u-boot:${PKGS} AS pkg-u-boot-arm64
FROM --platform=arm64 ghcr.io/talos-systems/raspberrypi-firmware:${PKGS} AS pkg-raspberrypi-firmware-arm64

# Resolve package images using ${EXTRAS} to be used later in COPY --from=.

FROM ghcr.io/talos-systems/talosctl-cni-bundle-install:${EXTRAS} AS extras-talosctl-cni-bundle-install

# The tools target provides base toolchain for the build.

FROM $IMPORTVET as importvet

FROM --platform=${BUILDPLATFORM} $TOOLS AS tools
ENV PATH /toolchain/bin:/toolchain/go/bin
RUN ["/toolchain/bin/mkdir", "/bin", "/tmp"]
RUN ["/toolchain/bin/ln", "-svf", "/toolchain/bin/bash", "/bin/sh"]
RUN ["/toolchain/bin/ln", "-svf", "/toolchain/etc/ssl", "/etc/ssl"]
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b /toolchain/bin v1.38.0
ARG GOFUMPT_VERSION
RUN go install mvdan.cc/gofumpt/gofumports@${GOFUMPT_VERSION} \
    && mv /go/bin/gofumports /toolchain/go/bin/gofumports
ARG STRINGER_VERSION
RUN go install golang.org/x/tools/cmd/stringer@${STRINGER_VERSION} \
    && mv /go/bin/stringer /toolchain/go/bin/stringer
ARG DEEPCOPY_GEN_VERSION
RUN go install k8s.io/code-generator/cmd/deepcopy-gen@${DEEPCOPY_GEN_VERSION} \
    && mv /go/bin/deepcopy-gen /toolchain/go/bin/deepcopy-gen
ARG VTPROTOBUF_VERSION
RUN go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@${VTPROTOBUF_VERSION} \
    && mv /go/bin/protoc-gen-go-vtproto /toolchain/go/bin/protoc-gen-go-vtproto
RUN curl -sfL https://github.com/uber/prototool/releases/download/v1.10.0/prototool-Linux-x86_64.tar.gz | tar -xz --strip-components=2 -C /toolchain/bin prototool/bin/prototool
COPY ./hack/docgen /go/src/github.com/talos-systems/talos-hack-docgen
RUN cd /go/src/github.com/talos-systems/talos-hack-docgen \
    && go build -o docgen . \
    && mv docgen /toolchain/go/bin/
COPY --from=importvet /importvet /toolchain/go/bin/importvet

# The build target creates a container that will be used to build Talos source
# code.

FROM --platform=${BUILDPLATFORM} tools AS build
SHELL ["/toolchain/bin/bash", "-c"]
ENV PATH /toolchain/bin:/toolchain/go/bin
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org
ARG CGO_ENABLED
ENV CGO_ENABLED ${CGO_ENABLED}
ENV GOCACHE /.cache/go-build
ENV GOMODCACHE /.cache/mod
ENV SOURCE_DATE_EPOCH=0
WORKDIR /src

# The build-go target creates a container to build Go code with Go modules downloaded and verified.

FROM build AS build-go
COPY ./go.mod ./go.sum ./
COPY ./pkg/machinery/go.mod ./pkg/machinery/go.sum ./pkg/machinery/
WORKDIR /src/pkg/machinery
RUN --mount=type=cache,target=/.cache go mod download
WORKDIR /src
RUN --mount=type=cache,target=/.cache go mod download
RUN --mount=type=cache,target=/.cache go mod verify

# The generate target generates code from protobuf service definitions and machinery config.

FROM build AS generate-build
# Common needs to be at or near the top to satisfy the subsequent imports
COPY ./api/vendor/ /api/vendor/
COPY ./api/common/common.proto /api/common/common.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size common/common.proto
COPY ./api/security/security.proto /api/security/security.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size security/security.proto
COPY ./api/storage/storage.proto /api/storage/storage.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size storage/storage.proto
COPY ./api/machine/machine.proto /api/machine/machine.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size machine/machine.proto
COPY ./api/time/time.proto /api/time/time.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size time/time.proto
COPY ./api/network/network.proto /api/network/network.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size network/network.proto
COPY ./api/cluster/cluster.proto /api/cluster/cluster.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size cluster/cluster.proto
COPY ./api/resource/resource.proto /api/resource/resource.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size resource/resource.proto
COPY ./api/resource/secrets/secrets.proto /api/resource/secrets/secrets.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size resource/secrets/secrets.proto
COPY ./api/inspect/inspect.proto /api/inspect/inspect.proto
RUN protoc -I/api -I/api/vendor/ --go_out=paths=source_relative:/api --go-grpc_out=paths=source_relative:/api --go-vtproto_out=paths=source_relative:/api --go-vtproto_opt=features=marshal+unmarshal+size inspect/inspect.proto
# Gofumports generated files to adjust import order
RUN gofumports -w -local github.com/talos-systems/talos /api/

# run docgen for machinery config
FROM build-go AS go-generate
COPY ./pkg ./pkg
COPY ./hack/boilerplate.txt ./hack/boilerplate.txt
RUN --mount=type=cache,target=/.cache go generate ./pkg/...
WORKDIR /src/pkg/machinery
RUN --mount=type=cache,target=/.cache go generate ./...

FROM --platform=${BUILDPLATFORM} scratch AS generate
COPY --from=generate-build /api/common/*.pb.go /pkg/machinery/api/common/
COPY --from=generate-build /api/security/*.pb.go /pkg/machinery/api/security/
COPY --from=generate-build /api/machine/*.pb.go /pkg/machinery/api/machine/
COPY --from=generate-build /api/time/*.pb.go /pkg/machinery/api/time/
COPY --from=generate-build /api/network/*.pb.go /pkg/machinery/api/network/
COPY --from=generate-build /api/cluster/*.pb.go /pkg/machinery/api/cluster/
COPY --from=generate-build /api/storage/*.pb.go /pkg/machinery/api/storage/
COPY --from=generate-build /api/resource/*.pb.go /pkg/machinery/api/resource/
COPY --from=generate-build /api/resource/secrets/*.pb.go /pkg/machinery/api/resource/secrets/
COPY --from=generate-build /api/inspect/*.pb.go /pkg/machinery/api/inspect/
COPY --from=go-generate /src/pkg/resources/network/ /pkg/resources/network/
COPY --from=go-generate /src/pkg/machinery/config/types/v1alpha1/ /pkg/machinery/config/types/v1alpha1/
COPY --from=go-generate /src/pkg/machinery/nethelpers/ /pkg/machinery/nethelpers/

# The base target provides a container that can be used to build all Talos
# assets.

FROM build-go AS base
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./internal ./internal
COPY --from=generate /pkg/machinery/ ./pkg/machinery/
RUN --mount=type=cache,target=/.cache go list -mod=readonly all >/dev/null
RUN --mount=type=cache,target=/.cache ! go mod tidy -v 2>&1 | grep .
WORKDIR /src/pkg/machinery
RUN --mount=type=cache,target=/.cache go mod download
RUN --mount=type=cache,target=/.cache go list -mod=readonly all >/dev/null
RUN --mount=type=cache,target=/.cache ! go mod tidy -v 2>&1 | grep .
WORKDIR /src

# The init target builds the init binary.

FROM base AS init-build-amd64
WORKDIR /src/internal/app/init
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=amd64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /init
RUN chmod +x /init

FROM base AS init-build-arm64
WORKDIR /src/internal/app/init
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=arm64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /init
RUN chmod +x /init

FROM init-build-${TARGETARCH} AS init-build

FROM scratch AS init
COPY --from=init-build /init /init

# The machined target builds the machined binary.

FROM base AS machined-build-amd64
WORKDIR /src/internal/app/machined
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=amd64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /machined
RUN chmod +x /machined

FROM base AS machined-build-arm64
WORKDIR /src/internal/app/machined
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=arm64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /machined
RUN chmod +x /machined

FROM machined-build-${TARGETARCH} AS machined-build

FROM scratch AS machined
COPY --from=machined-build /machined /machined

# The talosctl targets build the talosctl binaries.

FROM base AS talosctl-linux-amd64-build
WORKDIR /src/cmd/talosctl
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=amd64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /talosctl-linux-amd64
RUN chmod +x /talosctl-linux-amd64

FROM base AS talosctl-linux-arm64-build
WORKDIR /src/cmd/talosctl
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=arm64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /talosctl-linux-arm64
RUN chmod +x /talosctl-linux-arm64

FROM base AS talosctl-linux-armv7-build
WORKDIR /src/cmd/talosctl
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=arm GOARM=7 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /talosctl-linux-armv7
RUN chmod +x /talosctl-linux-armv7

FROM scratch AS talosctl-linux
COPY --from=talosctl-linux-amd64-build /talosctl-linux-amd64 /talosctl-linux-amd64
COPY --from=talosctl-linux-arm64-build /talosctl-linux-arm64 /talosctl-linux-arm64
COPY --from=talosctl-linux-armv7-build /talosctl-linux-armv7 /talosctl-linux-armv7

FROM scratch as talosctl
ARG TARGETARCH
COPY --from=talosctl-linux /talosctl-linux-${TARGETARCH} /talosctl
ARG TAG
ENV VERSION ${TAG}
LABEL "alpha.talos.dev/version"="${VERSION}"
LABEL org.opencontainers.image.source https://github.com/talos-systems/talos
ENTRYPOINT ["/talosctl"]

FROM base AS talosctl-darwin-amd64-build
WORKDIR /src/cmd/talosctl
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=darwin GOARCH=amd64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /talosctl-darwin-amd64
RUN chmod +x /talosctl-darwin-amd64

FROM base AS talosctl-darwin-arm64-build
WORKDIR /src/cmd/talosctl
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=darwin GOARCH=arm64 go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /talosctl-darwin-arm64
RUN chmod +x /talosctl-darwin-arm64

FROM scratch AS talosctl-darwin
COPY --from=talosctl-darwin-amd64-build /talosctl-darwin-amd64 /talosctl-darwin-amd64
COPY --from=talosctl-darwin-arm64-build /talosctl-darwin-arm64 /talosctl-darwin-arm64

# The kernel target is the linux kernel.

FROM scratch AS kernel
ARG TARGETARCH
COPY --from=pkg-kernel /boot/vmlinuz /vmlinuz-${TARGETARCH}

# The rootfs target provides the Talos rootfs.

FROM build AS rootfs-base-amd64
COPY --from=pkg-fhs / /rootfs
COPY --from=pkg-ca-certificates / /rootfs
COPY --from=pkg-cryptsetup-amd64 / /rootfs
COPY --from=pkg-containerd-amd64 / /rootfs
COPY --from=pkg-dosfstools-amd64 / /rootfs
COPY --from=pkg-eudev-amd64 / /rootfs
COPY --from=pkg-iptables-amd64 / /rootfs
COPY --from=pkg-libjson-c-amd64 / /rootfs
COPY --from=pkg-libpopt-amd64 / /rootfs
COPY --from=pkg-libressl-amd64 / /rootfs
COPY --from=pkg-libseccomp-amd64 / /rootfs
COPY --from=pkg-linux-firmware-amd64 /lib/firmware/bnx2 /rootfs/lib/firmware/bnx2
COPY --from=pkg-linux-firmware-amd64 /lib/firmware/bnx2x /rootfs/lib/firmware/bnx2x
COPY --from=pkg-lvm2-amd64 / /rootfs
COPY --from=pkg-libaio-amd64 / /rootfs
COPY --from=pkg-musl-amd64 / /rootfs
COPY --from=pkg-open-iscsi-amd64 / /rootfs
COPY --from=pkg-open-isns-amd64 / /rootfs
COPY --from=pkg-runc-amd64 / /rootfs
COPY --from=pkg-xfsprogs-amd64 / /rootfs
COPY --from=pkg-util-linux-amd64 /lib/libblkid.* /rootfs/lib/
COPY --from=pkg-util-linux-amd64 /lib/libuuid.* /rootfs/lib/
COPY --from=pkg-util-linux-amd64 /lib/libmount.* /rootfs/lib/
COPY --from=pkg-kmod-amd64 /usr/lib/libkmod.* /rootfs/lib/
COPY --from=pkg-kernel-amd64 /lib/modules /rootfs/lib/modules
COPY --from=machined-build-amd64 /machined /rootfs/sbin/init
# NB: We run the cleanup step before creating extra directories, files, and
# symlinks to avoid accidentally cleaning them up.
COPY ./hack/cleanup.sh /toolchain/bin/cleanup.sh
RUN cleanup.sh /rootfs
COPY hack/containerd.toml /rootfs/etc/cri/containerd.toml
RUN touch /rootfs/etc/resolv.conf
RUN touch /rootfs/etc/hosts
RUN touch /rootfs/etc/os-release
RUN mkdir -pv /rootfs/{boot,usr/local/share,mnt,system}
RUN mkdir -pv /rootfs/{etc/kubernetes/manifests,etc/cni,usr/libexec/kubernetes}
RUN ln -s /etc/ssl /rootfs/etc/pki
RUN ln -s /etc/ssl /rootfs/usr/share/ca-certificates
RUN ln -s /etc/ssl /rootfs/usr/local/share/ca-certificates
RUN ln -s /etc/ssl /rootfs/etc/ca-certificates

FROM build AS rootfs-base-arm64
COPY --from=pkg-fhs / /rootfs
COPY --from=pkg-ca-certificates / /rootfs
COPY --from=pkg-cryptsetup-arm64 / /rootfs
COPY --from=pkg-containerd-arm64 / /rootfs
COPY --from=pkg-dosfstools-arm64 / /rootfs
COPY --from=pkg-eudev-arm64 / /rootfs
COPY --from=pkg-iptables-arm64 / /rootfs
COPY --from=pkg-libjson-c-arm64 / /rootfs
COPY --from=pkg-libpopt-arm64 / /rootfs
COPY --from=pkg-libressl-arm64 / /rootfs
COPY --from=pkg-libseccomp-arm64 / /rootfs
COPY --from=pkg-linux-firmware-arm64 /lib/firmware/bnx2 /rootfs/lib/firmware/bnx2
COPY --from=pkg-linux-firmware-arm64 /lib/firmware/bnx2x /rootfs/lib/firmware/bnx2x
COPY --from=pkg-lvm2-arm64 / /rootfs
COPY --from=pkg-libaio-arm64 / /rootfs
COPY --from=pkg-musl-arm64 / /rootfs
COPY --from=pkg-open-iscsi-arm64 / /rootfs
COPY --from=pkg-open-isns-arm64 / /rootfs
COPY --from=pkg-runc-arm64 / /rootfs
COPY --from=pkg-xfsprogs-arm64 / /rootfs
COPY --from=pkg-util-linux-arm64 /lib/libblkid.* /rootfs/lib/
COPY --from=pkg-util-linux-arm64 /lib/libuuid.* /rootfs/lib/
COPY --from=pkg-util-linux-arm64 /lib/libmount.* /rootfs/lib/
COPY --from=pkg-kmod-arm64 /usr/lib/libkmod.* /rootfs/lib/
COPY --from=pkg-kernel-arm64 /lib/modules /rootfs/lib/modules
COPY --from=machined-build-arm64 /machined /rootfs/sbin/init
# NB: We run the cleanup step before creating extra directories, files, and
# symlinks to avoid accidentally cleaning them up.
COPY ./hack/cleanup.sh /toolchain/bin/cleanup.sh
RUN cleanup.sh /rootfs
COPY hack/containerd.toml /rootfs/etc/cri/containerd.toml
RUN touch /rootfs/etc/resolv.conf
RUN touch /rootfs/etc/hosts
RUN touch /rootfs/etc/os-release
RUN mkdir -pv /rootfs/{boot,usr/local/share,mnt,system}
RUN mkdir -pv /rootfs/{etc/kubernetes/manifests,etc/cni,usr/libexec/kubernetes}
RUN ln -s /etc/ssl /rootfs/etc/pki
RUN ln -s /etc/ssl /rootfs/usr/share/ca-certificates
RUN ln -s /etc/ssl /rootfs/usr/local/share/ca-certificates
RUN ln -s /etc/ssl /rootfs/etc/ca-certificates

FROM rootfs-base-${TARGETARCH} AS rootfs-base

FROM rootfs-base-arm64 AS rootfs-squashfs-arm64
RUN find /rootfs -print0 \
    | xargs -0r touch --no-dereference --date="@${SOURCE_DATE_EPOCH}"
RUN mksquashfs /rootfs /rootfs.sqsh -all-root -noappend -comp xz -Xdict-size 100% -no-progress

FROM rootfs-base-amd64 AS rootfs-squashfs-amd64
RUN find /rootfs -print0 \
    | xargs -0r touch --no-dereference --date="@${SOURCE_DATE_EPOCH}"
RUN mksquashfs /rootfs /rootfs.sqsh -all-root -noappend -comp xz -Xdict-size 100% -no-progress

FROM scratch AS squashfs-arm64
COPY --from=rootfs-squashfs-arm64 /rootfs.sqsh /

FROM scratch AS squashfs-amd64
COPY --from=rootfs-squashfs-amd64 /rootfs.sqsh /

FROM scratch AS rootfs
COPY --from=rootfs-base /rootfs /

# The initramfs target provides the Talos initramfs image.

FROM build AS initramfs-archive-arm64
WORKDIR /initramfs
COPY --from=squashfs-arm64 /rootfs.sqsh .
COPY --from=init-build-arm64 /init .
RUN find . -print0 \
    | xargs -0r touch --no-dereference --date="@${SOURCE_DATE_EPOCH}"
RUN set -o pipefail \
    && find . 2>/dev/null \
    | LC_ALL=c sort \
    | cpio --reproducible -H newc -o \
    | xz -v -C crc32 -0 -e -T 0 -z \
    > /initramfs.xz

FROM build AS initramfs-archive-amd64
WORKDIR /initramfs
COPY --from=squashfs-amd64 /rootfs.sqsh .
COPY --from=init-build-amd64 /init .
RUN find . -print0 \
    | xargs -0r touch --no-dereference --date="@${SOURCE_DATE_EPOCH}"
RUN set -o pipefail \
    && find . 2>/dev/null \
    | LC_ALL=c sort \
    | cpio --reproducible -H newc -o \
    | xz -v -C crc32 -0 -e -T 0 -z \
    > /initramfs.xz

FROM initramfs-archive-${TARGETARCH} AS initramfs-archive

FROM scratch AS initramfs
ARG TARGETARCH
COPY --from=initramfs-archive /initramfs.xz /initramfs-${TARGETARCH}.xz

# The talos target generates a docker image that can be used to run Talos
# in containers.

FROM scratch AS talos
COPY --from=rootfs / /
LABEL org.opencontainers.image.source https://github.com/talos-systems/talos
ENTRYPOINT ["/sbin/init"]

# The installer target generates an image that can be used to install Talos to
# various environments.

FROM base AS installer-build
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
WORKDIR /src/cmd/installer
ARG TARGETARCH
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=${TARGETARCH} go build ${GO_BUILDFLAGS} -ldflags "${GO_LDFLAGS}" -o /installer
RUN chmod +x /installer

FROM alpine:3.14.0 AS unicode-pf2
RUN apk add --no-cache --update grub

FROM scratch AS install-artifacts-amd64
COPY --from=pkg-grub-amd64 /usr/lib/grub /usr/lib/grub
COPY --from=pkg-kernel-amd64 /boot/vmlinuz /usr/install/amd64/vmlinuz
COPY --from=pkg-kernel-amd64 /dtb /usr/install/amd64/dtb
COPY --from=initramfs-archive-amd64 /initramfs.xz /usr/install/amd64/initramfs.xz

FROM scratch AS install-artifacts-arm64
COPY --from=pkg-grub-arm64 /usr/lib/grub /usr/lib/grub
COPY --from=pkg-kernel-arm64 /boot/vmlinuz /usr/install/arm64/vmlinuz
COPY --from=pkg-kernel-arm64 /dtb /usr/install/arm64/dtb
COPY --from=initramfs-archive-arm64 /initramfs.xz /usr/install/arm64/initramfs.xz
COPY --from=pkg-u-boot-arm64 / /usr/install/arm64/u-boot
COPY --from=pkg-raspberrypi-firmware-arm64 / /usr/install/arm64/raspberrypi-firmware

FROM scratch AS install-artifacts-all
COPY --from=install-artifacts-amd64 / /
COPY --from=install-artifacts-arm64 / /

FROM install-artifacts-${TARGETARCH} AS install-artifacts-targetarch

FROM install-artifacts-${INSTALLER_ARCH} AS install-artifacts
COPY --from=pkg-grub / /
COPY --from=unicode-pf2 /usr/share/grub/unicode.pf2 /usr/share/grub/unicode.pf2

FROM alpine:3.14.0 AS installer
RUN apk add --no-cache --update \
    bash \
    ca-certificates \
    efibootmgr \
    mtools \
    qemu-img \
    util-linux \
    xfsprogs \
    xorriso \
    xz
ARG TARGETARCH
ENV TARGETARCH ${TARGETARCH}
COPY --from=install-artifacts / /
COPY --from=installer-build /installer /bin/installer
RUN ln -s /bin/installer /bin/talosctl
ARG TAG
ENV VERSION ${TAG}
LABEL "alpha.talos.dev/version"="${VERSION}"
LABEL org.opencontainers.image.source https://github.com/talos-systems/talos
ENTRYPOINT ["/bin/installer"]
ONBUILD RUN apk add --no-cache --update \
    cpio \
    squashfs-tools \
    xz
ONBUILD WORKDIR /initramfs
ONBUILD ARG RM
ONBUILD RUN xz -d /usr/install/${TARGETARCH}/initramfs.xz \
    && cpio -idvm < /usr/install/${TARGETARCH}/initramfs \
    && unsquashfs -f -d /rootfs rootfs.sqsh \
    && for f in ${RM}; do rm -rfv /rootfs$f; done \
    && rm /usr/install/${TARGETARCH}/initramfs \
    && rm rootfs.sqsh
ONBUILD COPY --from=customization / /rootfs
ONBUILD RUN find /rootfs \
    && mksquashfs /rootfs rootfs.sqsh -all-root -noappend -comp xz -Xdict-size 100% -no-progress \
    && set -o pipefail && find . 2>/dev/null | cpio -H newc -o | xz -v -C crc32 -0 -e -T 0 -z >/usr/install/${TARGETARCH}/initramfs.xz \
    && rm -rf /rootfs \
    && rm -rf /initramfs
ONBUILD WORKDIR /

# The test target performs tests on the source code.

FROM base AS unit-tests-runner
RUN unlink /etc/ssl
COPY --from=rootfs / /
ARG TESTPKGS
ENV PLATFORM container
RUN --security=insecure --mount=type=cache,id=testspace,target=/tmp --mount=type=cache,target=/.cache go test -v -covermode=atomic -coverprofile=coverage.txt -coverpkg=${TESTPKGS} -count 1 -p 4 ${TESTPKGS}
FROM scratch AS unit-tests
COPY --from=unit-tests-runner /src/coverage.txt /coverage.txt

# The unit-tests-race target performs tests with race detector.

FROM base AS unit-tests-race
RUN unlink /etc/ssl
COPY --from=rootfs / /
ARG TESTPKGS
ENV PLATFORM container
ENV CGO_ENABLED 1
RUN --security=insecure --mount=type=cache,id=testspace,target=/tmp --mount=type=cache,target=/.cache go test -v -race -count 1 -p 4 ${TESTPKGS}

# The integration-test targets builds integration test binary.

FROM base AS integration-test-linux-build
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=amd64 go test -v -c ${GO_BUILDFLAGS} \
    -ldflags "${GO_LDFLAGS}" \
    -tags integration,integration_api,integration_cli,integration_k8s \
    ./internal/integration

FROM scratch AS integration-test-linux
COPY --from=integration-test-linux-build /src/integration.test /integration-test-linux-amd64

FROM base AS integration-test-darwin-build
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=darwin GOARCH=amd64 go test -v -c ${GO_BUILDFLAGS} \
    -ldflags "${GO_LDFLAGS}" \
    -tags integration,integration_api,integration_cli,integration_k8s \
    ./internal/integration

FROM scratch AS integration-test-darwin
COPY --from=integration-test-darwin-build /src/integration.test /integration-test-darwin-amd64

# The integration-test-provision target builds integration test binary with provisioning tests.

FROM base AS integration-test-provision-linux-build
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
RUN --mount=type=cache,target=/.cache GOOS=linux GOARCH=amd64 go test -v -c ${GO_BUILDFLAGS} \
    -ldflags "${GO_LDFLAGS}" \
    -tags integration,integration_provision \
    ./internal/integration

FROM scratch AS integration-test-provision-linux
COPY --from=integration-test-provision-linux-build /src/integration.test /integration-test-provision-linux-amd64

# The lint target performs linting on the source code.

FROM base AS lint-go
COPY .golangci.yml .
ENV GOGC 50
ENV GOLANGCI_LINT_CACHE /.cache/lint
RUN --mount=type=cache,target=/.cache golangci-lint run --config .golangci.yml
WORKDIR /src/pkg/machinery
RUN --mount=type=cache,target=/.cache golangci-lint run --config ../../.golangci.yml
WORKDIR /src
RUN --mount=type=cache,target=/.cache importvet github.com/talos-systems/talos/...
RUN find . -name '*.pb.go' -o -name '*_string_*.go' | xargs rm
RUN --mount=type=cache,target=/.cache FILES="$(gofumports -l -local github.com/talos-systems/talos .)" && test -z "${FILES}" || (echo -e "Source code is not formatted with 'gofumports -w -local github.com/talos-systems/talos .':\n${FILES}"; exit 1)

# The protolint target performs linting on protobuf files.

FROM base AS lint-protobuf
WORKDIR /src/api
COPY api .
COPY prototool.yaml .
RUN prototool lint --protoc-bin-path=/toolchain/bin/protoc --protoc-wkt-path=/toolchain/include

# The markdownlint target performs linting on Markdown files.

FROM node:16.4.2-alpine AS lint-markdown
RUN apk add --no-cache findutils
RUN npm i -g markdownlint-cli@0.23.2
RUN npm i -g textlint@11.7.6
RUN npm i -g textlint-filter-rule-comments@1.2.2
RUN npm i -g textlint-rule-one-sentence-per-line@1.0.2
WORKDIR /src
COPY . .
RUN markdownlint \
    --ignore '**/LICENCE.md' \
    --ignore '**/CHANGELOG.md' \
    --ignore '**/CODE_OF_CONDUCT.md' \
    --ignore '**/node_modules/**' \
    --ignore '**/hack/chglog/**' \
    --ignore 'website/content/docs/*/Reference/*' \
    .
RUN find . \
    -name '*.md' \
    -not -path './LICENCE.md' \
    -not -path './CHANGELOG.md' \
    -not -path './CODE_OF_CONDUCT.md' \
    -not -path '*/node_modules/*' \
    -not -path './hack/chglog/**' \
    -not -path './website/content/docs/*/Reference/*' \
    -print0 \
    | xargs -0 textlint

# The docs target generates documentation.

FROM base AS docs-build
WORKDIR /src
COPY --from=talosctl-linux /talosctl-linux-amd64 /bin/talosctl
RUN env HOME=/home/user TAG=latest /bin/talosctl docs --config /tmp \
    && env HOME=/home/user TAG=latest /bin/talosctl docs --cli /tmp

FROM pseudomuto/protoc-gen-doc as proto-docs-build
COPY --from=generate-build /api /protos
COPY ./hack/protoc-gen-doc/markdown.tmpl /tmp/markdown.tmpl
RUN protoc \
    -I/protos \
    -I/protos/common \
    -I/protos/inspect \
    -I/protos/machine \
    -I/protos/network \
    -I/protos/resource \
    -I/protos/security \
    -I/protos/storage \
    -I/protos/time \
    -I/protos/vendor \
    --doc_opt=/tmp/markdown.tmpl,api.md \
    --doc_out=/tmp \
    /protos/common/*.proto \
    /protos/inspect/*.proto \
    /protos/machine/*.proto \
    /protos/network/*.proto \
    /protos/resource/*.proto \
    /protos/security/*.proto \
    /protos/storage/*.proto \
    /protos/time/*.proto

FROM scratch AS docs
COPY --from=docs-build /tmp/configuration.md /website/content/docs/v0.12/Reference/
COPY --from=docs-build /tmp/cli.md /website/content/docs/v0.12/Reference/
COPY --from=proto-docs-build /tmp/api.md /website/content/docs/v0.12/Reference/

# The talosctl-cni-bundle builds the CNI bundle for talosctl.

FROM scratch AS talosctl-cni-bundle
ARG TARGETARCH
COPY --from=extras-talosctl-cni-bundle-install /opt/cni/bin/ /talosctl-cni-bundle-${TARGETARCH}/
