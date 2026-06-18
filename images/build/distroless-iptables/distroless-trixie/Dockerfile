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

COPY stage-binary-and-deps.sh /
COPY stage-binaries-from-package.sh /
COPY package-utils.sh /

# 1. Install compilation toolchain and dependencies
RUN apt-get -y update && \
    apt-get -y dist-upgrade && \
    apt-get -y install --no-install-recommends \
        git ca-certificates build-essential autoconf automake libtool \
        bison flex pkg-config libgmp-dev libreadline-dev \
        libedit-dev libjansson-dev

# 2. Build libmnl
ARG LIBMNL_VERSION="libmnl-1.0.5"
ARG LIBMNL_CLONE_URL="https://git.netfilter.org/libmnl"
RUN git clone --filter=tree:0 "${LIBMNL_CLONE_URL}" /libmnl && \
    cd /libmnl && \
    git checkout "${LIBMNL_VERSION}" && \
    ./autogen.sh && ./configure --enable-static --disable-shared && make && make install

# 3. Build libnftnl
ARG LIBNFTNL_VERSION="libnftnl-1.2.9"
ARG LIBNFTNL_CLONE_URL="https://git.netfilter.org/libnftnl"
RUN git clone --filter=tree:0 "${LIBNFTNL_CLONE_URL}" /libnftnl && \
    cd /libnftnl && \
    git checkout "${LIBNFTNL_VERSION}" && \
    ./autogen.sh && ./configure --enable-static --disable-shared && make && make install

# 4. Build nftables
ARG NFTABLES_VERSION="v1.1.3"
ARG NFTABLES_CLONE_URL="https://git.netfilter.org/nftables"
COPY patches /patches
RUN git clone --filter=tree:0 "${NFTABLES_CLONE_URL}" /nftables && \
    cd /nftables && \
    git checkout "${NFTABLES_VERSION}" && \
    for p in $(cat /patches/series); do patch -p1 --forward < "/patches/$p" || true; done && \
    ./autogen.sh && \
    ./configure --disable-python --disable-man-doc --sbindir=/usr/sbin --enable-static --disable-shared && \
    make && make install

RUN ldconfig

# TODO: /bin/sleep is used by a specific e2e test
ARG STAGE_DIR="/opt/stage"
RUN apt-get -y install bash curl && \
    mkdir -p "${STAGE_DIR}" && \
    echo "libmnl ${LIBMNL_VERSION} $(git -C /libmnl rev-parse HEAD)" > "${STAGE_DIR}/nftables-source.txt" && \
    echo "libnftnl ${LIBNFTNL_VERSION} $(git -C /libnftnl rev-parse HEAD)" >> "${STAGE_DIR}/nftables-source.txt" && \
    echo "nftables ${NFTABLES_VERSION} $(git -C /nftables rev-parse HEAD)" >> "${STAGE_DIR}/nftables-source.txt" && \
    /stage-binaries-from-package.sh "${STAGE_DIR}" ipset \
    iptables    \
    kmod        && \
    /stage-binary-and-deps.sh "${STAGE_DIR}" /usr/sbin/nft && \
    if ldd "${STAGE_DIR}/usr/sbin/nft" | grep -qE '(libnftables|libnftnl|libmnl)'; then echo "Error: nft is dynamically linked to netfilter libraries" && exit 1; fi && \
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

# We need to use distroless:static here as tzdata, and some other base packages are required.
# We supply our own C libraries alongside some additional C binaries.
FROM gcr.io/distroless/static-debian13 as intermediate

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
