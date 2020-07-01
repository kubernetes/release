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
#+            [--private-bucket]
#+     $PROG  [--helpshort|--usage|-?]
#+     $PROG  [--help|-man]
#+
#+ DESCRIPTION
#+     Replaces kubernetes/build/push-*-build.sh.
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
#+     [--nomock]                  - Enables a real push (--ci only)
#+     [--federation]              - Enable FEDERATION push
#+     [--ci]                      - Used when called from Jenkins (for ci
#+                                   runs)
#+     [--extra-publish-file=]     - [DEPRECATED - use --extra-version-markers
#+                                   instead] Used when need to upload
#+                                   additional version file to GCS. The path
#+                                   is relative and is append to a GCS path.
#+                                   (--ci only)
#+     [--extra-version-markers=]  - Used when need to upload additional
#+                                   version markers to GCS. The path is
#+                                   relative and is append to a GCS path.
#+                                   (--ci only)
#+     [--bucket=]                 - Specify an alternate bucket for pushes
#+     [--release-type=]           - Override auto-detected release type
#+                                   (normally devel or ci)
#+     [--release-kind=]           - Kind of release to push to GCS. Supported
#+                                   values are kubernetes(default) or
#+                                   federation.
#+     [--gcs-suffix=]             - Specify a suffix to append to the upload
#+                                   destination on GCS.
#+     [--docker-registry=]        - If set, push docker images to specified
#+                                   registry/project
#+     [--version-suffix=]         - Append suffix to version name if set.
#+     [--noupdatelatest]          - Do not update the latest file
#+     [--private-bucket]          - Do not mark published bits on GCS as
#+                                   publicly readable.
#+     [--allow-dup]               - Do not exit error if the build already
#+                                   exists on the gcs path.
#+     [--help | -man]             - display man page for this script
#+     [--usage | -?]              - display in-line usage
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

if [[ $(uname) == "Darwin" ]]; then
  READLINK_CMD=greadlink
else
  READLINK_CMD=readlink
fi

source $(dirname $($READLINK_CMD -ne $BASH_SOURCE))/lib/common.sh
source $TOOL_LIB_PATH/gitlib.sh
source $TOOL_LIB_PATH/releaselib.sh

##############################################################################
# Initialize logs
##############################################################################
# Initialize and save up to 10 (rotated logs)
MYLOG=$TMPDIR/$PROG.log
common::logfileinit $MYLOG 10

# BEGIN script
common::timestamp begin

###############################################################################
# MAIN
###############################################################################
RELEASE_BUCKET="${FLAGS_bucket:-"$CI_BUCKET"}"

# Compatibility with incoming global args
[[ $KUBE_GCS_UPDATE_LATEST == "n" ]] && FLAGS_noupdatelatest=1

# Default to kubernetes
: ${FLAGS_release_kind:="kubernetes"}

# This will canonicalize the path
KUBE_ROOT=$(pwd -P)

# TODO: this should really just read the version from the release tarball always
USE_BAZEL=false
if release::was_built_with_bazel $KUBE_ROOT $FLAGS_release_kind; then
  USE_BAZEL=true
  # The check for version in bazel-genfiles can be removed once everyone is off
  # of versions before 0.25.0.
  # https://github.com/bazelbuild/bazel/issues/8651
  if [[ -r  "$KUBE_ROOT/bazel-genfiles/version" ]]; then
    LATEST=$(cat $KUBE_ROOT/bazel-genfiles/version)
  else
    LATEST=$(cat $KUBE_ROOT/bazel-bin/version)
  fi
else
  LATEST=$(tar -O -xzf $KUBE_ROOT/_output/release-tars/$FLAGS_release_kind.tar.gz $FLAGS_release_kind/version)
fi

if [[ "$LATEST" =~ (${VER_REGEX[release]}(\.${VER_REGEX[build]})?(-dirty)?) ]]; then
  if ((FLAGS_ci)) && [[ "$LATEST" =~ "dirty" ]]; then
    logecho "Refusing to push dirty build with --ci flag given."
    logecho "CI builds should always be performed from clean commits."
    logecho
    logecho "version output:"
    logecho $LATEST
    common::exit 1
  fi
else
  logecho "Unable to get latest version from build tree!"
  logecho
  logecho "version output:"
  logecho $LATEST
  common::exit 1
fi

if [[ -n "${FLAGS_version_suffix:-}" ]]; then
  LATEST+="-${FLAGS_version_suffix}"
fi

GCS_DEST="devel"
((FLAGS_ci)) && GCS_DEST="ci"
GCS_DEST=${FLAGS_release_type:-$GCS_DEST}
GCS_DEST+="$FLAGS_gcs_suffix"
GCS_EXTRA_VERSION_MARKERS_STRING="${FLAGS_extra_version_markers:${FLAGS_extra_publish_file:-}}"

PREVIOUS_IFS=$IFS
IFS=',' read -ra GCS_EXTRA_VERSION_MARKERS <<< "$GCS_EXTRA_VERSION_MARKERS_STRING"
IFS=$PREVIOUS_IFS

if ((FLAGS_nomock)); then
  logecho
  logecho "$PROG is running a *REAL* push!!"
  logecho
else
  # Point to a $USER playground
  RELEASE_BUCKET+=-$USER
fi

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

# Nothing should work without this.  The entire release workflow depends
# on it whether running from the desktop or GCB
GCP_USER=$($GCLOUD auth list --filter=status:ACTIVE \
                             --format="value(account)" 2>/dev/null)
[[ -n "$GCP_USER" ]] || common::exit 1 "Unable to set a valid GCP credential!"

logecho -n "Check release bucket $RELEASE_BUCKET: "
logrun -s release::gcs::check_release_bucket $RELEASE_BUCKET || common::exit 1

# These operations can hit bumps and are re-entrant so retry up to 3 times
max_attempts=3
##############################################################################
common::stepheader COPY RELEASE ARTIFACTS
##############################################################################
attempt=0
while ((attempt<max_attempts)); do
  if $USE_BAZEL; then
    release::gcs::bazel_push_build $GCS_DEST $LATEST $KUBE_ROOT/_output \
                                   $RELEASE_BUCKET && break
  else
    release::gcs::locally_stage_release_artifacts $GCS_DEST $LATEST \
                                                  $KUBE_ROOT/_output \
                                                  $FLAGS_release_kind
    release::gcs::push_release_artifacts \
     $KUBE_ROOT/_output/gcs-stage/$LATEST \
     gs://$RELEASE_BUCKET/$GCS_DEST/$LATEST && break
  fi
  ((attempt++))
done
((attempt>=max_attempts)) && common::exit 1 "Exiting..."

if [[ -n "${FLAGS_docker_registry:-}" ]]; then
  ##############################################################################
  common::stepheader PUSH DOCKER IMAGES
  ##############################################################################
  # TODO: support Bazel too
  # Docker tags cannot contain '+'
  release::docker::release $FLAGS_docker_registry ${LATEST/+/_} \
    $KUBE_ROOT/_output
fi

# If not --ci, then we're done here.
((FLAGS_ci)) || common::exit 0 "Exiting..."

if ! ((FLAGS_noupdatelatest)); then
  ##############################################################################
  common::stepheader UPLOAD to $RELEASE_BUCKET
  ##############################################################################
  attempt=0
  while ((attempt<max_attempts)); do
    release::gcs::publish_version $GCS_DEST $LATEST $KUBE_ROOT/_output \
                                  $RELEASE_BUCKET $GCS_EXTRA_VERSION_MARKERS && break
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
  logrun -s ${KUBE_ROOT}/federation/develop/push-federation-images.sh
fi

# END script
common::timestamp end
