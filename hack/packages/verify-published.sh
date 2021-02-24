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

# TODO: Add logic to support checking all the existence of all packages, not just kubeadm
#       The full list is: kubeadm, kubelet, kubectl, cri-tools, kubernetes-cni

PACKAGE_TYPE="$1"

num_supported_versions=3
# supported_versions will be dynamically populated
supported_versions=""
latest_minor_version=$(curl -sSL dl.k8s.io/release/stable.txt | cut -f2 -d'.')
unstable_versions="alpha|beta|rc"
skipped=""
missing=""
available=""

for (( i = 0; i < num_supported_versions; i++ )); do
	supported_versions="1.$((latest_minor_version-i)) ${supported_versions}"
done

for release in $(curl -s https://api.github.com/repos/kubernetes/kubernetes/releases | jq -r '.[].name'); do
	minor=$(echo "${release}" | cut -f1,2 -d'.')
	if [[ "${release}" =~ $unstable_versions ]]; then
		# alpha, beta, rc releases should be ignored
		echo "Unstable version ${release} ignored"
	elif [[ "${release}" == "v1.20.3" ]]; then
		# v1.20.3 was interrupted due to a conformance metada
		# problem. We ignore this release as no packages were
		# published
		#
		# ref: https://groups.google.com/g/kubernetes-dev/c/oUpY9vWgzJo
		echo "Ignoring v1.20.3: no packages were created"
	elif [[ $supported_versions != *"${minor#v}"* ]]; then
		# release we don't care about (e.g. older releases)
		skipped="${skipped} ${release}"
	else
		case "${PACKAGE_TYPE}" in
		"debs")
			# Install dependencies
			apt-get update && apt-get install -y apt-transport-https curl jq gnupg2

			# Set up Kubernetes packages via apt
			curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
			cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
			apt-get update

			deb_policy=$(apt-cache policy kubeadm)

			if [[ "${deb_policy}" != *"${release#v}"* ]]; then
				# release we care about but has missing debs
				missing="${missing} ${release}"
			else
				# All good, the expected deb package is available
				available="${available} ${release}"
			fi
		;;
		"rpms")
			# Set up Kubernetes packages via yum
			mkdir -p /etc/yum.repos.d
			cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF
			yum update -y

			rpm_policy=$(yum --showduplicates list kubeadm)

			if [[ "${rpm_policy}" != *"${release#v}"* ]]; then
				# release we care about but has missing rpms
				missing="${missing} ${release}"
			else
				# All good, the expected deb package is available
				available="${available} ${release}"
			fi
		;;
		esac
	fi
done

if [[ -n "${skipped}" ]]; then
	echo "Skipped these versions because they aren't supported:"
	echo "${skipped}"
fi

if [[ -n "${available}" ]]; then
	echo "These expected packages were found:"
	echo "${available}"
fi

if [[ -n "${missing}" ]]; then
	echo "ERROR: These versions do not have matching packages:"
	echo "${missing}"
	exit 1
else
	echo ""
	echo "TESTS PASSED!! All necessary packages are pushed!"
	echo ""
fi
