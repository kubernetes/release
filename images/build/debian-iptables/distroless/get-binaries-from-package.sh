#!/usr/bin/env sh

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

get_binary_list() {
    BIN_LIST="$(dpkg -L "${1}" | grep -E '/s?bin/' |sed 's/\n/ /g')"
    for binary in $BIN_LIST; do
        /stage-binary-and-deps.sh "${binary}" "${2}"
    done
}

main() {
    PACKAGE=$1
    STAGE_DIR="${2}/"
    apt -y update
    apt -y install "${PACKAGE}"
    get_binary_list "${PACKAGE}" "$STAGE_DIR"
}

main "$@"
