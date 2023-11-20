#!/bin/bash

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

# USAGE: stage-binaries-from-package.sh /opt/stage package1 package2
#
# Stages all the packages and its dependencies (+ libraries and copyrights) to $1
#
# This is intended to be used in a multi-stage docker build with a distroless/base
# or distroless/cc image.
set -e

. package-utils.sh

stage_file_list() {
    IFS="
    "
    REQUIRED_FILES="$(dpkg -L "${1}" | grep -vE '(/\.|/s?bin/|/usr/share/(man|doc|.*-completion))' | sed 's/\n/ /g')"
    for file in $REQUIRED_FILES; do 
        if [ -f "$file" ]; then
            stage_file "${file}" "${STAGE_DIR}"
        fi
    done

    BIN_LIST="$(dpkg -L "${1}" | grep -E '/s?bin/' |sed 's/\n/ /g')"
    for binary in $BIN_LIST; do
        /stage-binary-and-deps.sh "${2}" "${binary}"
    done
}

get_dependent_packages() {
    apt-cache depends "${1}" |grep Depends|awk -F '.*Depends:[[:space:]]?' '{print $2}'
}

main() {
    STAGE_DIR="${1}/"
    mkdir -p "${STAGE_DIR}"/var/lib/dpkg/status.d/
    apt -y update
    shift
    while (( "$#" )); do        # While there are arguments still to be shifted
        PACKAGE="${1}"
        apt -y install "${PACKAGE}"
        stage_file_list "${PACKAGE}" "$STAGE_DIR"
        while IFS= read -r c_dep; do
            stage_file_list "${c_dep}" "${STAGE_DIR}"
        done < <(get_dependent_packages "${PACKAGE}")
        shift
    done
}

main "$@"
