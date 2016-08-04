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
#+     $PROG - Push Continuous Integration Build
#+
#+ SYNOPSIS
#+     $PROG  [--nomock] [--federation] [--noupdatelatest]
#+            [--bucket=<alt GS bucket>] [--bucket-mirror=<mirror GS bucket>]
#+     $PROG  [--helpshort|--usage|-?]
#+     $PROG  [--help|-man]
#+
#+ DESCRIPTION
#+     Replaces kubernetes/build/push-ci-build.sh.
#+     Used for pushing CI builds during Jenkins' runs.
#+     Runs in mock mode by default. Use --nomock to do a real push.
#+
#+     Federation values are just passed through as exported global vars still
#+     due to the fact that we're still leveraging the existing federation 
#+     interface in kubernetes proper.
#+
#+ OPTIONS
#+     [--nomock]                - Enables a real push
#+     [--federation]            - Enable FEDERATION push
#+     [--bucket=]               - Specify an alternate bucket for pushes
#+     [--bucket-mirror=]        - Specify a mirror bucket for pushes
#+     [--noupdatelatest]        - Do not update the latest file
#+     [--help | -man]           - display man page for this script
#+     [--usage | -?]            - display in-line usage
#+
#+ EXAMPLES
#+     $PROG                     - Do a CI push
#+     $PROG --federation        - Do a CI push with federation
#+     $PROG --bucket=kubernetes-release-$USER
#+                               - Do a CI push to kubernetes-release-$USER
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
RELEASE_BUCKET_MIRROR=$FLAGS_bucket_mirror
# Compatibility with incoming global args
[[ $KUBE_GCS_UPDATE_LATEST == "n" ]] && FLAGS_noupdatelatest=1

if [[ $(cluster/kubectl.sh version --client 2>&1) =~ \
      GitVersion:\"(${VER_REGEX[release]}\.${VER_REGEX[build]})\", ]]; then
  LATEST=${BASH_REMATCH[1]}
else
  common::exit 1 "Unable to get latest version from build tree.  Exiting..."
fi

if ((FLAGS_nomock)); then
  logecho
  logecho "$PROG is running a *REAL* push!!"
  logecho
else
  # This is passed to logrun() where appropriate when we want to mock
  # specific activities like pushes
  LOGRUN_MOCK="-m"
  # Point to a $USER playground
  RELEASE_BUCKET+=-$USER
  if [[ -n $RELEASE_BUCKET_MIRROR ]]; then
    logecho "--bucket-mirror disabled for mock runs..."
    unset RELEASE_BUCKET_MIRROR
  fi
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

for bucket in $RELEASE_BUCKET $RELEASE_BUCKET_MIRROR; do
  logecho -n "Check/make release bucket $bucket: "
  logrun -s release::gcs::ensure_release_bucket $bucket || common::exit 1
done

# These operations can hit bumps and are re-entrant so retry up to 3 times
max_attempts=3
##############################################################################
common::stepheader COPY RELEASE ARTIFACTS
##############################################################################
attempt=0
while ((attempt<max_attempts)); do
  release::gcs::copy_release_artifacts ci $LATEST $KUBE_ROOT/_output \
                                       $RELEASE_BUCKET \
                                       $RELEASE_BUCKET_MIRROR && break
  ((attempt++))
done
((attempt>=max_attempts)) && common::exit 1 "Exiting..."

if ! ((FLAGS_noupdatelatest)); then
  ##############################################################################
  common::stepheader UPLOAD to $RELEASE_BUCKET
  ##############################################################################
  attempt=0
  while ((attempt<max_attempts)); do
    release::gcs::publish_version ci $LATEST $KUBE_ROOT/_output \
                                     $RELEASE_BUCKET \
                                     $RELEASE_BUCKET_MIRROR && break
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
  if ((FLAGS_nomock)); then
    logecho -n "Push federation images: "
    # Because this executes outside the scope and control of the release infra
    # we can't provide a true mock, so just print with $LOGRUN_MOCK
    logrun -s $LOGRUN_MOCK ${KUBE_ROOT}/build/push-federation-images.sh
  else
    logecho "FEDERATION Image push skipped for mocked workflow..."
  fi
fi

# END script
common::timestamp end
