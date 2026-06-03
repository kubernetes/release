# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG BASEIMAGE
ARG GORUNNERIMAGE
# Using directly debian base, as we are going to just copy the required binaries
FROM ${BASEIMAGE} as build

# Build patched nftables package
ARG PKGNAME=nftables_1.0.6
RUN apt-get -y update && \
    apt-get -y dist-upgrade && \
    apt-get -y install --no-install-recommends \
        build-essential devscripts debhelper dh-make \
        bison flex libeditreadline-dev libgmp-dev libjansson-dev libmnl-dev libnftnl-dev libxtables-dev
WORKDIR /build
ADD --checksum=sha256:2407430ddd82987670e48dc2fda9e280baa8307abec04ab18d609df3db005e4c http://deb.debian.org/debian/pool/main/n/nftables/nftables_1.0.6.orig.tar.xz /build
RUN tar xf nftables_1.0.6.orig.tar.xz
COPY nftables/ nftables-1.0.6/
ENV DEBFULLNAME="Kubernetes Release Team"
ENV DEBEMAIL=release-team@kubernetes.io
WORKDIR /build/nftables-1.0.6
RUN debuild -us -uc

WORKDIR /
COPY stage-binary-and-deps.sh /
COPY stage-binaries-from-package.sh /
COPY package-utils.sh /

# TODO: /bin/sleep is used by a specific e2e test
ARG STAGE_DIR="/opt/stage"
RUN apt-get -y install bash curl && \
    apt-get -y -f install /build/nftables_*.deb /build/libnftables1_*.deb && \
    mkdir -p "${STAGE_DIR}" && \
    /stage-binaries-from-package.sh "${STAGE_DIR}" ipset \
    iptables    \
    nftables    \
    kmod        && \
    `# below binaries and dash are used by iptables-wrapper-installer.sh` \
    /stage-binary-and-deps.sh "${STAGE_DIR}" /bin/dash \
    /bin/cat \
    /bin/chmod \
    /bin/grep \
    /bin/ln  \
    /bin/rm \
    /bin/sleep \
    /usr/bin/wc

RUN curl -o "${STAGE_DIR}"/iptables-wrapper-installer.sh https://raw.githubusercontent.com/kubernetes-sigs/iptables-wrappers/v2/iptables-wrapper-installer.sh && \
    chmod +x "${STAGE_DIR}"/iptables-wrapper-installer.sh && \
    ln -sf /bin/dash "${STAGE_DIR}"/bin/sh

# we're going to borrow the /go-runner binary in the final image
# dedupe this binary by just copying it from the go-runner image
FROM ${GORUNNERIMAGE} as gorunner

# We need to use distroless:base here as tzdata, glibc and some other base packages
# are required
FROM gcr.io/distroless/base-debian12 as intermediate

ARG SKIP_WRAPPER_CHECK

COPY clean-distroless.sh /clean-distroless.sh
COPY --from=build /opt/stage /
COPY --from=gorunner /go-runner /go-runner
# iptables-wrapper-installer needs to know that iptables exists before doing all its magic
RUN echo "" >  /usr/sbin/iptables && \
    `# We skip sanity check in iptables-wrapper-installer due to qemu multiarch problems `\
    /iptables-wrapper-installer.sh ${SKIP_WRAPPER_CHECK} && \
    /clean-distroless.sh

FROM scratch
COPY --from=intermediate / /
