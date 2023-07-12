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
FROM golang:${GO_VERSION}-${OS_CODENAME} as builder

RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | \
    tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | \
    apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -

RUN apt-get update && \
    apt-get install -y \
        bsdmainutils \
        build-essential \
        google-cloud-sdk \
        jq && \
    rm -rf /var/lib/apt/lists/*

# install goreleaser
ARG GORELEASER_VERSION=v1.19.2
ARG GORELEASER_SHA=27c7397b816c43098f88cbccc5aeec3df929fb857f28b2cb8e885d09458ada1e
RUN  \
    GORELEASER_DOWNLOAD_FILE=goreleaser_Linux_x86_64.tar.gz && \
    GORELEASER_DOWNLOAD_URL=https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/${GORELEASER_DOWNLOAD_FILE} && \
    wget ${GORELEASER_DOWNLOAD_URL} && \
    echo "$GORELEASER_SHA $GORELEASER_DOWNLOAD_FILE" | sha256sum -c - || exit 1 && \
    tar -xzf $GORELEASER_DOWNLOAD_FILE -C /usr/bin/ goreleaser && \
    rm $GORELEASER_DOWNLOAD_FILE && \
    goreleaser -v

# install ko
ARG KO_VERSION=v0.14.1
ARG KO_SHA=3f8f8e3fb4b78a4dfc0708df2b58f202c595a66c34195786f9a279ea991f4eae
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
