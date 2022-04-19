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

ARG GORUNNERIMAGE
# Using directly debian base, as we are going to just copy the required binaries
FROM debian:bullseye as build

COPY stage-binary-and-deps.sh /
COPY get-binaries-from-package.sh /

ARG STAGE_DIR="/opt/stage"
RUN apt -y update && \
    apt -y install bash curl && \
    mkdir -p "${STAGE_DIR}" && \
    /get-binaries-from-package.sh conntrack "${STAGE_DIR}" && \
    /get-binaries-from-package.sh ebtables "${STAGE_DIR}" && \
    /get-binaries-from-package.sh ipset "${STAGE_DIR}" && \
    /get-binaries-from-package.sh iptables "${STAGE_DIR}" && \
    /get-binaries-from-package.sh kmod "${STAGE_DIR}" && \
    `# dash is the real Debian basic shell` \
    /stage-binary-and-deps.sh /bin/dash "${STAGE_DIR}" && \
    `# below binaries are used by iptables-wrapper-installer.sh` \
    /stage-binary-and-deps.sh /bin/cat "${STAGE_DIR}" && \
    /stage-binary-and-deps.sh /bin/chmod "${STAGE_DIR}" && \
    /stage-binary-and-deps.sh /bin/grep "${STAGE_DIR}" && \
    /stage-binary-and-deps.sh /bin/ln "${STAGE_DIR}" && \
    /stage-binary-and-deps.sh /bin/rm "${STAGE_DIR}" && \
    /stage-binary-and-deps.sh /usr/bin/wc "${STAGE_DIR}" && \
    `# iptables modules needs to be copied` \
    cp -r /usr/lib/"$(uname -m)"-linux-gnu/xtables/ "${STAGE_DIR}"/usr/lib/"$(uname -m)"-linux-gnu/ && \
    curl -o "${STAGE_DIR}"/iptables-wrapper-installer.sh https://raw.githubusercontent.com/kubernetes-sigs/iptables-wrappers/master/iptables-wrapper-installer.sh && \
    chmod +x "${STAGE_DIR}"/iptables-wrapper-installer.sh && \
    ln -sf /bin/dash "${STAGE_DIR}"/bin/sh

# we're going to borrow the /go-runner binary in the final image
# dedupe this binary by just copying it from the go-runner image
FROM ${GORUNNERIMAGE} as gorunner

# We need to use distroless:base here as tzdata, glibc and some other base packages
# are required
FROM gcr.io/distroless/base as intermediate
# FROM debian:bullseye
COPY clean-distroless.sh /clean-distroless.sh
COPY --from=build /opt/stage /
COPY --from=gorunner /go-runner /go-runner
# iptables-wrapper-installer needs to know that iptables exists before doing all its magic
RUN ln -sf /usr/sbin/iptables-legacy /usr/sbin/iptables && \
    /iptables-wrapper-installer.sh && \
    /clean-distroless.sh

FROM scratch
COPY --from=intermediate / /
