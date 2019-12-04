#!/usr/bin/env bash

# Copyright 2016 The Kubernetes Authors.
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

###############################################################################
# GIT-related constants and functions

###############################################################################
# CONSTANTS
###############################################################################

GITHUB_TOKEN=${FLAGS_github_token:-$GITHUB_TOKEN}

# This is a safety measure to trim extraneous leading/trailing whitespace from
# the OAuth token provided. Otherwise, authentication to GitHub will fail here.
GITHUB_TOKEN="$( echo "$GITHUB_TOKEN" | tr -d '[:space:]' )"
[[ -n "$GITHUB_TOKEN" ]] && GITHUB_TOKEN_FLAG=("-u" "${GITHUB_TOKEN}:x-oauth-basic")
GHCURL="curl -s --fail --retry 10 ${GITHUB_TOKEN_FLAG[*]}"
JCURL="curl -g -s --fail --retry 10"
GITHUB_API='https://api.github.com'
GITHUB_API_GRAPHQL="${GITHUB_API}/graphql"
K8S_GITHUB_API_ROOT="${GITHUB_API}/repos"
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
# Disable shellcheck for dynamically defined variable
# shellcheck disable=SC2034,SC2154
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
  done < <($GHCURL "${K8S_GITHUB_API}/pulls?state=open&base=${branch}" |\
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
  local expectedRemote="${RELEASE_TOOL_REPO:-[:/]kubernetes/release}"
  local expectedBranch="${RELEASE_TOOL_BRANCH:-master}"

  local branch
  branch=$(gitlib::current_branch "$TOOL_ROOT") || return 1

  if [ "${expectedBranch}" != "$branch" ]
  then
    logecho "$FAILED checked out branch $branch is not the same as $expectedBranch"
    return 1
  fi

  local remotePattern="${expectedRemote}(.git)* \(fetch\)$"

  local remote=$( git -C "$TOOL_ROOT" remote -v | grep -E "$remotePattern" -m 1 | cut -f1 )
  local commit=$(git -C "$TOOL_ROOT" \
                     ls-remote --heads "$remote" refs/heads/master | cut -f1)
  local output=$(git -C "$TOOL_ROOT" branch --contains "$commit" "$branch" 2>/dev/null)

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

##############################################################################
# Publish an issue on github
gitlib::create_issue() {
  local repo="${1?expected repo for gitlib::create_issue}"
  local title="${2?expected title for gitlib::create_issue}"
  local body="${3?expected body for gitlib::create_issue}"
  local milestone="${4:-null}"
  local template data

  # shellcheck disable=SC2016
  template='{
    "title": $title,
    "body": $body,
    "milestone": $milestone,
  }'

  data="$( jq \
    --argjson milestone "$milestone" \
    --arg     body      "$body" \
    --arg     title     "$title" \
    -c -n "$template"
  )"

  ${GHCURL} "${K8S_GITHUB_API_ROOT}/${repo}/issues" --data "$data"
}


###############################################################################
# Extract the issue url from an issue creation response
gitlib::get_issue_url() {
  local url=''
  url="$( jq -r '.html_url // ""' || echo '' )"
  echo "$url"
}

###############################################################################
# Post an issue for the publishing bot, to ask them to update their
# configuration
gitlib::create_publishing_bot_issue() {
  local branch="$1"
  local repo title body bot_commands milestone

  title="Update publishing-bot for ${branch}"
  repo="k8s-release-robot/sig-release"
  bot_commands=( '<!-- no assignments -->' )
  milestone="v${branch#release-}"

  if ((FLAGS_nomock)); then
    local team_slug='publishing-bot-reviewers'
    local gh_user_tags=()
    local team_members

    repo="kubernetes/kubernetes"
    team_members="$( gitlib::get_team_members 'kubernetes' "$team_slug" )"
    for m in $team_members ; do gh_user_tags+=( "@${m}" ); done
    if (( ${#gh_user_tags[@]} > 0 ))
    then
      bot_commands=( "/assign ${gh_user_tags[*]}" )
    fi
    bot_commands+=( "/cc @kubernetes/${team_slug}" )
  fi

  body="$(cat <<EOF
The branch [\`${branch}\`][branch] was just created.

Please update the [publishing-bot's configuration][config] to also publish this
new branch.

[branch]: https://github.com/kubernetes/kubernetes/tree/${branch}
[config]: https://github.com/kubernetes/kubernetes/tree/master/staging/publishing

$( local IFS=$'\n' ; echo "${bot_commands[*]}" )
/sig release
/area release-eng
/milestone ${milestone}
EOF
)"

  gitlib::create_issue "$repo" "$title" "$body"
}

###############################################################################
# Get all the members of an organisation's team
gitlib::get_team_members() {
  local org="${1:-kubernetes}"
  local team="${2:-sig-release}"

  local query query_tmpl query_body resp_transformation

  # shellcheck disable=SC2016
  query='query ($org:String!, $team:String!) {
    organization(login: $org) {
      team(slug: $team) {
        members {
          nodes {
            login
          }
        }
      }
    }
  }'
  # shellcheck disable=SC2016
  query_tmpl='{
    "query" : $query,
    "variables": {
      "org": $org,
      "team": $team
    }
  }'
  query_body="$( jq \
    --arg query "$query" \
    --arg org   "$org" \
    --arg team  "$team" \
    -c -n "$query_tmpl" \
  )"
  resp_transformation='(.data.organization.team.members.nodes // {})[].login'

  {
    $GHCURL "$GITHUB_API_GRAPHQL" -X POST -d "$query_body" \
      | jq -r "$resp_transformation"
  } || echo -n ''
}
