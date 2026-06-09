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

ARG KUBE_CROSS_VERSION
FROM registry.k8s.io/build-image/kube-cross:${KUBE_CROSS_VERSION}

##------------------------------------------------------------
# global ARGs & ENVs

ARG DEBIAN_FRONTEND=noninteractive

##------------------------------------------------------------

# Install packages
RUN apt-get -q update \
    && apt-get install -qqy \
        apt-transport-https \
        bsdmainutils \
        ca-certificates \
        curl \
        gettext-base \
        git \
        gnupg2 \
        grep \
        jq \
        libassuan-dev \
        libbtrfs-dev \
        libdevmapper-dev \
        libgpgme-dev \
        lsb-release \
        make \
        net-tools \
        pandoc \
        rsync \
        software-properties-common \
        tzdata \
        unzip

# We want to get rid of python2, we want only python3
#
# Right now, the image chain looks like this:
#  k8s-cloud-builder <- registry.k8s.io/kube-cross:v1.13.4-1 <- golang:1.13.4 <- buildpack-deps:buster-scm <- debian:buster
# python2 comes in with buildpack-deps:buster-scm, because that image installs
# mercurial which in turn has a hard dependency on python2
RUN apt-get -qqy purge ".*python2.*" \
    && apt-get -qqy install \
        python3-minimal \
        python3-pip \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3 90 \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 90

# Install Pip packages
RUN pip3 install --no-cache-dir \
      # for gcloud https://cloud.google.com/storage/docs/gsutil/addlhelp/CRC32CandInstallingcrcmod
      crcmod \
      osc \
      yq

# common::set_cloud_binaries() looks for it in this path
ARG GOOGLE_DIR='/opt/google'

# Install gcloud
#
# Pin to the last release that still supports Python 3.9. The kube-cross base
# image is Debian bullseye (Python 3.9), but gcloud CLI >= 554.0.0 requires
# Python 3.10+ and fails to install on 3.9 with:
#   TypeError: unsupported operand type(s) for |: 'type' and 'type'
# TODO: migrate the kube-cross base image to bookworm (Python 3.11) or trixie,
# then drop this pin and install the latest gcloud SDK again.
ARG CLOUD_SDK_VERSION=553.0.0
RUN mkdir -p "${GOOGLE_DIR}" \
    && curl -fsSL "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz" \
      | tar -xz -C "${GOOGLE_DIR}" \
    && "${GOOGLE_DIR}/google-cloud-sdk/install.sh" \
        --disable-installation-options \
        --bash-completion=false \
        --path-update=false \
        --usage-reporting=false \
        --quiet \
        >/dev/null

ENV PATH="${GOOGLE_DIR}/google-cloud-sdk/bin:${PATH}"

# Install docker cli
# https://docs.docker.com/install/linux/docker-ce/debian/
# Pin to version 24.0.x for API version 1.43 compatibility with DinD in CI
# See: https://github.com/kubernetes/release/issues/4180
# Exclude buildx and compose plugins which pull in newer API versions
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - \
    && apt-key fingerprint 0EBFCD88 \
    && add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/debian \
      $(lsb_release -cs) \
      stable" \
    && apt-get -y update \
    && apt-get -qqy install \
        --no-install-recommends \
        docker-ce-cli=5:24.0.* \
    && apt-mark hold docker-ce-cli

# Install docker-buildx plugin manually to avoid apt pulling in newer API versions
# Buildx v0.14.1 supports Docker API 1.41+ (Docker Engine 20.10+)
# See: https://github.com/docker/buildx/releases
ARG BUILDX_VERSION=0.14.1
RUN mkdir -p /usr/lib/docker/cli-plugins \
    && curl -fsSL "https://github.com/docker/buildx/releases/download/v${BUILDX_VERSION}/buildx-v${BUILDX_VERSION}.linux-amd64" \
        -o /usr/lib/docker/cli-plugins/docker-buildx \
    && chmod +x /usr/lib/docker/cli-plugins/docker-buildx

# Cleanup a bit
RUN apt-get -qqy remove \
      wget \
    && apt-get clean \
    && rm -rf -- \
        /var/lib/apt/lists/* \
        ~/.config/gcloud

ENTRYPOINT []
