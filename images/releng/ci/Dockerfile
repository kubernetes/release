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

ARG GO_VERSION
ARG OS_CODENAME

# Get Go from official Alpine image
FROM golang:${GO_VERSION}-alpine as go-source

FROM debian:${OS_CODENAME}-slim

# Copy Go installation from official image
COPY --from=go-source /usr/local/go /usr/local/go

# Set up Go environment
ENV PATH=/usr/local/go/bin:$PATH
ENV GOROOT=/usr/local/go
ENV GOLANG_VERSION=${GO_VERSION}
ENV GOTOOLCHAIN=auto
ENV GOPATH=/usr/local/go

# Install required packages for adding Google Cloud SDK repository
RUN apt-get update && \
    apt-get install -y curl gnupg

RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | \
    tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

# Use gpg instead of apt-key (apt-key is deprecated and not present)
RUN install -m 0755 -d /usr/share/keyrings && \
    curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | \
    gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

RUN apt-get update && \
    apt-get install -y \
        bsdmainutils \
        build-essential \
        git \
        google-cloud-sdk \
        jq \
        rpmlint \
        wget \
    && rm -rf /var/lib/apt/lists/*

# install goreleaser
ARG GORELEASER_VERSION=v2.12.2
ARG GORELEASER_SHA=b7c923e1ff208a5aa4c800ee0331d9fdc8d3ea434f1126e43088676b34d5717b
RUN  \
    GORELEASER_DOWNLOAD_FILE=goreleaser_Linux_x86_64.tar.gz && \
    GORELEASER_DOWNLOAD_URL=https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/${GORELEASER_DOWNLOAD_FILE} && \
    wget ${GORELEASER_DOWNLOAD_URL} && \
    echo "$GORELEASER_SHA $GORELEASER_DOWNLOAD_FILE" | sha256sum -c - || exit 1 && \
    tar -xzf $GORELEASER_DOWNLOAD_FILE -C /usr/bin/ goreleaser && \
    rm $GORELEASER_DOWNLOAD_FILE && \
    goreleaser -v

# install ko
ARG KO_VERSION=v0.18.0
ARG KO_SHA=ce8c8776b243357e0a822c279b06c34302460221e834765dee5f4e9e2c0b7b38
RUN  \
    KO_DOWNLOAD_FILE=ko_${KO_VERSION#v}_Linux_x86_64.tar.gz && \
    KO_DOWNLOAD_URL=https://github.com/ko-build/ko/releases/download/${KO_VERSION}/${KO_DOWNLOAD_FILE} && \
    wget ${KO_DOWNLOAD_URL} && \
    echo "$KO_SHA $KO_DOWNLOAD_FILE" | sha256sum -c - || exit 1 && \
    tar -xzf $KO_DOWNLOAD_FILE -C /usr/bin/ ko && \
    rm $KO_DOWNLOAD_FILE && \
    ko version

# Some tests require a working git executable, so we configure a pseudo-user
RUN git config --global user.name releng-ci-user && \
    git config --global user.email nobody@k8s.io
