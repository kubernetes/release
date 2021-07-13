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

set -o nounset
set -o errexit
set -o nounset
set -o xtrace


declare -r BUILD_TAG="$(date '+%y%m%d%H%M%S')"
declare -r IMG_NAME="debian-builder:${BUILD_TAG}"
declare -r DEB_RELEASE_BUCKET="gs://k8s-release-dev/debian"

docker build -t "${IMG_NAME}" "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker run -it --rm -v "${PWD}/bin:/src/bin" "${IMG_NAME}" $@

gsutil -m cp -nrc bin "${DEB_RELEASE_BUCKET}/${BUILD_TAG}"
printf "%s" "${BUILD_TAG}" | gsutil cp - "${DEB_RELEASE_BUCKET}/latest"
