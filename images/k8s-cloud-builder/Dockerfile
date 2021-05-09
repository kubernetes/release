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
FROM k8s.gcr.io/build-image/kube-cross:${KUBE_CROSS_VERSION}

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
#  k8s-cloud-builder <- k8s.gcr.io/kube-cross:v1.13.4-1 <- golang:1.13.4 <- buildpack-deps:buster-scm <- debian:buster
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
      yq

# common::set_cloud_binaries() looks for it in this path
ARG GOOGLE_DIR='/opt/google'

# Install gcloud
RUN bash -c \
      'bash <(curl -sSL https://sdk.cloud.google.com) \
        --install-dir="${GOOGLE_DIR}" \
        --disable-prompts \
        >/dev/null'

ENV PATH="${GOOGLE_DIR}/google-cloud-sdk/bin:${PATH}"

# Install docker cli
# https://docs.docker.com/install/linux/docker-ce/debian/
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - \
    && apt-key fingerprint 0EBFCD88 \
    && add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/debian \
      $(lsb_release -cs) \
      stable" \
    && apt-get -y update \
    && apt-get -qqy install \
        docker-ce-cli

# Cleanup a bit
RUN apt-get -qqy remove \
      wget \
    && apt-get clean \
    && rm -rf -- \
        /var/lib/apt/lists/* \
        ~/.config/gcloud

ENTRYPOINT []
