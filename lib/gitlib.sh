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
GITHUB_TOKEN=${FLAGS_github_token:-$GITHUB_TOKEN}
[[ -n $GITHUB_TOKEN ]] && GITHUB_TOKEN_FLAG=("-u" "$GITHUB_TOKEN:x-oauth-basic")
GHCURL="curl -s --fail --retry 10 ${GITHUB_TOKEN_FLAG[*]}"
JCURL="curl -g -s --fail --retry 10"
K8S_GITHUB_API_ROOT='https://api.github.com/repos'
K8S_GITHUB_API="$K8S_GITHUB_API_ROOT/kubernetes/kubernetes"
K8S_GITHUB_RAW_ORG='https://raw.githubusercontent.com/kubernetes'

K8S_GITHUB_SEARCHAPI_ROOT='https://api.github.com/search/issues?per_page=100'
K8S_GITHUB_SEARCHAPI="$K8S_GITHUB_SEARCHAPI_ROOT&q=is:pr%20repo:kubernetes/kubernetes%20"
K8S_GITHUB_URL='https://github.com/kubernetes/kubernetes'
if ((FLAGS_gcb)); then
  K8S_GITHUB_AUTH_ROOT="https://git@github.com/"
else
  # ssh
  K8S_GITHUB_AUTH_ROOT="git@github.com:"
fi
K8S_GITHUB_AUTH_URL="${K8S_GITHUB_AUTH_ROOT}kubernetes/kubernetes.git"

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
declare -A VER_REGEX=([release]="v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[a-zA-Z0-9]+)*\.*(0|[1-9][0-9]*)?"
                      [dotzero]="v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.0$"
                      [build]="([0-9]{1,})\+([0-9a-f]{5,40})"
                     )

###############################################################################
# FUNCTIONS
###############################################################################

###############################################################################
# Attempt to authenticate to github using ssh and if unsuccessful provide
# guidance for setup
#
gitlib::ssh_auth () {
  logecho -n "Checking ssh auth to github.com: "
  if ssh -T ${K8S_GITHUB_AUTH_URL%%:*} 2>&1 |fgrep -wq denied; then
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
# Validates github token using the standard $GITHUB_TOKEN in your env
# Ensures you have 'private' access to the repo
# returns 0 if token is valid
# returns 1 if token is invalid
gitlib::github_api_token () {
  logecho -n "Checking for a valid github API token: "
  if [[ $($GHCURL $K8S_GITHUB_API -I) =~ Cache-Control:\ private ]]; then
    logecho -r "$OK"
  else
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
}

##############################################################################
# Checks github ACLs
# returns 1 on failure
PROGSTEP[gitlib::github_acls]="CHECK GITHUB AUTH"
gitlib::github_acls () {

  gitlib::github_api_token || return 1
  ((FLAGS_gcb)) || gitlib::ssh_auth || return 1
  gitlib::is_repo_admin || return 1
}

###############################################################################
# Sets up basic git config elements for running within GCB
#
# returns 1 on failure
gitlib::git_config_for_gcb () {
  logrun git config --global user.email "nobody@k8s.io" || return 1
  logrun git config --global user.name "Anago GCB" || return 1
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
  for release in $($GHCURL $K8S_GITHUB_API/releases|\
                   jq -r '.[] | select(.draft==false) | .tag_name'); do
    # Alpha releases only on master branch
    if [[ $release =~ -alpha ]]; then
      LAST_RELEASE[master]=${LAST_RELEASE[master]:-$release}
    elif [[ $release =~ v([0-9]+\.[0-9]+)\.([0-9]+(-.+)?) ]]; then
      # Latest vx.x.0 release goes on both master and release branch.
      if [[ ${BASH_REMATCH[2]} == "0" ]]; then
        LAST_RELEASE[master]=${LAST_RELEASE[master]:-$release}
      fi
      branch_name=release-${BASH_REMATCH[1]}
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

    # for https, update the remotes so we don't have to call the git command-line
    # every time with a token
    (
    cd $dest
    git remote set-url origin $(git remote get-url origin |\
     sed "s,https://git@github.com,https://git:${GITHUB_TOKEN:-$FLAGS_github_token}@github.com,g")
    )
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
    logecho "$ git pull"
    return 1
  fi
}

# Set up git config for GCB
if ((FLAGS_gcb)); then
  gitlib::git_config_for_gcb || common::exit "Exiting..."
fi

###############################################################################
# Search for a matching release tracking issue
# @param version - RELEASE_VERSION_PRIME
# @param repo - org/repo
# returns 1 if none found
# prints most recent open issue matching
gitlib::search_release_issue () {
  local version=$1
  local repo=$2
  local issue

  issue=$($GHCURL "$K8S_GITHUB_SEARCHAPI_ROOT&q=Release+$version+Tracking+in:title+type:issue+state:open+repo:$repo" | jq -r '.items[] | (.number | tostring)' |sort -n |tail -1)

  [[ -z $issue ]] && return 1

  echo $issue
}
