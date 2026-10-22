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
        make \
        net-tools \
        pandoc \
        rsync \
        tzdata \
        unzip

# Use a virtual environment for Debian's externally managed Python.
RUN apt-get -qqy install \
        python3-dev \
        python3-venv \
    && python3 -m venv /opt/venv

ENV PATH="/opt/venv/bin:${PATH}" \
    CLOUDSDK_PYTHON=/opt/venv/bin/python3 \
    CLOUDSDK_GSUTIL_PYTHON=/opt/venv/bin/python3 \
    CLOUDSDK_PYTHON_SITEPACKAGES=1

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
# Keep this pin for the Bullseye variants, which use Python 3.9.
# gcloud CLI >= 554.0.0 requires
# Python 3.10+ and fails to install on 3.9 with:
#   TypeError: unsupported operand type(s) for |: 'type' and 'type'
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
# https://docs.docker.com/engine/install/binaries/
# Pin to version 24.0.x for API version 1.43 compatibility with DinD in CI
# See: https://github.com/kubernetes/release/issues/4180
# Use the static CLI archive because Trixie has no Docker 24 apt package.
ARG DOCKER_VERSION=24.0.9
RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
        -o /tmp/docker.tgz \
    && tar -xzf /tmp/docker.tgz -C /usr/local/bin --strip-components=1 docker/docker \
    && rm /tmp/docker.tgz

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
