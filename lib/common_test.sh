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
# shellcheck disable=SC1091
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/testing.sh"

# shellcheck source=./lib/common.sh
# shellcheck disable=SC1091
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/common.sh"
readonly TESTDATA="$( cd "$(dirname "$0")" && pwd )/testdata"

# TODO: We should see to
#       - move that to the start of the script
#       - add `set -o nounset`
#
#       We can do that when all the things we source do not rely on unset
#       variables and ignoring errors. This will require quite some
#       refactoring, so this is the best we can do for now.
set -o errexit
set -o pipefail

TEST_run_stateful() {
  test_scaffold

  assert_equal_content \
    <( common::run_stateful --strip-args 'printf %s\n%s arg1 arg2' ) \
    <( echo -e "\n\nprintf\n\n\narg1\narg2" ) \
    "passing command and arguments"
}

TEST_generate_sha() {
  test_scaffold

  local expected_dir
  local testfilename1
  local testfilename2
  local testfile1
  local testfile2

  expected_dir="$TESTDATA/common/shas"

  testfilename1="shafile1.txt"
  testfilename2="shafile2.txt"

  # shellcheck disable=SC2154
  testfile1="${tmp_dir}/${testfilename1}"
  testfile2="${tmp_dir}/${testfilename2}"

  # shellcheck disable=SC2154
  echo -n "This is the first file to test the generation of SHA hashes." > "$testfile1"

  # shellcheck disable=SC2154
  echo -n "This is the second file to test the generation of SHA hashes." > "$testfile2"

  find "$tmp_dir" -type f | sort | while read -r path; do
    sum="$(common::md5 "${path}")" || return 1
    echo "$sum" > "${path}.md5"

    for bits in "1" "256" "512"; do
      sum="$(common::sha "${path}" "${bits}" "full")" || return 1
      echo "$sum" > "${path}.sha${bits}"
      echo "$sum" >> "${tmp_dir}/SHA${bits}SUMS"
    done
  done

  for sha_file in "$testfile1" "$testfile2"; do
    assert_equal_content \
      "${tmp_dir}/$(basename "${sha_file}").md5" \
      "${expected_dir}/$(basename "${sha_file}").md5" \
      "Validated md5 hash for ${sha_file}"

    for bits in "1" "256" "512"; do
      assert_equal_content \
        "${tmp_dir}/$(basename "${sha_file}").sha${bits}" \
        "${expected_dir}/$(basename "${sha_file}").sha${bits}" \
        "Validated sha${bits} hash for ${sha_file}"
    done
  done

  for bits in "256" "512"; do
    assert_equal_content \
      "$tmp_dir/SHA${bits}SUMS" \
      "$expected_dir/SHA${bits}SUMS" \
      "Validated SHA${bits}SUMS"
  done
}

test_main "$@"
