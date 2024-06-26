#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
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

REPO_ROOT=$(git rev-parse --show-toplevel)
DIR="$( cd "${REPO_ROOT}" && pwd )"
WORK_DIR=$(mktemp -d -p "${DIR}")

if [[ -z "${WORK_DIR}" || ! -d "${WORK_DIR}" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

echo "Setting up environment..."
if [[ -z "${RELEASE_TOOL_BIN:-}" ]]; then
  if [[ -n "${GOBIN:-}" ]]; then
    export RELEASE_TOOL_BIN="${GOBIN}"
  else
    GOPATH=$(go env GOPATH)
    export RELEASE_TOOL_BIN="${GOPATH}/bin"
  fi
fi
export PATH="${PATH}:${RELEASE_TOOL_BIN}"

function cleanup {
  rm -rf "${WORK_DIR}"
  echo "Deleted temp working directory ${WORK_DIR}"
}

trap cleanup EXIT

cd "${WORK_DIR}"

dir_path="${REPO_ROOT}/cmd/krel/templates/latest"
if [[ -d  ${dir_path} ]]; then
  package_names=$(find "${dir_path}" -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)

  echo "Directory names in ${dir_path}:"
  echo "${package_names}"

  for package_name in ${package_names}; do
    echo "Processing Package: ${package_name}"
    krel obs specs --spec-only --package="${package_name}" --template-dir="${dir_path}"

    rpmlintrcArgs=''
    if [[ -f "${package_name}.rpmlintrc" ]]; then
      rpmlintrcArgs="-r=${package_name}.rpmlintrc"
    fi
    rpmlint "${package_name}".spec "${rpmlintrcArgs}" -v
  done
else
  echo "Directory ${dir_path} does not exist."
  exit 1
fi
