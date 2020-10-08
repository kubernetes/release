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
# CONSTANTS
###############################################################################

# TODO(vdf): Need to reference K8s Infra projects here
readonly DEFAULT_PROJECT="kubernetes-release-test"
readonly PROD_PROJECT="kubernetes-release"
readonly TEST_PROJECT="kubernetes-release-test"

# TODO(vdf): Need to reference K8s Infra buckets here
readonly DEFAULT_BUCKET="kubernetes-release-gcb"
readonly PROD_BUCKET="kubernetes-release"
readonly TEST_BUCKET="kubernetes-release-gcb"
readonly CI_BUCKET="kubernetes-release-dev"

readonly GCRIO_PATH_PROD="k8s.gcr.io"
readonly GCRIO_PATH_STAGING="gcr.io/k8s-staging-kubernetes"
readonly GCRIO_PATH_MOCK="${GCRIO_PATH_STAGING}/mock"

readonly KUBE_CROSS_REGISTRY="${GCRIO_PATH_PROD}/build-image"
readonly KUBE_CROSS_IMAGE="${KUBE_CROSS_REGISTRY}/kube-cross"
readonly KUBE_CROSS_CONFIG_LOCATION="build/build-image/cross"

# Set a globally usable variable for the changelog directory since we've been
# piecemeal search/replace-ing this and missing some cases.
readonly CHANGELOG_DIR="CHANGELOG"

###############################################################################
# FUNCTIONS
###############################################################################

###############################################################################
# Check that the GCS bucket exists and is writable.
#
# @param bucket - The gs release bucket name
# @return 1 if bucket does not exist or is not writable.
release::gcs::check_release_bucket() {
  local bucket=$1
  local tempfile=$TMPDIR/$PROG-gcs-write.$$

  if ! $GSUTIL ls "gs://$bucket" >/dev/null 2>&1 ; then
    logecho "Google Cloud Storage bucket does not exist: $bucket. Create the bucket with this command:"
    logecho "$GSUTIL mb -p \"$GCLOUD_PROJECT\" \"gs://$bucket\""
    logecho "If the bucket should be publicly readable, make it so with this command:"
    logecho "$GSUTIL defacl ch -u AllUsers:R \"gs://$bucket\""
    logecho "WARNING: This affects all objects uploaded to the bucket!"
    return 1
  fi

  logecho -n "Checking write access to bucket $bucket: "
  if logrun touch $tempfile && \
     logrun $GSUTIL cp $tempfile gs://$bucket && \
     logrun $GSUTIL rm gs://$bucket/${tempfile##*/} && \
     logrun rm -f $tempfile; then
    logecho $OK
  else
    logecho "$FAILED: You do not have access/write permission on $bucket." \
            "Unable to continue."
    return 1
  fi
}

##############################################################################
# Sets major global variables
# Used only in anago
# RELEASE_GB - space requirements per build
# GCP_USER - The user that drives the entire release
# RELEASE_BUCKET - mock or standard release bucket location
# BUCKET_TYPE - stage or release
# WRITE_RELEASE_BUCKETS - array of writable buckets
# READ_RELEASE_BUCKETS - array of readable buckets for multiple sourcing of
#                        mock staged builds
# GCRIO_PATH - GCR path based on mock or --nomock
release::set_globals () {
  logecho -n "Setting global variables: "

  # Default disk requirements per version - Modified in found_staged_location()
  RELEASE_GB="75"

  if ((FLAGS_stage)); then
    BUCKET_TYPE="stage"
  else
    BUCKET_TYPE="release"
  fi

  if ((FLAGS_nomock)); then
    GCRIO_PATH="${FLAGS_gcrio_path:-$GCRIO_PATH_STAGING}"
  else
    GCRIO_PATH="${FLAGS_gcrio_path:-$GCRIO_PATH_MOCK}"
  fi

  if ((FLAGS_nomock)); then
    RELEASE_BUCKET="$PROD_BUCKET"
  else
    RELEASE_BUCKET="$TEST_BUCKET"

    # This is passed to logrun() where appropriate when we want to mock
    # specific activities like pushes
    LOGRUN_MOCK="-m"
  fi

  WRITE_RELEASE_BUCKETS=("$RELEASE_BUCKET")
  READ_RELEASE_BUCKETS+=("$RELEASE_BUCKET")

  # TODO:
  # These KUBE_ globals extend beyond the scope of the new release refactored
  # tooling so to pass these through as flags will require fixes across
  # kubernetes/kubernetes and kubernetes/release which we can do at a later time
  export KUBE_DOCKER_REGISTRY="$GCRIO_PATH"
  export KUBE_RELEASE_RUN_TESTS=n
  export KUBE_SKIP_CONFIRMATIONS=y

  logecho $OK
}
