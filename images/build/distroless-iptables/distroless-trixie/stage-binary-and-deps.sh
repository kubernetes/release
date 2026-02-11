#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
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

# USAGE: stage-binary-and-deps.sh haproxy /opt/stage
#
# Stages $1 and its dependencies + their copyright files to $2
#
# This is intended to be used in a multi-stage docker build with a distroless/base
# or distroless/cc image.
# This script was originally created by KinD maintainers and can be found at:
#   https://github.com/kubernetes-sigs/kind/blob/v0.14.0/images/haproxy/stage-binary-and-deps.sh

set -o errexit
set -o nounset
set -o pipefail


. package-utils.sh

# binary_to_libraries identifies the library files needed by the binary $1 with ldd
binary_to_libraries() {
    # see: https://man7.org/linux/man-pages/man1/ldd.1.html
    ldd "${1}" \
    `# strip the leading '${name} => ' if any so only '/lib-foo.so (0xf00)' remains` \
    | sed -E 's#.* => /#/#' \
    `# we want only the path remaining, not the (0x${LOCATION})` \
    | awk '{print $1}' \
    `# linux-vdso.so.1 is a special virtual shared object from the kernel` \
    `# see: http://man7.org/linux/man-pages/man7/vdso.7.html` \
    | grep -v 'linux-vdso.so.1'
}

# main script logic
main(){
    local STAGE_DIR="${1}/"
    shift
    while (( "$#" )); do
        BINARY="${1}"
        # locate the path to the binary
        local binary_path
        binary_path="$(which "${BINARY}")"
    
        # ensure package metadata dir
        mkdir -p "${STAGE_DIR}"/var/lib/dpkg/status.d/
    
        # stage the binary itself
        stage_file "${binary_path}" "${STAGE_DIR}"
    
        # stage the dependencies of the binary
        while IFS= read -r c_dep; do
            stage_file "${c_dep}" "${STAGE_DIR}"
        done < <(binary_to_libraries "${binary_path}")
        shift
    done
}

main "$@"
