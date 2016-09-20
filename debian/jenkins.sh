#!/bin/bash

set -o nounset
set -o errexit
set -o nounset
set -o xtrace


declare -r BUILD_TAG="$(date '+%y%m%d%H%M%S')"
declare -r IMG_NAME="debian-builder-${DEB_CHANNEL}:${BUILD_TAG}"
declare -r DEB_RELEASE_BUCKET="gs://kubernetes-release-dev/debian/${DEB_CHANNEL}"

function cleanup() {
  docker rmi "${IMG_NAME}" || true
}

trap cleanup EXIT SIGINT SIGTERM

docker build -t "${IMG_NAME}" debian/
docker run --rm -v "${PWD}/bin:/src/bin" "${IMG_NAME}"

gsutil -m cp -nc bin/* "${DEB_RELEASE_BUCKET}/${BUILD_TAG}/"
echo "${BUILD_NUMBER}" | gsutil -m cp - "${DEB_RELEASE_BUCKET}/latest"
