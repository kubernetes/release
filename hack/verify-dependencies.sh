#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

VERSION=0.0.17
OS="$(uname -s)"
OS="${OS,,}"
URL_BASE=https://github.com/kubernetes-sigs/zeitgeist/releases/download
FILENAME="zeitgeist_${VERSION}_${OS}_amd64.tar.gz"
URL="${URL_BASE}/v${VERSION}/${FILENAME}"

mkdir -p ./bin ./zeitgeist
PATH=$PATH:bin

if ! command -v zeitgeist; then
  curl -sfL "${URL}" -o "${FILENAME}"
  tar -xzf "${FILENAME}" -C zeitgeist
  mv ./zeitgeist/zeitgeist ./bin
  rm -rf ./zeitgeist "${FILENAME}"
fi

zeitgeist validate
