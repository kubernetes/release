#!/bin/bash
#
# Copyright 2016 The Kubernetes Authors All rights reserved.
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
###############################################################################
# GIT-related constants and functions

###############################################################################
# CONSTANTS
###############################################################################
GHCURL="curl -s --fail --retry 10 -u ${GITHUB_TOKEN:-$FLAGS_github_token}:x-oauth-basic"
JCURL="curl -g -s --fail --retry 10"

# Regular expressions for bash regex matching
# 0=entire branch name
# 1=Major
# 2=Minor
# 3=.Patch
# 4=Patch
BRANCH_REGEX="master|release-([0-9]{1,})\.([0-9]{1,})(\.([0-9]{1,}))*$"
# release - 1=Major, 2=Minor, 3=Patch, 4=-(alpha|beta|rc), 5=rev
# dotzero - 1=Major, 2=Minor
# build - 1=build number, 2=sha1
# Define what a "digit" looks like
declare -A VER_REGEX=([digit]="(0|[1-9][0-9]*)")
# v is optional (Istio)
VER_REGEX+=([x.y]="v?${VER_REGEX[digit]}\.${VER_REGEX[digit]}")
VER_REGEX+=([release]="${VER_REGEX[x.y]}\.${VER_REGEX[digit]}(-[a-zA-Z0-9]+)*\.*${VER_REGEX[digit]}?"
            [dotzero]="${VER_REGEX[x.y]}\.0$"
            [build]="([0-9]{1,})\+([0-9a-f]{5,40})"
           )

###############################################################################
# FUNCTIONS
###############################################################################

###############################################################################
# Set API roots and git URLs
# @optparam --use-remote - instruct to try to get a remote from git clone
# Sets GLOBALS for use throughout
gitlib::set_api_roots () {
  local arg=$1
  local remote
  local org
  local repo

  if [[ $arg == "--use-remote" ]]; then
    if remote=$(git config --get remote.origin.url) && \
       [[ $remote =~ .*github.com.([^/]+)/([^/]+) ]]; then
      org="${BASH_REMATCH[1]}"
      repo="${BASH_REMATCH[2]}"
    fi
  fi

  GITHUB_ORG=${org:-"kubernetes"}
  GITHUB_REPO=${repo:-"kubernetes"}

  K8S_GITHUB_API_ROOT="https://api.github.com/repos"
  K8S_GITHUB_API="$K8S_GITHUB_API_ROOT/$GITHUB_ORG/$GITHUB_REPO"
  K8S_GITHUB_RAW_ORG="https://raw.githubusercontent.com/$GITHUB_ORG"

  K8S_GITHUB_SEARCHAPI="https://api.github.com/search/issues?per_page=100&q=is:pr%20state:closed%20repo:$GITHUB_ORG/$GITHUB_REPO%20"
  K8S_GITHUB_URL="https://github.com/$GITHUB_ORG/$GITHUB_REPO"
  K8S_GITHUB_SSH="git@github.com:$GITHUB_ORG/$GITHUB_REPO"
}

###############################################################################
# Attempt to authenticate to github using ssh and if unsuccessful provide
# guidance for setup
#
gitlib::ssh_auth () {

  logecho -n "Checking ssh auth to github.com: "
  if ssh -T ${K8S_GITHUB_SSH%%:*} 2>&1 |fgrep -wq denied; then
    logecho $FAILED
    logecho
    logecho "See https://help.github.com/categories/ssh"
    return 1
  fi

  logecho $OK
}


###############################################################################
# Check if authenticated (GITHUB_TOKEN) user is a repo admin.
# Repo admins always have access to push to any restricted branch.
# See: https://github.com/kubernetes/kubernetes/settings/branches
#
# returns 1 if authenticated user is NOT a repo admin
gitlib::is_repo_admin () {
  local result

  logecho -n "Checking repo admin state: "
  result=$($GHCURL $K8S_GITHUB_API | jq -r '.permissions.admin')

  if [[ $result == "true" ]]; then
    logecho $OK
    return 0
  else
    logecho $FAILED
    logecho
    logecho "You must be a repo admin to continue."
    logecho "1. Ensure you are a member - https://github.com/kubernetes/community/blob/master/community-membership.md#requirements"
    logecho "2. Use the 'Request to Join' button on https://github.com/orgs/kubernetes/teams/kubernetes-release-managers/members"
    return 1
  fi
}

###############################################################################
# Looks up the list of releases on github and puts the last release per branch
# into a global branch-indexed dictionary LAST_RELEASE[$branch]
#
# USEFUL: LAST_TAG=$(git describe --abbrev=0 --tags)
gitlib::last_releases () {
  local release
  local branch_name
  local latest_branch
  declare -Ag LAST_RELEASE

  logecho -n "Setting last releases by branch: "
  # Releases are always defined at the toplevel $GITHUB_ORG (for now)
  for release in $($GHCURL \
                   $K8S_GITHUB_API_ROOT/$GITHUB_ORG/$GITHUB_ORG/releases |\
                   jq -r '.[] | select(.draft==false) | .tag_name'); do
    # Alpha releases only on master branch
    # NOTE: The logic here is mostly Kubernetes specific for now
    # As I understand it, Istio's branching and tagging strategy is in flux
    # so don't try to guess what it is and will be here for now.
    # As of 2017/10/12 last release on master will end up being the x.0 release
    # and on branches will be the last x.y release. That may even be the desired
    # result and of course this is only the 'feature' of guessing tag
    # boundaries.  The user can always put specific ranges on the command-line.
    if [[ $release =~ -alpha ]]; then
      LAST_RELEASE[master]=${LAST_RELEASE[master]:-$release}
    # 1=Major, 2=Minor, 3=Patch
    elif [[ $release =~ ${VER_REGEX[x.y]}\.${VER_REGEX[digit]} ]]; then
      # Latest vx.x.0 release goes on both master and release branch.
      if [[ ${BASH_REMATCH[3]} == "0" ]]; then
        LAST_RELEASE[master]=${LAST_RELEASE[master]:-$release}
      fi
      branch_name=release-${BASH_REMATCH[1]}.${BASH_REMATCH[2]}
      LAST_RELEASE[$branch_name]=${LAST_RELEASE[$branch_name]:-$release}
    fi
  done

  logecho -r "$OK"
}

