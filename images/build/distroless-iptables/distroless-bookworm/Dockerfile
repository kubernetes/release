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

# TODO: /bin/sleep is used by a specific e2e test
ARG STAGE_DIR="/opt/stage"
RUN apt -y update && \
    apt -y dist-upgrade && \
    apt -y install bash curl && \
    mkdir -p "${STAGE_DIR}" && \
    /stage-binaries-from-package.sh "${STAGE_DIR}" ebtables \
    ipset       \
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
