# Copyright 2017 The Kubernetes Authors.
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

IMGNAME = debian-base
BUILD_IMAGE ?= debian-build

IMAGE_VERSION ?= trixie-v1.0.1
CONFIG ?= trixie

# Build args
BASEIMAGE ?= debian:$(CONFIG)-slim

IMAGE = $(REGISTRY)/$(IMGNAME)
TAG ?= $(shell git describe --tags --always --dirty)

TAR_FILE ?= rootfs.tar

PLATFORMS ?= linux/amd64 linux/arm64 linux/ppc64le linux/s390x

ARCHS = $(patsubst linux/%,%,$(PLATFORMS))

check-env:
ifndef REGISTRY
	$(error REGISTRY is undefined, please set to registry you want to push)
endif

.PHONY: container
container: check-env clean
	echo "Building $(IMGNAME) for the following platforms: $(PLATFORMS)"
	for platform in $(PLATFORMS); do \
		arch=$${platform##*/}; \
		echo "Starting build for $${platform} platform"; \
		docker build \
			--pull \
			--load \
			--progress plain \
			--platform $${platform} \
			--tag $(BUILD_IMAGE)-$${arch} \
			--build-arg=BASEIMAGE=$(BASEIMAGE) \
			--file $(CONFIG)/Dockerfile.build \
			$(CONFIG) || exit 1; \
		docker rm -f $(BUILD_IMAGE)-$${arch} >/dev/null 2>&1 || true; \
		docker create --platform $${platform} --name $(BUILD_IMAGE)-$${arch} $(BUILD_IMAGE)-$${arch} || exit 1; \
		docker export $(BUILD_IMAGE)-$${arch} > $(CONFIG)/$(TAR_FILE) || exit 1; \
		docker rm -f $(BUILD_IMAGE)-$${arch}; \
		docker build \
			--load \
			--progress plain \
			--platform $${platform} \
			--tag $(IMAGE)-$${arch}:$(IMAGE_VERSION) \
			--tag $(IMAGE)-$${arch}:$(TAG)-$(CONFIG) \
			--tag $(IMAGE)-$${arch}:latest-$(CONFIG) \
			$(CONFIG) || exit 1; \
		rm -f $(CONFIG)/$(TAR_FILE); \
	done

.PHONY: push
push: container
	echo "Pushing $(IMGNAME) tags"
	@for platform in $(PLATFORMS); do \
		arch=$${platform##*/}; \
		echo "Pushing tags for $${platform} platform"; \
		docker push $(IMAGE)-$${arch}:$(IMAGE_VERSION) || exit 1; \
		docker push $(IMAGE)-$${arch}:$(TAG)-$(CONFIG) || exit 1; \
		docker push $(IMAGE)-$${arch}:latest-$(CONFIG) || exit 1; \
	done

.PHONY: manifest
manifest: push
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(shell echo $(ARCHS) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(IMAGE_VERSION)~g")
	@for platform in $(ARCHS); do docker manifest annotate --arch "$${platform}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${platform}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

.PHONY: clean
clean:
	@for arch in $(ARCHS); do \
		docker rmi -f $(IMAGE)-$${arch}:$(IMAGE_VERSION) || true; \
		docker rm  -f $(BUILD_IMAGE)-$${arch} || true; \
		docker rmi -f $(BUILD_IMAGE)-$${arch} || true; \
	done
	rm -f $(CONFIG)/$(TAR_FILE)
