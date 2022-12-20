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

# set default shell
SHELL=/bin/bash -o pipefail

REGISTRY ?= gcr.io/k8s-staging-releng
IMGNAME = k8s-ci-builder
CONFIG ?= default

IMAGE = $(REGISTRY)/$(IMGNAME)

TAG ?= $(shell git describe --tags --always --dirty)

# Build args
GO_VERSION ?= 1.19.4
OS_CODENAME ?= bullseye
IMAGE_ARG ?= $(IMAGE):$(TAG)-$(CONFIG)

BUILD_ARGS = --build-arg=GO_VERSION=$(GO_VERSION) \
             --build-arg=OS_CODENAME=$(OS_CODENAME) \
			 --build-arg=IMAGE_ARG=$(IMAGE_ARG)

# Ensure support for 'docker buildx' and 'docker manifest' commands
export DOCKER_CLI_EXPERIMENTAL=enabled

# build with buildx
# https://github.com/docker/buildx/issues/59
.PHONY: build
build: init-docker-buildx
	echo "Building $(IMGNAME)..."
	docker buildx build \
		--load \
		--progress plain \
		--platform linux/amd64 \
		--tag $(IMAGE):$(CONFIG) \
		--tag $(IMAGE):$(TAG)-$(CONFIG) \
		--tag $(IMAGE):latest-$(CONFIG) \
		$(BUILD_ARGS) \
		-f $(CURDIR)/Dockerfile \
		../../../.

.PHONY: push
push: build
	echo "Pushing $(IMGNAME) tags"
	docker push $(IMAGE):$(CONFIG)
	docker push $(IMAGE):$(TAG)-$(CONFIG)
	docker push $(IMAGE):latest-$(CONFIG)

# enable buildx
.PHONY: init-docker-buildx
init-docker-buildx:
	./../../../hack/init-buildx.sh
