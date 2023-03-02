#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

VERSION=0.4.1

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
BIN_PATH="$REPO_ROOT/bin"
ZEITGEIST_BIN="$BIN_PATH/zeitgeist"

if [[ ! -f "$ZEITGEIST_BIN" ]]; then
    mkdir -p "$BIN_PATH"
    curl -sSfL -o "$ZEITGEIST_BIN" \
        https://github.com/kubernetes-sigs/zeitgeist/releases/download/v$VERSION/zeitgeist_${VERSION}_linux_amd64
    chmod +x "$ZEITGEIST_BIN"
fi

"${ZEITGEIST_BIN}" validate \
    --local-only \
    --base-path "${REPO_ROOT}" \
    --config "${REPO_ROOT}"/dependencies.yaml
