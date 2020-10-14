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

new_version="v1.19.2"
fix_version_file="fix_version.txt"
gcs_bucket="gs://kubernetes-release/release"

markers_to_fix=(
  "stable"
  "stable-1"
)

markers_to_remove=(
  "stable-1.20"
)

echo "Writing ${new_version} to ${fix_version_file}..."
echo "${new_version}" > "${fix_version_file}"

for marker in "${markers_to_fix[@]}"; do
  echo "Updating $marker to $new_version..."
  gsutil cp "${fix_version_file}" "$gcs_bucket/$marker.txt"
done

for marker in "${markers_to_remove[@]}"; do
  echo "Removing $marker marker..."
  gsutil rm "$gcs_bucket/$marker.txt"
done
