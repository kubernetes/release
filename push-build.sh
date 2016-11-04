#!/bin/bash
#
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
#
# Set PROGram name
PROG=${0##*/}
########################################################################
#+
#+ NAME
#+     $PROG - Push Kubernetes Release Artifacts up to GCS
#+
#+ SYNOPSIS
#+     $PROG  [--nomock] [--federation] [--noupdatelatest] [--ci]
#+            [--bucket=<GS bucket>]
#+     $PROG  [--helpshort|--usage|-?]
#+     $PROG  [--help|-man]
#+
#+ DESCRIPTION
#+     Replaces kubernetes/build-tools/push-*-build.sh.
#+     Used for pushing developer builds and Jenkins' continuous builds.
#+
#+     Developer pushes simply run as they do pushing to devel/ on GCS.
#+     In --ci mode, $PROG runs in mock mode by default.  Use --nomock to do
#+     a real push.
#+
#+     Federation values are just passed through as exported global vars still
#+     due to the fact that we're still leveraging the existing federation 
#+     interface in kubernetes proper.
#+
#+ OPTIONS
#+     [--nomock]                - Enables a real push (--ci only)
#+     [--federation]            - Enable FEDERATION push
#+     [--ci]                    - Used when called from Jekins (for ci runs)
#+     [--bucket=]               - Specify an alternate bucket for pushes
#+     [--gcs-suffix=]           - Specify a suffix to append to the upload
#+                                 destination on GCS.
#+     [--noupdatelatest]        - Do not update the latest file
#+     [--help | -man]           - display man page for this script
#+     [--usage | -?]            - display in-line usage
#+
#+ EXAMPLES
#+     $PROG                     - Do a developer push
#+     $PROG --nomock --federation --ci
#+                               - Do a (non-mocked) CI push with federation
#+     $PROG --bucket=kubernetes-release-$USER
#+                               - Do a developer push to
#+                                 kubernetes-release-$USER
#+
#+ FILES
#+
#+ SEE ALSO
#+     common.sh                 - base function entry points
#+     releaselib.sh             - ::release:: namespace entry points
#+     kubernetes/hack/jenkins/build.sh
#+                               - caller
#+
#+ BUGS/TODO
#+     * Should federation be pulled into release repo?  It would offer more
#+       control.
#+
########################################################################
# If NO ARGUMENTS should return *usage*, uncomment the following line:
#usage=${1:-yes}

source $(dirname $(readlink -ne $BASH_SOURCE))/lib/common.sh
source $TOOL_LIB_PATH/gitlib.sh
source $TOOL_LIB_PATH/releaselib.sh

##############################################################################
# Initialize logs
##############################################################################
# Initialize and save up to 10 (rotated logs)
MYLOG=/tmp/$PROG.log
common::logfileinit $MYLOG 10

# BEGIN script
common::timestamp begin

###############################################################################
# MAIN
###############################################################################
RELEASE_BUCKET=${FLAGS_bucket:-"kubernetes-release-dev"}
# Compatibility with incoming global args
[[ $KUBE_GCS_UPDATE_LATEST == "n" ]] && FLAGS_noupdatelatest=1

KUBECTL_OUTPUT=$(cluster/kubectl.sh version --client 2>&1 || true)
if [[ "$KUBECTL_OUTPUT" =~ GitVersion:\"(${VER_REGEX[release]}\.${VER_REGEX[build]})\", ]]; then
  LATEST=${BASH_REMATCH[1]}
else
  logecho "Unable to get latest version from build tree!"
  logecho
  logecho "kubectl version output:"
  logecho $KUBECTL_OUTPUT
  common::exit 1
fi

GCS_DEST="devel"
((FLAGS_ci)) && GCS_DEST="ci"
GCS_DEST+="$FLAGS_gcs_suffix"

if ((FLAGS_nomock)); then
  logecho
  logecho "$PROG is running a *REAL* push!!"
  logecho
else
  # Point to a $USER playground
  RELEASE_BUCKET+=-$USER
fi

# This will canonicalize the path
KUBE_ROOT=$(pwd -P)

##############################################################################
common::stepheader CHECK PREREQUISITES
##############################################################################
if ! common::set_cloud_binaries; then
  logecho "Releasing Kubernetes requires gsutil and gcloud. Please download,"
  logecho "install and authorize through the Google Cloud SDK:"
  logecho
  logecho "https://developers.google.com/cloud/sdk/"
  common::exit 1
fi

logecho -n "Check/make release bucket $RELEASE_BUCKET: "
logrun -s release::gcs::ensure_release_bucket $RELEASE_BUCKET || common::exit 1

# These operations can hit bumps and are re-entrant so retry up to 3 times
max_attempts=3
##############################################################################
common::stepheader COPY RELEASE ARTIFACTS
##############################################################################
attempt=0
while ((attempt<max_attempts)); do
  release::gcs::copy_release_artifacts $GCS_DEST $LATEST $KUBE_ROOT/_output \
                                       $RELEASE_BUCKET && break
  ((attempt++))
done
((attempt>=max_attempts)) && common::exit 1 "Exiting..."

# If not --ci, then we're done here.
((FLAGS_ci)) || common::exit 0 "Exiting..."

if ! ((FLAGS_noupdatelatest)); then
  ##############################################################################
  common::stepheader UPLOAD to $RELEASE_BUCKET
  ##############################################################################
  attempt=0
  while ((attempt<max_attempts)); do
    release::gcs::publish_version $GCS_DEST $LATEST $KUBE_ROOT/_output \
                                  $RELEASE_BUCKET && break
    ((attempt++))
  done
  ((attempt>=max_attempts)) && common::exit 1 "Exiting..."
fi

# Leave push-federation-images.sh for now.  Not sure if this makes sense to
# pull into the release repo.
if ((FLAGS_federation)); then
  ############################################################################
  common::stepheader PUSH FEDERATION
  ############################################################################
  logecho -n "Push federation images: "
  # FEDERATION_PUSH_REPO_BASE should be set by the calling job (yaml)
  # TODO: remove once we don't support k8s versions with build/ anymore
  build_dir=${KUBE_ROOT}/build-tools
  [[ -d $build_dir ]] || build_dir=${KUBE_ROOT}/build
  logrun -s $build_dir/push-federation-images.sh
fi

# END script
common::timestamp end
