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

IMGNAME = kube-cross
IMAGE_VERSION ?= v1.16.4-1
CONFIG ?= go1.16
TYPE ?= default

# Build args
GO_VERSION?=1.16.4
BASEIMAGE?=golang:$(GO_VERSION)
PROTOBUF_VERSION?=3.7.0
ETCD_VERSION?=v3.4.13

IMAGE = $(REGISTRY)/$(IMGNAME)
TAG ?= $(shell git describe --tags --always --dirty)

# Ensure support for 'docker buildx' and 'docker manifest' commands
export DOCKER_CLI_EXPERIMENTAL=enabled

# TODO: Support multi-arch kube-cross images for linux/arm linux/s390x
#       Currently some of the components references in the Dockerfile are
#       not supported in specific architectures (example etcd does not support s390x)
PLATFORMS ?= linux/amd64 linux/arm64 linux/ppc64le #linux/arm linux/s390x

# for legacy images only build linux/amd64
ifeq ($(TYPE), legacy)
  PLATFORMS = linux/amd64
endif

ARCHS = $(patsubst linux/%,%,$(PLATFORMS))

check-env:
ifndef REGISTRY
	$(error REGISTRY is undefined, please set to registry you want to push)
endif

# build with buildx
# https://github.com/docker/buildx/issues/59
.PHONY: container
container: check-env init-docker-buildx
	echo "Building $(IMGNAME) for the following platforms: $(PLATFORMS)"
	@for platform in $(PLATFORMS); do \
		echo "Starting build for $${platform} platform"; \
		docker buildx build \
			--load \
			--progress plain \
			--platform $${platform} \
			--tag $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION) \
			--tag $(IMAGE)-$${platform##*/}:$(TAG)-$(CONFIG)-$(TYPE) \
			--tag $(IMAGE)-$${platform##*/}:latest-$(CONFIG)-$(TYPE) \
			--build-arg=BASEIMAGE=$(BASEIMAGE) \
			--build-arg=PROTOBUF_VERSION=$(PROTOBUF_VERSION) \
			--build-arg=ETCD_VERSION=$(ETCD_VERSION) \
			--file $(TYPE)/Dockerfile \
			.; \
	done

.PHONY: push
push: container
	echo "Pushing $(IMGNAME) tags"
	@for platform in $(PLATFORMS); do \
		echo "Pushing tags for $${platform} platform"; \
		docker push $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION); \
		docker push $(IMAGE)-$${platform##*/}:$(TAG)-$(CONFIG)-$(TYPE); \
		docker push $(IMAGE)-$${platform##*/}:latest-$(CONFIG)-$(TYPE); \
	done

.PHONY: manifest
manifest: push
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(shell echo $(ARCHS) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(IMAGE_VERSION)~g")
	@for platform in $(ARCHS); do docker manifest annotate --arch "$${platform}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${platform}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

# enable buildx
.PHONY: init-docker-buildx
init-docker-buildx:
	./../../../hack/init-buildx.sh
