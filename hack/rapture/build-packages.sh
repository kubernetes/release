#!/usr/bin/env bash

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

# This script builds the Kubernetes debian and RPM packages.
#
# Example usage:
#   git clone https://github.com/kubernetes/release.git
#   cd release
#   build-packages.sh 1.6.12

set -o errexit
set -o nounset
set -o pipefail

RELEASE_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

log() {
  echo "$*" >&2
}

fatal() {
  log "FATAL: $*"
  exit 1
}


[[ -n "$1" ]] || fatal "no version specified"
version="$1"


build_debs() {
  local distro=xenial

  log "Clearing output dir"
  DEBDIR="${RELEASE_ROOT:?}"/packages/deb
  rm -rf "${DEBDIR:?}"/bin
  cd "${DEBDIR:?}"

  log "Setting all Revisions to \"00\" in build.go"
  sed -i -r -e 's/\b(Revision:\s*)"[0-9]{2}"/\1"00"/' build.go

  log "Building debs for Kubernetes v${version}"
  ./jenkins.sh --kube-version $version --distros $distro

  log "Changing file owner from root to ${USER}"
  sudo chown -R "${USER}" bin

  cd "${RELEASE_ROOT:?}"
}

build_rpms() {
  local distro=el7
  local keyfile
  local RPMDIR

  log "Clearing output dir"
  RPMDIR="${RELEASE_ROOT:?}"/packages/rpm
  rm -rf "${RPMDIR:?}"/output
  cd "${RPMDIR:?}"

  log "Setting version in kubelet.spec"
  local vparts=(${version//./ })
  sed -i -r \
    -e "s/(%global\\s+KUBE_MAJOR\\s+)[0-9]+/\\1${vparts[0]}/" \
    -e "s/(%global\\s+KUBE_MINOR\\s+)[0-9]+/\\1${vparts[1]}/" \
    -e "s/(%global\\s+KUBE_PATCH\\s+)[0-9]+/\\1${vparts[2]}/" \
    -e "s/(%global\\s+RPM_RELEASE\\s+)[0-9]+/\\10/" \
    kubelet.spec

  log "Building RPMs for Kubernetes v${version}"
  ./docker-build.sh
  cd "${RELEASE_ROOT:?}"
}

build_debs
build_rpms
