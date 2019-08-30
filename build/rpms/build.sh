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

# TODO: Allow running this script locally. Right now, this can only be used as a docker entrypoint.

BUILD_DIR="/home/builder/workspace"

declare -a ARCHS

if [ $# -gt 0 ]; then
  IFS=','; ARCHS=("$1"); unset IFS;
else
  #GOARCH/RPMARCH
  ARCHS=(
    amd64/x86_64
    arm/armhfp
    arm64/aarch64
    ppc64le/ppc64le
    s390x/s390x
  )
fi

# TODO: Add support for multiple spec files once we break packages out into separate specs.
for ARCH in "${ARCHS[@]}"; do
  IFS=/ read -r GOARCH RPMARCH<<< "${ARCH}"; unset IFS;
  SRC_PATH="${BUILD_DIR}/SOURCES/${RPMARCH}"
  mkdir -p "${SRC_PATH}"
  cp -r "${BUILD_DIR}/SPECS"/* "${SRC_PATH}"
  echo "Building RPM's for ${GOARCH}....."
  sed -i "s/\%global ARCH.*/\%global ARCH ${GOARCH}/" "${SRC_PATH}/kubelet.spec"
  # Download sources if not already available
  cd "${SRC_PATH}" && spectool -gf kubelet.spec
  /usr/bin/rpmbuild --target "${RPMARCH}" --define "_sourcedir ${SRC_PATH}" -bb "${SRC_PATH}/kubelet.spec"
  mkdir -p "${BUILD_DIR}/RPMS/${RPMARCH}"
  createrepo -o "${BUILD_DIR}/RPMS/${RPMARCH}/" "${BUILD_DIR}/RPMS/${RPMARCH}"
done
