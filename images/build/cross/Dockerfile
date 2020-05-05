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

# This file creates a standard build environment for building cross
# platform go binary for the architecture kubernetes cares about.

FROM golang:1.14.2

##------------------------------------------------------------
# global ARGs & ENVs

ARG DEBIAN_FRONTEND=noninteractive

ENV GOARM 7
ENV KUBE_DYNAMIC_CROSSPLATFORMS \
  armhf \
  arm64 \
  s390x \
  ppc64el

ENV KUBE_CROSSPLATFORMS \
  linux/386 \
  linux/arm linux/arm64 \
  linux/ppc64le \
  linux/s390x \
  darwin/amd64 darwin/386 \
  windows/amd64 windows/386

##------------------------------------------------------------

# Pre-compile the standard go library when cross-compiling. This is much easier now when we have go1.5+
RUN for platform in ${KUBE_CROSSPLATFORMS}; do GOOS=${platform%/*} GOARCH=${platform##*/} go install std; done \
    && go clean -cache

# Install packages
RUN apt-get -q update \
    && apt-get install -qqy \
        apt-utils \
        apt-transport-https \
        bsdmainutils \
        ca-certificates \
        curl \
        file \
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
        patch \
        rsync \
        software-properties-common \
        tzdata \
        unzip

# Use dynamic cgo linking for architectures other than amd64 for the server platforms
# To install crossbuild essential for other architectures add the following repository.
RUN echo "deb http://archive.ubuntu.com/ubuntu xenial main universe" > /etc/apt/sources.list.d/cgocrosscompiling.list \
  && apt-key adv --no-tty --keyserver keyserver.ubuntu.com --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32 \
  && apt-get update \
  && apt-get install -y build-essential \
  && for platform in ${KUBE_DYNAMIC_CROSSPLATFORMS}; do apt-get install -y crossbuild-essential-${platform}; done

RUN PROTOBUF_VERSION=3.0.2; ZIPNAME="protoc-${PROTOBUF_VERSION}-linux-x86_64.zip"; \
  mkdir /tmp/protoc && cd /tmp/protoc \
  && wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/${ZIPNAME}" \
  && unzip "${ZIPNAME}" \
  && chmod -R +rX /tmp/protoc \
  && cp -pr bin /usr/local \
  && cp -pr include /usr/local \
  && rm -rf /tmp/protoc \
  && protoc --version

# work around 64MB tmpfs size in Docker 1.6
ENV TMPDIR /tmp.k8s
RUN mkdir $TMPDIR \
  && chmod a+rwx $TMPDIR \
  && chmod o+t $TMPDIR

# Get the code coverage tool and goimports
RUN go get golang.org/x/tools/cmd/cover \
           golang.org/x/tools/cmd/goimports \
    && go clean -cache

# Download and symlink etcd. We need this for our integration tests.
RUN export ETCD_VERSION=v3.2.24; \
  mkdir -p /usr/local/src/etcd \
  && cd /usr/local/src/etcd \
  && curl -fsSL https://github.com/coreos/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz | tar -xz \
  && ln -s ../src/etcd/etcd-${ETCD_VERSION}-linux-amd64/etcd /usr/local/bin/

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

ARG SKOPEO_VERSION=v0.1.41
RUN git clone https://github.com/containers/skopeo $GOPATH/src/github.com/containers/skopeo
RUN cd $GOPATH/src/github.com/containers/skopeo \
    && git checkout ${SKOPEO_VERSION} \
    && make binary-local \
    && cp skopeo /usr/local/bin \
    && rm -rf $GOPATH/src/github.com/containers/skopeo

# Cleanup a bit
RUN apt-get -qqy remove \
      wget \
    && apt-get clean \
    && rm -rf -- \
        /var/lib/apt/lists/* \
        ~/.config/gcloud

ENTRYPOINT []
