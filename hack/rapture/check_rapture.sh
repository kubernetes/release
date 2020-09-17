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

if [ $# -eq 0 ]
  then
    echo "No argument supplied for version. Ex: 1.19.1"
    exit 1
fi

curl https://packages.cloud.google.com/apt/dists/kubernetes-xenial/main/binary-amd64/Packages | grep -F $1 2>&1
if [ $? != 0 ]; then
   echo "Unable to find version $1 published in debs"
fi
curl https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64/repodata/primary.xml | grep -F $1 2>&1
if [ $? != 0 ]; then
   echo "Unable to find version $1 published in rpms"
fi