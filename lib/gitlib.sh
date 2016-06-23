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
K8S_GITHUB_API='https://api.github.com/repos/kubernetes/kubernetes'
K8S_GITHUB_URL='https://github.com/kubernetes/kubernetes'
JENKINS_URL="http://kubekins.dls.corp.google.com/job"

# Regular expressions for bash regex matching
# 0=entire branch name
# 1=Major
# 2=Minor
# 3=.Patch
# 4=Patch
BRANCH_REGEX="master|release-([0-9]{1,})\.([0-9]{1,})(\.([0-9]{1,}))*$"
# release - 1=Major, 2=Minor, 3=Patch, 4=-(alpha|beta), 5=rev
# dotzero - 1=Major, 2=Minor
# build - 1=build number, 2=sha1
declare -A VER_REGEX=([release]="v([0-9]{1,})\.([0-9]{1,})\.([0-9]{1,})(-alpha|-beta)*\.*([0-9]{1,})*"
                      [dotzero]="v([0-9]{1,})\.([0-9]{1,})\.0$"
                      [build]="([0-9]{1,})\+([0-9a-f]{5,40})"
                     )

###############################################################################
# FUNCTIONS
###############################################################################

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
  for release in $($GHCURL $K8S_GITHUB_API/releases|\
                   jq -r '.[] | select(.draft==false) | .tag_name'); do
    if [[ $release =~ v([0-9]+\.[0-9]+)\.[0-9]+ ]]; then
      branch_name=release-${BASH_REMATCH[1]}
      # Keep the latest(first) branch
      : ${latest_branch:=$branch_name}
      # Does branch exist?  If not, default tag to master branch
      git rev-parse --verify origin/$branch_name &>/dev/null ||\
       branch_name=master
  
      LAST_RELEASE[$branch_name]=${LAST_RELEASE[$branch_name]:-$release}
    fi
  done

  # If ${LAST_RELEASE[master]} is unset, set it to the last release-* branch
  : ${LAST_RELEASE[master]:=${LAST_RELEASE[$latest_branch]}}

  logecho -r "$OK"
}

###############################################################################
# What branch am I on?
# prints current branch name
# returns 1 if current working directory is not git repository
gitlib::current_branch () {
  if ! git rev-parse --abbrev-ref HEAD 2>/dev/null; then
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
    printf "%-8s $sep %-4s $sep %-10s $sep %-18s $sep %s\n" \
           "#$pr" "$milestone" "@$login" "$(date +"%F %R" -d "$date")" "$msg"
  done < <($GHCURL $K8S_GITHUB_API/pulls\?state\=open\&base\=$branch |\
           jq -r \
            '.[] | "\(.number)\t\(.milestone.title)\t\(.user.login)\t\(.updated_at)\t\(.title)"')
}

###############################################################################
# Validates github credentials using the standard $GITHUB_TOKEN in your env
# returns 0 if credentials are valid
# returns 1 if credentials are invalid
gitlib::check_credentials () {
  logecho -n "Checking for valid github credentials: "
  if ! $GHCURL $K8S_GITHUB_API >/dev/null 2>&1; then
    logecho -r "$FAILED"
    logecho
    logecho "You must set a github token one of two ways:"
    logecho "* Set GITHUB_TOKEN in your environment"
    logecho "* Specify your --github-token=<token> on the command line"
    logecho
    logecho "If you don't have a token yet, go get one at" \
               "https://github.com/settings/tokens"
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
