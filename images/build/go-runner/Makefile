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

REGISTRY ?= gcr.io/k8s-staging-build-image
IMGNAME = go-runner
IMAGE = $(REGISTRY)/$(IMGNAME)

TAG ?= $(shell git describe --tags --always --dirty)
IMAGE_VERSION ?= buster-v1.0.0
CONFIG ?= buster

# Build args
GO_VERSION ?= 1.13.15
DISTROLESS_IMAGE ?= static-debian10

PLATFORMS ?= linux/amd64 linux/arm64 linux/arm linux/ppc64le linux/s390x

HOST_GOOS ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)
GO_BUILD ?= go build

.PHONY: all build clean

.PHONY: all
all: build

.PHONY: build
build:
	$(GO_BUILD)

.PHONY: clean
clean:
	rm go-runner

.PHONY: container
container: init-docker-buildx
	echo "Building go-runner for the following platforms: $(PLATFORMS)"
	# https://github.com/docker/buildx/issues/59
	$(foreach PLATFORM,$(PLATFORMS), \
		echo "Starting build for $(PLATFORM) platform"; \
		DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build \
		--load \
		--progress plain \
		--platform $(PLATFORM) \
		--tag $(IMAGE)-$(PLATFORM):$(IMAGE_VERSION) \
		--tag $(IMAGE)-$(PLATFORM):$(TAG)-$(CONFIG) \
		--tag $(IMAGE)-$(PLATFORM):latest-$(CONFIG) \
		--build-arg=GO_VERSION=$(GO_VERSION) \
		--build-arg=DISTROLESS_IMAGE=$(DISTROLESS_IMAGE) .;)

.PHONY: push
push: container
	$(foreach PLATFORM,$(PLATFORMS), \
		docker push $(IMAGE)-$(PLATFORM):$(IMAGE_VERSION);)
	$(foreach PLATFORM,$(PLATFORMS), \
		docker push $(IMAGE)-$(PLATFORM):$(TAG)-$(CONFIG);)
	$(foreach PLATFORM,$(PLATFORMS), \
		docker push $(IMAGE)-$(PLATFORM):latest-$(CONFIG);)

.PHONY: manifest
manifest: push
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(shell echo $(PLATFORMS) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(IMAGE_VERSION)~g")
	@for arch in $(PLATFORMS); do docker manifest annotate --arch "$${arch##*/}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${arch}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	# Ensure qemu is in binfmt_misc
	# NOTE: Please always pin this to a digest for predictability/auditability
	# Last updated: 08/21/2020
	docker run --rm --privileged multiarch/qemu-user-static@sha256:96c2efd64db20bc326788773f6fa7cecef71b9d2c3de4c7d95a9910f38295c42 --reset -p yes

	# Ensure we use a builder that can leverage it (the default on linux will not)
	docker buildx rm multiarch-go-runner || true
	docker buildx create --name multiarch-go-runner --use
	docker buildx inspect --bootstrap
