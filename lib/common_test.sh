#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

# common.sh unit tests
#
# shellcheck source=../lib/testing.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/testing.sh"

# shellcheck source=../lib/common.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/common.sh"
readonly TESTDATA="$( cd "$(dirname "$0")" && pwd )/testdata"

set -e
set -o pipefail

TEST_run_stateful() {
  tmpDir="$( mktemp -d )"
  trap 'rm -rf -- "$tmpDir"' EXIT

  # override some vars and func to not clutter output
  common::timestamp() { :; }
  # shellcheck disable=SC2034
  PROGSTATE="${tmpDir}/whats-a-progstate-even.txt" \
    LOGFILE="${tmpDir}/some-log-file.log" \
    HR='' \
    TPUT[BOLD]='' \
    TPUT[OFF]=''

  assertEqualContent \
    <( common::run_stateful --strip-args 'printf %s\n%s arg1 arg2' ) \
    <( echo -e "\n\nprintf\n\n\narg1\narg2" ) \
    "passing command and arguments"
}

testMain "$@"
