# Copyright 2022 The Kubernetes Authors.
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

.PHONY:	build push all all-build all-push-images all-push push-manifest

REGISTRY?="gcr.io/k8s-staging-build-image"
IMAGE=$(REGISTRY)/distroless-iptables

TAG ?= $(shell git describe --tags --always --dirty)
IMAGE_VERSION ?= v0.7.6
CONFIG ?= distroless-bookworm
BASEIMAGE ?= debian:bookworm-slim
GORUNNER_VERSION ?= v2.4.0-go1.24.4-bookworm.0

ARCH?=amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x

BASE_REGISTRY?=registry.k8s.io/build-image

GORUNNERIMAGE?=$(BASE_REGISTRY)/go-runner:$(GORUNNER_VERSION)

QEMUVERSION=7.2.0-1
QEMUIMAGE ?= multiarch/qemu-user-static

ifneq ($(ARCH), amd64)
	SKIP_WRAPPER_CHECK=--no-sanity-check
endif

# This option is for running docker manifest command
export DOCKER_CLI_EXPERIMENTAL := enabled

SUDO=$(if $(filter 0,$(shell id -u)),,sudo)

build:
	# Fix possible issues with the local umask
	umask 0022

	# Enable execution of multi-architecture containers
	docker run --rm --privileged $(QEMUIMAGE):$(QEMUVERSION) --reset -p yes
	docker buildx version
	BUILDER=$(shell docker buildx create --use)
	docker buildx build \
		--pull \
		--load \
		--platform linux/$(ARCH) \
		-t $(IMAGE)-$(ARCH):$(IMAGE_VERSION) \
		-t $(IMAGE)-$(ARCH):$(TAG)-$(CONFIG) \
		-t $(IMAGE)-$(ARCH):latest-$(CONFIG) \
		--build-arg=BASEIMAGE=$(BASEIMAGE) \
		--build-arg=GORUNNERIMAGE=$(GORUNNERIMAGE) \
		--build-arg=SKIP_WRAPPER_CHECK=$(SKIP_WRAPPER_CHECK) \
		$(CONFIG)
	docker buildx rm $$BUILDER

push: build
	docker push $(IMAGE)-$(ARCH):$(IMAGE_VERSION)
	docker push $(IMAGE)-$(ARCH):$(TAG)-$(CONFIG)
	docker push $(IMAGE)-$(ARCH):latest-$(CONFIG)

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

all: all-push
