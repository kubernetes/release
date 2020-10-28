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

# include the common image-building Makefiles
#include $(CURDIR)/../../Makefile.common-image

# set default shell
SHELL=/bin/bash -o pipefail

REGISTRY ?= gcr.io/k8s-staging-releng
IMGNAME = kubepkg
IMAGE_VERSION ?= v0.2.0

IMAGE = $(REGISTRY)/$(IMGNAME)

TAG ?= $(shell git describe --tags --always --dirty)

# Build args
GO_VERSION ?= 1.15.3

PLATFORMS ?= linux/amd64

HOST_GOOS ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)
GO_BUILD ?= go build

BUILD_ARGS = --build-arg=GO_VERSION=$(GO_VERSION)

# Ensure support for 'docker buildx' and 'docker manifest' commands
export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: all build clean

.PHONY: all
all: build

.PHONY: build
build:
	$(GO_BUILD)

.PHONY: clean
clean:
	rm kubepkg

# build with buildx
# https://github.com/docker/buildx/issues/59
.PHONY: container
container: init-docker-buildx
	echo "Building $(IMGNAME) for the following platforms: $(PLATFORMS)"
	@for platform in $(PLATFORMS); do \
		echo "Starting build for $${platform} platform"; \
		docker buildx build \
			--load \
			--progress plain \
			--platform $${platform} \
			--tag $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION) \
			--tag $(IMAGE)-$${platform##*/}:$(TAG) \
			--tag $(IMAGE)-$${platform##*/}:latest \
			$(BUILD_ARGS) \
			-f $(CURDIR)/Dockerfile \
			../../.; \
	done

.PHONY: container-rpm
container-rpm: init-docker-buildx
	IMGNAME=kubepkg-rpm
	IMAGE=$(REGISTRY)/$(IMGNAME)
	echo "Building $(IMGNAME) for the following platforms: $(PLATFORMS)"
	@for platform in $(PLATFORMS); do \
		echo "Starting build for $${platform} platform"; \
		docker buildx build \
			--load \
			--progress plain \
			--platform $${platform} \
			--tag $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION) \
			--tag $(IMAGE)-$${platform##*/}:$(TAG) \
			--tag $(IMAGE)-$${platform##*/}:latest \
			$(BUILD_ARGS) \
			-f $(CURDIR)/Dockerfile-rpm \
			../../.; \
	done

.PHONY: push
push: container
	echo "Pushing $(IMGNAME) tags"
	@for platform in $(PLATFORMS); do \
		echo "Pushing tags for $${platform} platform"; \
		docker push $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION); \
		docker push $(IMAGE)-$${platform##*/}:$(TAG); \
		docker push $(IMAGE)-$${platform##*/}:latest; \
	done

.PHONY: push-rpm
push-rpm: container-rpm
	IMGNAME=kubepkg-rpm
	IMAGE=$(REGISTRY)/$(IMGNAME)
	echo "Pushing $(IMGNAME) tags"
	@for platform in $(PLATFORMS); do \
		echo "Pushing tags for $${platform} platform"; \
		docker push $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION); \
		docker push $(IMAGE)-$${platform##*/}:$(TAG); \
		docker push $(IMAGE)-$${platform##*/}:latest; \
	done

.PHONY: manifest
manifest: push
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(IMAGE)-$(subst linux/,,$(firstword $(PLATFORMS))):$(IMAGE_VERSION)
	@for platform in $(PLATFORMS); do docker manifest annotate --arch "$${platform##*/}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${platform##*/}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

.PHONY: manifest-rpm
manifest-rpm: push-rpm
	IMGNAME=kubepkg-rpm
	IMAGE=$(REGISTRY)/$(IMGNAME)
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(IMAGE)-$(subst linux/,,$(firstword $(PLATFORMS))):$(IMAGE_VERSION)
	@for platform in $(PLATFORMS); do docker manifest annotate --arch "$${platform##*/}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${platform##*/}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

# enable buildx
.PHONY: init-docker-buildx
init-docker-buildx:
	./../../hack/init-buildx.sh
