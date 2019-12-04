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

#
# releaselib.sh unit tests
#
# shellcheck source=./lib/common.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/common.sh"
# shellcheck source=./lib/testing.sh
source "${TOOL_LIB_PATH}/testing.sh"
# shellcheck source=./lib/gitlib.sh
source "${TOOL_LIB_PATH}/gitlib.sh"
# shellcheck source=./lib/releaselib.sh
source "${TOOL_LIB_PATH}/releaselib.sh"

# TODO: We should see to
#       - move that to the start of the script
#       - add `set -o nounset`
#
#       We can do that when all the things we source do not rely on unset
#       varaibales and ignoring errors. This will require quite some
#       refactoring, so this is the best we can do for now.
set -o errexit
set -o pipefail

TEST_verify_latest_update() {
  ##############################################################################
  # TESTING release::gcs::verify_latest_update()
  ##############################################################################
  # stable is always vX.Y.Z
  # latest is always vX.Y.Z-(alpha|beta|rc).N

  # > vs >= scenarios only relevant to 'latest' .N's.
  # type release updates on >
  # type ci updates on >=

  published_file="$( mktemp -t 'published.XXXXXXXX')"
  trap 'rm -f "$published_file"' EXIT

  # We want to make sure that we explicitly not use gsutils in this test, so
  # that we do a `cat $file` instead of a `gsutil cat $file` (`gsutil cat` only
  # supports remote urls, but not local files).
  export GSUTIL=''

  # Fill $data with testing values and expected state
  read -r -d '' data <<'EOF' || true
###############################################
# stable test scenarios
# release and ci use same logic here
#  type    incoming       published      State
#  ----    --------       ---------      -----
## release v2.0.0         v1.0.0         0
## release v1.0.0         v1.0.0         1
## release v0.9.9         v1.0.0         1

## release v1.1.0         v1.0.0         0
## release v1.1.0         v1.1.0         1
## release v1.0.0         v1.1.0         1

## release v1.3.4         v1.3.3         0
## release v1.3.3         v1.3.3         1
## release v1.3.2         v1.3.3         1


###############################################
# latest test scenarios
#  type    incoming       published      State
#  ----    --------       ---------      -----
## release v1.4.0-alpha.0 v1.4.0         1
## release v1.4.0-alpha.0 v1.4.0-beta.0  1
## release v1.4.0-beta.0  v1.4.0-alpha.0 0
## ci      v1.4.0-alpha.1 v1.4.0-alpha.0 0
## ci      v1.4.0-alpha.0 v1.4.0-alpha.0 0
## release v1.4.0-alpha.1 v1.4.0-alpha.0 0
## release v1.4.0-alpha.0 v1.4.0-alpha.0 1
## release v1.4.0-rc.1    v1.4.0-beta.5  0
## release v1.4.0         v1.4.0-rc.1    0
EOF

  # Test the data
  # disable shellcheck for comment variable
  # shellcheck disable=SC2034
  while read -r comment type version pub_version expected; do
    # Prepare test
    echo "$pub_version" > "$published_file"

    # $type value passed in simply to trigger > vs. >= condition
    # arg 2 (bucket) not used with optional arg 4 passed in
    # arg 3 (version) is the incoming version to check
    # arg 4 simply points to a local file to set a 'published' version
    if release::gcs::verify_latest_update "$type" "" "$version" "$published_file"; then
      echo -n "TEST CASE: "
      case $expected in
        0) echo "$PASSED" ;;
        *) echo "$FAILED" ; return 1 ;;
      esac
    else
      echo -n "TEST CASE: "
      case $expected in
        1) echo "$PASSED" ;;
        *) echo "$FAILED" ; retrun 1 ;;
      esac
    fi
    echo
  done < <(echo "$data" | grep -E '^## ')

  ##############################################################################
  # END TESTING release::gcs::verify_latest_update()
  ##############################################################################
}

TEST_kubecross_version() {
  echo 'Testing release::kubecross_version'
  echo

  local testName testCases testCase currentTest expRc rc expVer ver

  # array of comma-separated test case inputs:
  # [0]  expected return code
  # [1]  expected version
  # [2:] inputs for the thing under test, release::kubecross_version
  testCases=(
    # test against one existing branch
    '0,v1.10.8-1,release-1.12'
    # test against one non-existing branch
    '1,,does-not-exist'
    # test against a non-existing branch with one fallback that exists
    '0,v1.10.8-1,does-not-exist,release-1.12'
    # test against a non-existing branch with multiple fallbacks
    '0,v1.10.8-1,does-not-exist,release-1.12,release-1.14'
    # test fallthrough multiple non-existing branches
    '0,v1.10.8-1,does-not-exist,does-not-exist-either,release-1.12'
    # test with no branch specified
    '1,,'
  )

  # mock curl: the mocked curl only returns a valid version for release-1.12
  curl() {
    case "$*" in
      */kubernetes/release-1.12/build/build-image/cross/VERSION)
        echo 'v1.10.8-1'      ; return 0
        ;;
      *)
        echo '404: Not found' ; return 1
        ;;
    esac
  }

  for testCase in "${testCases[@]}"
  do
    IFS=',' read -r -a currentTest <<< "$testCase"

    # pop off stuff from the array
    expRc="${currentTest[0]}" ; currentTest=("${currentTest[@]:1}")
    expVer="${currentTest[0]}" ; currentTest=("${currentTest[@]:1}")

    testName="release::kubecross_version ${currentTest[*]}: "

    rc=0; ver="$( release::kubecross_version "${currentTest[@]}" 2>/dev/null )" || rc=$?

    if [ "$rc" != "$expRc" ]; then
      echo "${FAILED} ${testName}, expected to return ${expRc}, got ${rc}"
      continue
    fi

    if [ "$ver" != "$expVer" ]; then
      echo "${FAILED} ${testName}, expected version to be ${expVer}, got ${ver}"
      continue
    fi

    echo "${PASSED} ${testName}"
  done
}

test_main "$@"
