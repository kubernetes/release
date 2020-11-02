/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package release

// "context"
// "os"
// "path/filepath"
// "strings"

// "cloud.google.com/go/storage"
// "github.com/pkg/errors"
// "github.com/sirupsen/logrus"

// "k8s.io/release/pkg/gcp/gcs"
// "k8s.io/release/pkg/util"
// "k8s.io/utils/pointer"

// PushBuild is the main structure for pushing builds.
type Candidate struct {
}

func fastforward_check() {

	/*
	   	# When creating new alphas on the master branch, make sure that X.Y-1.0
	   # has been created first or warn.
	   # This is to ensure that `krel ff` can continue to be used as needed.  User
	   # can override here.
	   # @param buildver - Incoming JENKINS_BUILD_VERSION
	   fastforward_check () {
	     local buildver=$1
	     local latest_official

	     if [[ ! $buildver =~ ${VER_REGEX[release]} ]]; then
	       logecho "Invalid format: $buildver"
	       return 1
	     fi

	     # For master branch alpha builds, ensure we've released the previous .0
	     # first
	     if [[ -z "$PARENT_BRANCH" && -n ${RELEASE_VERSION[alpha]} ]]; then
	       latest_official=${BASH_REMATCH[1]}.$((${BASH_REMATCH[2]}-1))
	       # The $'\n` construct below is a word boundary.
	       if [[ ! "$($GHCURL $K8S_GITHUB_API/tags |jq -r '.[] .name')" =~ \
	          $'\n'v$latest_official.0$'\n' ]]; then
	         logecho
	         logecho "$WARNING:" \
	                 "$latest_official.0 hasn't been tagged/created yet." \
	                 "Creating ${RELEASE_VERSION[alpha]} *will* preclude any" \
	                 "future branch fast-forwards from master to" \
	                 "release-$latest_official."

	         if ! ((FLAGS_yes)); then
	           logecho "Are you *really* sure you want to do this?"
	           common::askyorn "Continue creating ${RELEASE_VERSION[alpha]} now" \
	            || return 1
	         fi
	       fi
	     fi
	   }
	*/

}

/*
##############################################################################
# Calls into Jenkins looking for a build to use for release
# Sets global PARENT_BRANCH when a new branch is created
# And global BRANCH_POINT when new branch is created from an existing tag


PROGSTEP[get_build_candidate]="SET BUILD CANDIDATE"
get_build_candidate () {
  local testing_branch
  local branch_head=$($GHCURL $K8S_GITHUB_API/commits/$RELEASE_BRANCH |\
                      jq -r '.sha')
  # Shorten
  branch_head=${branch_head:0:14}

  # Are we branching to a new branch?
  if gitlib::branch_exists $RELEASE_BRANCH; then
    logecho "RELEASE_BRANCH==$RELEASE_BRANCH already exists"
    # If the branch is a 3-part branch (ie. release-1.2.3)
    if [[ $RELEASE_BRANCH =~ $BRANCH_REGEX ]] && \
       [[ -n ${BASH_REMATCH[4]} ]]; then
      [[ "$FLAGS_type" == official ]] \
       || common::exit 1 "--official required on 3-part branches!"

      # The 'missing' . here between 2 and 3 is intentional. It's part of the
      # optional regex.
      BRANCH_POINT=v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}${BASH_REMATCH[3]}
    fi
    testing_branch=$RELEASE_BRANCH
  else
    logecho "RELEASE_BRANCH==$RELEASE_BRANCH does not yet exist"
    [[ $RELEASE_BRANCH =~ $BRANCH_REGEX ]]

    # Not a 3-part branch
    if [[ -z "${BASH_REMATCH[4]}" ]]; then
      if [[ "$FLAGS_type" == official ]]; then
        common::exit 1 "Can't do official releases when creating a new branch!"
      fi

      PARENT_BRANCH=master
      testing_branch=$PARENT_BRANCH
    # if 3 part branch name, check parent exists
    elif gitlib::branch_exists ${RELEASE_BRANCH%.*}; then
      PARENT_BRANCH=${RELEASE_BRANCH%.*}
      # The 'missing' . here between 2 and 3 is intentional. It's part of the
      # optional regex.
      BRANCH_POINT=v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}${BASH_REMATCH[3]}
      testing_branch=$PARENT_BRANCH
    else
      common::exit 1 "$FATAL! We should never get here! branch=$RELEASE_BRANCH"
    fi
  fi

  logecho "PARENT_BRANCH set to $PARENT_BRANCH"
  logecho "BRANCH_POINT set to $BRANCH_POINT"
  logecho "testing_branch set to $testing_branch"

  if [[ -z $BRANCH_POINT ]]; then
    JENKINS_BUILD_VERSION="$FLAGS_buildversion"
    logecho "JENKINS_BUILD_VERSION set to $JENKINS_BUILD_VERSION"

    # The RELEASE_BRANCH should always match with the JENKINS_BUILD_VERSION
    if [[ $RELEASE_BRANCH =~ release- ]] && \
       [[ ! $JENKINS_BUILD_VERSION =~ ^v${RELEASE_BRANCH/release-/} ]]; then
      logecho
      logecho "$FATAL!  branch/build mismatch!"
      logecho "buildversion=$JENKINS_BUILD_VERSION branch=$RELEASE_BRANCH"
      common::exit 1
    fi

    # Check state of master branch before continuing
    fastforward_check $JENKINS_BUILD_VERSION
  else
    # The build version should never be behind HEAD on release existing branches
    if [[ "$RELEASE_BRANCH" =~ release-([0-9]{1,})\. ]]; then
      if [[ $JENKINS_BUILD_VERSION =~ ${VER_REGEX[build]} && \
            ${BASH_REMATCH[2]} != $branch_head ]]; then
        logecho
        logecho "$FATAL: The $RELEASE_BRANCH HEAD is ahead of the chosen" \
                "commit. Releases on release branches must be run from HEAD."
        return 1
      fi
    fi
  fi
}

*/
