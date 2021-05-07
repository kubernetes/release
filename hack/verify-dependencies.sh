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

VERSION=v0.3.0
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# Ensure that we find the binaries we build before anything else.
gobin=${GOBIN:-$(go env GOBIN)}
if [[ -z $gobin ]]; then
  gobin="$(go env GOPATH)/bin"
fi
PATH="${gobin}:${PATH}"

# Install zeitgeist
cd "${REPO_ROOT}/internal"
GO111MODULE=on go install sigs.k8s.io/zeitgeist@"${VERSION}"
cd -

# Prefer full path for running zeitgeist
ZEITGEIST_BIN="$(which zeitgeist)"

"${ZEITGEIST_BIN}" validate \
  --local-only \
  --base-path "${REPO_ROOT}" \
  --config "${REPO_ROOT}"/dependencies.yaml
