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

# install mage
ARG MAGE_VERSION=1.12.1
ARG MAGE_SHA=bed46f9f509fd515382a9113c683841f029e86a8f2357380bfe80b75768ff099
RUN  \
    MAGE_DOWNLOAD_FILE=mage_${MAGE_VERSION}_Linux-64bit.tar.gz && \
    MAGE_DOWNLOAD_URL=https://github.com/magefile/mage/releases/download/v${MAGE_VERSION}/${MAGE_DOWNLOAD_FILE} && \
    wget ${MAGE_DOWNLOAD_URL} && \
    echo "$MAGE_SHA $MAGE_DOWNLOAD_FILE" | sha256sum -c - || exit 1 && \
    tar -xzf $MAGE_DOWNLOAD_FILE -C /usr/bin/ mage && \
    rm $MAGE_DOWNLOAD_FILE && \
    mage --help

# Some tests require a working git executable, so we configure a pseudo-user
RUN git config --global user.name releng-ci-user && \
    git config --global user.email nobody@k8s.io
