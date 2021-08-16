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

all: all-build

REGISTRY ?= gcr.io/k8s-staging-build-image
IMAGE ?= $(REGISTRY)/debian-base
BUILD_IMAGE ?= debian-build

TAG ?= $(shell git describe --tags --always --dirty)
IMAGE_VERSION ?= buster-v1.9.0
CONFIG ?= buster

TAR_FILE ?= rootfs.tar
ARCH ?= amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x

QEMUVERSION=5.2.0-2

# This option is for running docker manifest command
export DOCKER_CLI_EXPERIMENTAL := enabled

ifeq ($(ARCH),amd64)
	BASEIMAGE?=debian:$(CONFIG)-slim
	QEMUARCH=x86_64
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=arm32v7/debian:$(CONFIG)-slim
	QEMUARCH=arm
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=arm64v8/debian:$(CONFIG)-slim
	QEMUARCH=aarch64
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/debian:$(CONFIG)-slim
	QEMUARCH=ppc64le
endif
ifeq ($(ARCH),s390x)
	BASEIMAGE?=s390x/debian:$(CONFIG)-slim
	QEMUARCH=s390x
endif

sub-build-%:
	$(MAKE) ARCH=$* build

all-build: $(addprefix sub-build-,$(ALL_ARCH))

sub-push-image-%:
	$(MAKE) ARCH=$* push

all-push-images: $(addprefix sub-push-image-,$(ALL_ARCH))

all-push: all-push-images push-manifest

push-manifest:
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(shell echo $(ALL_ARCH) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(IMAGE_VERSION)~g")
	@for arch in $(ALL_ARCH); do docker manifest annotate --arch $${arch} ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${arch}:${IMAGE_VERSION}; done
	docker manifest push --purge ${IMAGE}:${IMAGE_VERSION}

build: clean
	# Fix possible issues with the local umask
	umask 0022

	# Enable execution of multi-architecture containers
	docker run --rm --privileged multiarch/qemu-user-static:$(QEMUVERSION) --reset -p yes
	docker buildx version
	BUILDER=$(shell docker buildx create --use)

	docker buildx build \
		--pull \
		--load \
		--platform linux/$(ARCH) \
		--build-arg BASEIMAGE=$(BASEIMAGE) \
		--build-arg ARCH=$(QEMUARCH) \
		-t $(BUILD_IMAGE) \
		-f $(CONFIG)/Dockerfile.build \
		$(CONFIG)
	docker create --name $(BUILD_IMAGE) $(BUILD_IMAGE)
	docker export $(BUILD_IMAGE) > $(CONFIG)/$(TAR_FILE)
	docker buildx build \
		--load \
		--platform linux/$(ARCH) \
		-t $(IMAGE)-$(ARCH):$(IMAGE_VERSION) \
		-t $(IMAGE)-$(ARCH):$(TAG)-$(CONFIG) \
		-t $(IMAGE)-$(ARCH):latest-$(CONFIG) \
		$(CONFIG)
	docker buildx rm $$BUILDER
	rm $(CONFIG)/rootfs.tar

push: build
	docker push $(IMAGE)-$(ARCH):$(IMAGE_VERSION)
	docker push $(IMAGE)-$(ARCH):$(TAG)-$(CONFIG)
	docker push $(IMAGE)-$(ARCH):latest-$(CONFIG)

clean:
	docker rmi -f $(IMAGE)-$(ARCH):$(IMAGE_VERSION) || true
	docker rmi -f $(BUILD_IMAGE)   || true
	docker rm  -f $(BUILD_IMAGE)   || true
