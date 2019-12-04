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

FROM k8s.gcr.io/kube-cross:v1.13.4-1

ARG DEBIAN_FRONTEND=noninteractive

# We want to get rid of python2, we want only python3
#
# Right now, the image chain looks like this:
#  k8s-cloud-builder <- k8s.gcr.io/kube-cross:v1.13.4-1 <- golang:1.13.4 <- buildpack-deps:buster-scm <- debian:buster
# python2 comes in with buildpack-deps:buster-scm, because that image installs
# mercurial which in turn has a hard dependency on python2
RUN apt-get -q update \
    && apt-get -qqy purge ".*python2.*" \
    && apt-get -qqy install \
        python3-minimal \
        python3-pip \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3 90 \
    && update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 90

# For now we do not re-install mercurial, we think we might not need it after all.
#
# # For that to work, we for now need to install mercurial via pip as opposed to
# # apt/dpkg. As soon as debian has a mercurial package that has no hard
# # dependency on python2, we can switch back to using the distributor's package
# # again.
# RUN pip3 install --no-cache-dir mercurial

# Install packages
RUN apt-get -q update \
    && apt-get install -qqy \
        apt-transport-https \
        build-essential \
        ca-certificates \
        curl \
        gettext-base \
        git \
        gnupg2 \
        jq \
        lsb-release \
        pandoc \
        software-properties-common \
        wget

# Install Pip packages
RUN pip3 install --no-cache-dir yq

# Packages required by the make in k8s
# localtime
RUN apt-get -q update \
    && apt-get install -qqy tzdata

# install net tools
# required by common.sh
RUN apt-get -q update \
    && apt-get install -qqy \
        grep \
        net-tools \
        rsync

# Install gcloud
# common::set_cloud_binaries() looks for it in this path
RUN mkdir /opt/google
RUN apt-get -y update \
    && apt-get -y install \
        gcc \
        wget \
        ca-certificates \
        # These are necessary for add-apt-respository
        software-properties-common \
    # Install Git >2.0.1
    && apt-get -y update \
    && apt-get -y install git \
    # Setup Google Cloud SDK (latest)
    && curl -sSL https://sdk.cloud.google.com > /tmp/install.sh \
    && bash /tmp/install.sh --install-dir=/opt/google --disable-prompts \
    # install crcmod: https://cloud.google.com/storage/docs/gsutil/addlhelp/CRC32CandInstallingcrcmod
    && pip3 install --no-cache-dir -U crcmod \
    # Clean up
    && apt-get -y remove \
        gcc \
        wget \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf ~/.config/gcloud

ENV PATH=/opt/google/google-cloud-sdk/bin:$PATH

# Install docker stuff
#---------------------
# Based on instructions from:
# https://docs.docker.com/install/linux/docker-ce/debian/
RUN apt-get -y update \
    && apt-get -y install \
        apt-transport-https \
        ca-certificates \
        curl \
        make \
        software-properties-common \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - \
    && apt-key fingerprint 0EBFCD88 \
    && add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/debian \
      $(lsb_release -cs) \
      stable" \
    && apt-get -y update \
    && apt-get -y install \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        unzip

# Cleanup a bit
RUN rm -rf /var/lib/apt/lists/*

ENTRYPOINT []