###############################################################################
# What branch am I on?
# @optparam repo_dir - An alternative (to current working dir) git repo
# prints current branch name
# returns 1 if current working directory is not git repository
gitlib::current_branch () {
  local repo_dir=$1
  local -a git_args

  [[ -n "$repo_dir" ]] && git_args=("-C" "$repo_dir")

  if ! git ${git_args[*]} rev-parse --abbrev-ref HEAD 2>/dev/null; then
    (
    logecho
    logecho "Not a git repository!"
    logecho
    ) >&2
    return 1
  fi
}

###############################################################################
# Show the pending/open PRs on a branch
# @param branch
# returns 1 if current working directory is not git repository
gitlib::pending_prs () {
  local branch=$1
  local pr
  local login
  local date
  local msg
  local sep

  if ((FLAGS_htmlize_md)); then
    echo "PR | Milestone | User | Date | Commit Message"
    echo "-- | --------- | ---- | ---- | --------------"
    sep="|"
  fi

  while read pr milestone login date msg; do
    # "escape" '*' in commit messages so they don't mess up formatting.
    msg=$(echo $msg |sed 's, *\* *, * ,g')
    printf "%-8s $sep %-4s $sep %-10s $sep %-18s $sep %s\n" \
           "#$pr" "$milestone" "@$login" "$(date +"%F %R" -d "$date")" "$msg"
  done < <($GHCURL $K8S_GITHUB_API/pulls\?state\=open\&base\=$branch |\
           jq -r \
            '.[] | "\(.number)\t\(.milestone.title)\t\(.user.login)\t\(.updated_at)\t\(.title)"')
}

###############################################################################
# Validates github token using the standard $GITHUB_TOKEN in your env
# returns 0 if token is valid
# returns 1 if token is invalid
gitlib::github_api_token () {
  logecho -n "Checking for a valid github API token: "
  if ! $GHCURL $K8S_GITHUB_API >/dev/null 2>&1; then
    logecho -r "$FAILED"
    logecho
    logecho "No valid github token found in environment or command-line!"
    logecho
    logecho "If you don't have a token yet, go get one at" \
            "https://github.com/settings/tokens/new"
    logecho "1. Fill in 'Token description': $PROG-token"
    logecho "2. Check the []repo box"
    logecho "3. Click the 'Generate token' button at the bottom of the page"
    logecho "4. Use your new token in one of two ways:"
    logecho "   * Set GITHUB_TOKEN in your environment"
    logecho "   * Specify your --github-token=<token> on the command line"
    common::exit 1
  fi
  logecho -r "$OK"
}

##############################################################################
# Git repo sync
# @param repo - full git url
# @param dest - destination directory
gitlib::sync_repo () {
  local repo=$1
  local dest=$2

  logecho -n "Syncing ${repo##*/} to $dest: "
  if [[ -d $dest ]]; then
    (
    cd $dest
    logrun git checkout master
    logrun -s git pull
    ) || return 1
  else
    logrun -s git clone $repo $dest || return 1
  fi
}

##############################################################################
# Does git branch exist?
# @param branch - branch
gitlib::branch_exists () {
  local branch=$1

  git ls-remote --exit-code $K8S_GITHUB_URL \
   refs/heads/$branch &>/dev/null
}

##############################################################################
# Fetch, rebase and push master.
gitlib::push_master () {
  local dryrun_flag=" --dry-run"
  ((FLAGS_nomock)) && dryrun_flag=""

  logecho -n "Checkout master branch to push objects: "
  logrun -s git checkout master || return 1

  logecho -n "Rebase master branch: "
  logrun git fetch origin || return 1
  logrun -s git rebase origin/master || return 1
  logecho -n "Pushing$dryrun_flag master branch: "
  logrun -s git push$dryrun_flag origin master || return 1
}

##############################################################################
# Ensure TOOL_ROOT running with a synced repo.
# 
gitlib::repo_state () {
  local branch=$(gitlib::current_branch $TOOL_ROOT) || return 1
  local remote=$(git -C $TOOL_ROOT remote -v |\
                 awk '/kubernetes\/release(.git)* \(fetch\)/ {print $1}')
  local commit=$(git -C $TOOL_ROOT \
                     ls-remote --heads $remote refs/heads/master | cut -f1)
  local output=$(git -C $TOOL_ROOT branch --contains $commit $branch 2>&-)

  logecho -n "Checking $TOOL_ROOT state: "
  if [[ -n "$output" ]]; then
    logecho $OK
  else
    logecho "$FAILED"
    logecho
    logecho "$TOOL_ROOT is not up to date."
    logecho "$ git pull # to try again"
    return 1
  fi
}
