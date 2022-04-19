#!/bin/sh

# Copyright 2022 The Kubernetes Authors.
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

# USAGE: stage-binary-and-deps.sh haproxy /opt/stage
#
# Stages $1 and it's dependencies + their copyright files to $2
#
# This is intended to be used in a multi-stage docker build with a distroless/base
# or distroless/cc image.

REMOVE="/usr/share/base-files
/usr/share/man
/usr/lib/*-linux-gnu/gconv/
/usr/lib/*-linux-gnu/libcrypto.*
/usr/lib/*-linux-gnu/libssl.*
/usr/bin/c_rehash
/usr/bin/openssl
/iptables-wrapper-installer.sh
/clean-distroless.sh"

IFS="
"

for item in ${REMOVE}; do
    rm -rf "${item}"
done
