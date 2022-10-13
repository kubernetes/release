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

# This is a temporary stopgap to add some automation to the Googler-only process.
# In the long term, we want to develop a different process that non-Googlers
# can perform.
#
# It signs and pushes both Debian and Red Hat packages, pausing for
# confirmation before promoting each repo to stable. It always sets the
# package revision to 0, so you should only use it for a new minor or patch
# release for which packages have never been pushed.
#
# This is based on a copy of the existing script from ex-Googler
# mehdy@google.com, with some modifications to reduce the number of manual
# prompts, which can potentially derail a release if the wrong option is chosen
# by the user accidentally.
#
# Example usage:
#   git clone https://github.com/kubernetes/release.git
#   cd release
#   publish-packages.sh 1.6.12
#
# NOTE: this currently requires build-packages.sh to be run locally first.

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

publish_debs() {
  local distro=xenial
  local DEBDIR
  DEBDIR="${RELEASE_ROOT:?}"/packages/deb
  cd "${DEBDIR:?}"

  log "Removing local debs that already exist on the server"
  local debpath
  for debfile in $(rapture --universe=cloud-apt listrepo "kubernetes-${distro}" | sed -r -e 's/^(\S+) (\S+) \[(\S+)\] .*$/\1_\2_\3.deb/'); do
    debpath="bin/stable/${distro}/${debfile}"
    if [[ -f "${debpath}" ]]; then
      log "Removing ${debpath}"
      rm "${debpath}"
    fi
  done
  if ls bin/stable/${distro}/*.deb 1> /dev/null 2>&1;
  then
    log "Pushing debs to kubernetes-${distro}-unstable"
    rapture --universe=cloud-apt addpkg -keepold "kubernetes-${distro}" bin/stable/${distro}/*.deb
  else
    log "No debs found in bin/stable/${distro}/*.deb, skipping rapture addpkg"
  fi

  log "Packages in kubernetes-${distro}-unstable:"
  rapture --universe=cloud-apt listrepo "kubernetes-${distro}-unstable"
  local target
  target=$(rapture --universe=cloud-apt showrepo "kubernetes-${distro}-unstable" | awk '/ Current indirection:/ {print $3}')
  log
  log "Promoting ${target} to stable"
  rapture --universe=cloud-apt settag "${target}" cloud-kubernetes-release.stable:true

  cd "${RELEASE_ROOT:?}"
}

publish_rpms() {
  local distro=el7
  local keyfile
  local RPMDIR

  RPMDIR="${RELEASE_ROOT:?}"/packages/rpm
  cd "${RPMDIR:?}"

  log "Signing RPMs"
  pushd /google/src/head/depot/google3/cloud/cluster/guest/cloud_rapture/rpmsign
  for rpmfile in $(find "${RPMDIR}" -name '*.rpm'); do
    ./sign.sh "${rpmfile}"
  done
  popd

  log "Importing key to check RPM signatures"
  pushd /tmp
  keyfile="$(mktemp)"
  curl -sL https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg > ${keyfile}
  rpm --import "${keyfile}"
  rm "${keyfile}"
  popd

  log "Checking RPM signatures"
  local check
  for rpmfile in $(find "${RPMDIR}" -name '*.rpm'); do
    check="$(rpm -Kv "${rpmfile}")"
    echo "${check}"
    echo "${check}" | egrep -q '\bOK\b' || fatal "RPM signing failed; check output doesn't contain OK: ${check}"
    echo "${check}" | egrep -q '\bV4 RSA/SHA512 Signature\b' || fatal "RPM signing failed; check output doesn't contain pgp: ${check}"
    echo "${check}" | egrep -vq 'NOT OK' || fatal "RPM signing failed; check output contains NOT OK: ${check}"
    echo "${check}" | egrep -vq 'MISSING KEYS' || fatal "RPM signing failed; check output contains MISSING KEYS: ${check}"
  done

  log "Removing local RPMs that already exist on the server"
  local archlist=()
  local arch
  local rpmfile
  local rpmpath
  mapfile -t archlist < <(find output/* -maxdepth 0 -type d -printf "%f\n")
  for arch in "${archlist[@]}"; do
    for rpmfile in $(rapture --universe=cloud-yum listrepo "kubernetes-${distro}-${arch}" | sed -r -e 's/^(\S+) [0-9]+:(\S+) \[(\S+)\] .*$/\1-\2.\3.rpm/'); do
      rpmpath="output/${arch}/${rpmfile}"
      if [[ -f "${rpmpath}" ]]; then
        log "Removing ${rpmpath}"
        rm "${rpmpath}"
      fi
    done
  done

  local repo
  for arch in "${archlist[@]}"; do
    repo="kubernetes-${distro}-${arch}"
    if ls output/${arch}/*.rpm 1> /dev/null 2>&1;
    then
      log "Pushing RPMs to ${repo}-unstable"
      rapture --universe=cloud-yum addpkg -keepold "${repo}" output/${arch}/*.rpm
    else
      log "No RPMs found in output/${arch}/ skipping rapture addpkg"
    fi
  done

  local target
  for arch in "${archlist[@]}"; do
    repo="kubernetes-${distro}-${arch}"
    log "Packages in ${repo}-unstable:"
    rapture --universe=cloud-yum listrepo "${repo}-unstable"
    target=$(rapture --universe=cloud-yum showrepo "${repo}-unstable" | awk '/ Current indirection:/ {print $3}')
    log
    log "Promoting ${target} to stable ${repo}"
    rapture --universe=cloud-yum settag "${target}" cloud-kubernetes-release.stable:true
  done
  cd "${RELEASE_ROOT:?}"
}

publish_debs
publish_rpms
