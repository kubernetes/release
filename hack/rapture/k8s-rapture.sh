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

# This is a temporary stopgap to add some automation to the Googler-only process.
# In the long term, we want to develop a different process that non-Googlers
# can perform.
#
# It builds and pushes both Debian and Red Hat packages, pausing for
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
#   k8s-rapture.sh 1.6.12

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

# TODO(mehdy) checks:
# 1. Check if on release repo
# 2. Check for prod access
# 3. Check if release repo up to date
# 4. Check if rpm and any other required command exists

# TODO(mehdy): Restructure the script into three phases:
# P1 - Building RPMs and DEBs
# P2 - Signing all packages
# P3 - Publish them.

# TODO(mehdy): Make the script rerunnable at any phase

#############################################################################
# Simple yes/no prompt
#
# @optparam default -n(default)/-y/-e (default to n, y or make (e)xplicit)
# @param message
askyorn () {
  local yorn=z
  local def=n
  local msg="y/N"

  case $1 in
  -y) # yes default
      def="y" msg="Y/n"
      shift
      ;;
  -e) # Explicit
      def="" msg="y/n"
      shift
      ;;
  -n) shift
      ;;
  esac

  while [[ $yorn != [yYnN] ]]; do
    echo -n "$*? ($msg): "
    read yorn
    : ${yorn:=$def}
  done

  # Final test to set return code
  [[ $yorn == [yY] ]]
}

[[ -n "$1" ]] || fatal "no version specified"

askyorn "Continue RPMs and DEBs release for $1" || exit 1

version="$1"
"${RELEASE_ROOT}"/hack/rapture/build-packages.sh "$version"
"${RELEASE_ROOT}"/hack/rapture/publish-packages.sh "$version"
