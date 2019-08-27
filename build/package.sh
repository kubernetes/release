#!/usr/bin/env bash

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

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

BUILD_TIME="$(date '+%y%m%d%H%M%S')"

declare -r PACKAGE_TYPE="${PACKAGE_TYPE?"PACKAGE_TYPE must be set"}"

declare -r BUILD_TAG="${BUILD_TAG:-$BUILD_TIME}"

declare -r PUBLISH="${PUBLISH:-no}"
declare -r GCS_BUCKET="${GCS_BUCKET:-"kubernetes-release-dev"}"

declare -r OUTPUT_DIR="${OUTPUT_DIR:-"${PWD}/_output/${PACKAGE_TYPE}"}"

case "${PACKAGE_TYPE}" in
"debs")
  declare -r IMG_NAME="deb-builder:${BUILD_TAG}"
;;
"rpms")
  declare -r IMG_NAME="rpm-builder:${BUILD_TAG}"
;;
*)
  echo >&2 "'${PACKAGE_TYPE}' is an invalid PACKAGE_TYPE, only 'debs' and 'rpms' supported"
  exit 1
;;
esac

docker build -t "${IMG_NAME}" "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/${PACKAGE_TYPE}"
echo "Cleaning output directory..."
[[ ! -d "${OUTPUT_DIR}" ]] \
  || find "${OUTPUT_DIR}" -maxdepth 1 -mindepth 1 -print0 | xargs -0 rm -rf --
mkdir -p "${OUTPUT_DIR}"


case "${PACKAGE_TYPE}" in
"debs")
  docker run -ti --rm -v "${OUTPUT_DIR}:/src/bin" "${IMG_NAME}" "$@"
  echo
  echo "----------------------------------------"
  echo
  echo "debs written to: ${OUTPUT_DIR}"
  ls -alth "${OUTPUT_DIR}"
;;
"rpms")
  docker run -ti --rm -v "${OUTPUT_DIR}:/root/rpmbuild/RPMS/" "${IMG_NAME}" "$@"
  echo
  echo "----------------------------------------"
  echo
  echo "rpms written to: ${OUTPUT_DIR}"
  ls -alth "${OUTPUT_DIR}"
  echo
  echo "yum repodata written to: "
  find "$OUTPUT_DIR" -type d -name repodata -print0 | xargs -0 ls -alth
;;
esac

chown -R "${USER}" "${OUTPUT_DIR}"

if [[ "${PUBLISH}" == "yes" ]]; then
  case "${PACKAGE_TYPE}" in
  "debs")
    declare -r GCS_FULL_PATH="gs://${GCS_BUCKET}/debian"
  ;;
  "rpms")
    declare -r GCS_FULL_PATH="gs://${GCS_BUCKET}/rpms"
  ;;
  esac

  gsutil -m cp -nrc "${OUTPUT_DIR}" "${GCS_FULL_PATH}/${BUILD_TAG}"
  gsutil -m cp <(printf "%s" "${BUILD_TAG}") "${GCS_FULL_PATH}/latest"
fi
