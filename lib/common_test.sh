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
# shellcheck source=./lib/testing.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/testing.sh"

# shellcheck source=./lib/common.sh
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

TEST_validate_command_line() {
  tmpDir="$( mktemp -d )"
  trap 'rm -rf -- "$tmpDir"' EXIT

  # shellcheck disable=SC2034
  PROGSTATE="${tmpDir}/whats-a-progstate-even.txt" \
    LOGFILE="${tmpDir}/some-log-file.log" \
    HR='' \
    TPUT[BOLD]='' \
    TPUT[RED]='' \
    TPUT[OFF]=''

  local one='./anago   release-1.12 --stage --prebuild --build-at-head  --official   --yes --gcb --basedir=/workspace --tmpdir=/workspace/tmp'
  local two='./anago release-1.12 --stage --buildonly --build-at-head  --official   --yes --gcb --basedir=/workspace --tmpdir=/workspace/tmp'
  local tre='   ./anago release-1.13 --stage --buildonly --build-at-head  --official   --yes --gcb --basedir=/workspace --tmpdir=/workspace/tmp'

  common::askyorn() {
    echo 'common::askyorn would have been called ...'
    return 1
  }

  local nl=$'\n'
  local expectedOutput="${nl}"
  expectedOutput+="1st run${nl}"
  expectedOutput+="${nl}"
  expectedOutput+="Continuing previous session${nl}"
  expectedOutput+="(${PROGSTATE}).${nl}"
  expectedOutput+="Use --clean to restart${nl}"
  expectedOutput+="${nl}"
  expectedOutput+="2nd run${nl}"
  expectedOutput+="${nl}"
  expectedOutput+="A previous incomplete run using different command-line values exists.${nl}"
  expectedOutput+="  current  (relevant) args: ./anago release-1.13 --stage --build-at-head --official --yes --gcb --basedir=/workspace --tmpdir=/workspace/tmp${nl}"
  expectedOutput+="  previous (relevant) args: ./anago release-1.12 --stage --build-at-head --official --yes --gcb --basedir=/workspace --tmpdir=/workspace/tmp${nl}"
  expectedOutput+="${nl}"
  expectedOutput+="Did you mean to --clean and start a new session?${nl}"
  expectedOutput+="${nl}"
  expectedOutput+="common::askyorn would have been called ..."

  assertEqualContent \
    <(
      common::validate_command_line "$one" ; echo '1st run'
      common::validate_command_line "$two" ; echo '2nd run'
      common::validate_command_line "$tre" ; echo '3rd run'
    ) \
    <( echo "$expectedOutput" ) \
    "should only fail when relevant arguments differ, but not if they are the same"

}

testMain "$@"
