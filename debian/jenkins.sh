#!/bin/bash

set -o nounset
set -o errexit
set -o nounset
set -o xtrace


declare -r BUILD_TAG="$(date '+%y%m%d%H%M%S')"
declare -r IMG_NAME="debian-builder-${DEB_CHANEL}:${BUILD_TAG}"
declare -r DEB_RELEASE_BUCKET="gs://kubernetes-release-dev/debian/${DEB_CHANEL}"

docker build -t "${IMG_NAME}" debian/
docker run --rm -v "${PWD}/bin:/src/bin" "${IMG_NAME}"

gsutil -m cp -nc bin/* "${DEB_RELEASE_BUCKET}/${BUILD_TAG}/"
gsuilt -m cp <(printf "${BUILD_NUMBER}") "${DEB_RELEASE_BUCKET}/latest"
