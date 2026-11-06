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
FROM ${BASEIMAGE} AS build

# Install compilation toolchain and dependencies
RUN apt-get -y update && \
    apt-get -y dist-upgrade && \
    apt-get -y install --no-install-recommends \
        git ca-certificates build-essential autoconf automake libtool \
        bison flex pkg-config libgmp-dev libreadline-dev \
        libedit-dev libjansson-dev \
        libmnl-dev libnftnl-dev libxtables-dev

# Build nftables
ADD --checksum=sha256:bef0c9cfdca5f8b988957046c2cb33ef9869730593da0eacae4748201acf1116 https://netfilter.org/projects/nftables/files/nftables-1.0.6.1.tar.xz /
COPY patches /patches
COPY setup-metadata.sh /setup-metadata.sh
RUN cd / && tar xf /nftables-1.0.6.1.tar.xz && cd nftables-1.0.6.1 && \
    for p in $(cat /patches/series); do patch -p1 --forward < "/patches/$p" || exit 1; done && \
    ./configure --with-json --with-xtables --disable-man-doc --enable-static --disable-shared --sbindir=/usr/sbin && make && make install && \
    /setup-metadata.sh

RUN ldconfig

WORKDIR /
COPY stage-binary-and-deps.sh /
COPY stage-binaries-from-package.sh /
COPY package-utils.sh /

# Stage iptables (for iptables and ipvs modes), ipset and modprobe (for ipvs) and nft (for
# nftables mode). The symlinks created by update-alternatives aren't considered part of
# the iptables package so we have to stage them by hand. (k/k requires at least
# `/usr/sbin/iptables` to be present as part of mode auto-detection, and users may expect
# to be able to `kubectl exec` the other binaries for debugging.)
ARG STAGE_DIR="/opt/stage"
RUN mkdir -p "${STAGE_DIR}" && \
    /stage-binaries-from-package.sh "${STAGE_DIR}" ipset iptables kmod && \
    /stage-binary-and-deps.sh "${STAGE_DIR}" /usr/sbin/iptables /usr/sbin/iptables-save /usr/sbin/iptables-restore /usr/sbin/ip6tables /usr/sbin/ip6tables-save /usr/sbin/ip6tables-restore && \
    /stage-binary-and-deps.sh "${STAGE_DIR}" /usr/sbin/nft

# we're going to borrow the /go-runner binary in the final image
# dedupe this binary by just copying it from the go-runner image
FROM ${GORUNNERIMAGE} AS gorunner

# We need to use distroless:static here, as tzdata and some other base packages are required.
# We supply our own C libraries alongside some additional C binaries.
FROM gcr.io/distroless/static-debian13 AS intermediate

COPY --from=build /opt/stage /
COPY --from=gorunner /go-runner /go-runner

FROM scratch
COPY --from=intermediate / /
