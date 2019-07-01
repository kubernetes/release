#!/usr/bin/env bash
#
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
