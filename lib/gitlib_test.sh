#!/usr/bin/env bash
#
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

# gitlib.sh unit tests
#
# shellcheck source=./lib/testing.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/testing.sh"

# shellcheck source=./lib/common.sh
source "$(dirname "$(readlink -ne "${BASH_SOURCE[0]}")")/common.sh"
# shellcheck source=./lib/gitlib.sh
source "$TOOL_LIB_PATH/gitlib.sh"
# shellcheck source=./lib/releaselib.sh
source "$TOOL_LIB_PATH/releaselib.sh"

readonly TESTDATA="$( cd "$(dirname "$0")" && pwd )/testdata"

set -e

TEST_ver_regex() {
  echo "Testing VER_REGEX:"
  echo

  # Ex. v1.5.0-alpha.2.435+8c67d08e3a535d
  local DOTZERO="v1.4.0"
  local SHORT="v1.4.5"
  local LONG="v1.4.5-alpha.0"
  local SHA="v1.4.5-alpha.0.435+8c67d08e3a535d"
  local RC="v1.4.5-rc.1"

  printf "%-40s" "$DOTZERO : "
  if [[ $DOTZERO =~ ${VER_REGEX[dotzero]} ]]; then
    echo "$PASSED Value: ${BASH_REMATCH[0]}"
  else
    echo "$FAILED"
  fi

  printf "%-40s" "$SHORT : "
  if [[ $SHORT =~ ${VER_REGEX[release]} ]]; then
    echo "$PASSED Value: ${BASH_REMATCH[0]}"
  else
    echo "$FAILED"
  fi

  printf "%-40s" "$LONG : "
  if [[ $LONG =~ ${VER_REGEX[release]} ]]; then
    echo "$PASSED Value: ${BASH_REMATCH[0]}"
  else
    echo "$FAILED"
  fi

  printf "%-40s" "$SHA : "
  if [[ $SHA =~ ${VER_REGEX[release]}\.${VER_REGEX[build]} ]]; then
    echo "$PASSED Value: ${BASH_REMATCH[0]}"
  else
    echo "$FAILED"
  fi

  printf "%-40s" "$RC : "
  if [[ $RC =~ ${VER_REGEX[release]} ]]; then
    echo "$PASSED Value: ${BASH_REMATCH[0]}"
  else
    echo "$FAILED"
  fi
}

TEST_create_issue() {
  echo "Testing gitlib::create_issue"
  echo

  local repo='some/repo'
  local title='the title'
  local body='and the body with some "strange" stuff
          and even new lines'

  # shellcheck disable=SC2034
  local GHCURL='echo'

  assertEqualContent \
    <( gitlib::create_issue "$repo" "$title" "$body" ) \
    "${TESTDATA}/gitlib/create_issue.txt" \
    'creating an issue'

  assertEqualContent \
    <( gitlib::create_issue "$repo" "$title" "$body" 12345 ) \
    "${TESTDATA}/gitlib/create_issue_milestone.txt" \
    'creating an issue with milestone'
}

TEST_create_publishing_bot_issue() {
  echo "Testing gitlib::create_publishing_bot_issue"
  echo

  # shellcheck disable=SC2034
  local GHCURL='echo'

  assertEqualContent \
    <( gitlib::create_publishing_bot_issue 'release-1.14' ) \
    "${TESTDATA}/gitlib/create_publishing_bot_issue.txt" \
    "simple, mock issue without special settings something"

  # shellcheck disable=SC2034
  local FLAGS_nomock=1
  # mock gitlib::get_team_members
  gitlib::get_team_members() {
    echo 'memberOne'
    echo 'memberTwo'
  }
  assertEqualContent \
    <( gitlib::create_publishing_bot_issue 'release-1.13' ) \
    "${TESTDATA}/gitlib/create_publishing_bot_issue_nomock.txt" \
    "for a nomock release, different repo & assignments are used"

  # shellcheck disable=SC2034
  local FLAGS_nomock=1
  # mock gitlib::get_team_members
  gitlib::get_team_members() { :; }
  assertEqualContent \
    <( gitlib::create_publishing_bot_issue 'release-1.13' ) \
    "${TESTDATA}/gitlib/create_publishing_bot_issue_nomock_noassign.txt" \
    "for a nomock release, different repo is used, but no assignments when team members cannot be found"
}

TEST_get_team_members() {
  echo "Testing gitlab::get_team_members"
  echo

  # shellcheck disable=SC2034
  local GHCURL='mock_github_api'

  mock_github_api() {
    echo '{"data": {"organization":{"team":{"members":{"nodes":[{"login":"blipp"},{"login":"blupp"}]}}}}}'
  }
  assertEqualContent \
    <( gitlib::get_team_members 'ignored' 'ignored' ) \
    <( echo -e "blipp\nblupp" ) \
    'get_team_members issues a graphql qeury against the github API and pareses the response'

  mock_github_api() {
    echo 'this is some invalid response'
  }
  assertEqualContent \
    <( gitlib::get_team_members 'ignored' 'ignored' ) \
    <( echo -n '' ) \
    'get_team_members can handle invalid responses and reports no members'
}

TEST_pending_prs() {
  echo "Testing gitlab::pendig_prs"
  echo

  local branch='release-1.12'
  local mockResponse='[{"number":123,"milestone":{"title":"foobar"},"user":{"login":"batman"},"updated_at":"2019-07-03T08:48:34Z","title":"best PR"}]'
  local expectedCallArgs="-s --fail --retry 10 https://api.github.com/repos/kubernetes/kubernetes/pulls?state=open&base=${branch}"

  curl() {
    [ "$expectedCallArgs" != "$*" ] && {
      echo 'Wrong call args:'
      echo "  Expected: ${expectedCallArgs}"
      echo "  Actual:   ${*}"
      return 42
    }
    echo "$mockResponse"
  }

  assertEqualContent \
    <( gitlib::pending_prs "$branch" ) \
    <( echo "#123      foobar  @batman     2019-07-03 08:48    best PR" ) \
    'calls the github API with the correct settings and parses response correctly'
}

testMain "$@"
