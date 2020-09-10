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
IMGNAME = kube-cross
IMAGE = $(REGISTRY)/$(IMGNAME)

TAG ?= $(shell git describe --tags --always --dirty)
IMAGE_VERSION ?= v1.15.1-2
CONFIG ?= go1.15

# Build args
GO_VERSION?=1.15.1
PROTOBUF_VERSION?=3.0.2
ETCD_VERSION?=v3.4.13

# TODO: Support multi-arch kube-cross images
#       ref:
#       - https://github.com/kubernetes/release/pull/1140/commits/e96924f58d9ce2bb73045fbba13f259b4345c0d9#r395056830
#       - https://github.com/kubernetes/kubernetes/issues/78964
#       - https://github.com/kubernetes/kubernetes/issues/75114
PLATFORMS ?= linux/amd64 #linux/arm linux/arm64 linux/ppc64le linux/s390x

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
			--tag $(IMAGE)-$${platform##*/}:$(TAG)-$(CONFIG) \
			--tag $(IMAGE)-$${platform##*/}:latest-$(CONFIG) \
			--build-arg=GO_VERSION=$(GO_VERSION) \
			--build-arg=PROTOBUF_VERSION=$(PROTOBUF_VERSION) \
			--build-arg=ETCD_VERSION=$(ETCD_VERSION) \
			.; \
	done

.PHONY: push
push: container
	echo "Pushing $(IMGNAME) tags"
	@for platform in $(PLATFORMS); do \
		echo "Pushing tags for $${platform} platform"; \
		docker push $(IMAGE)-$${platform##*/}:$(IMAGE_VERSION); \
		docker push $(IMAGE)-$${platform##*/}:$(TAG)-$(CONFIG); \
		docker push $(IMAGE)-$${platform##*/}:latest-$(CONFIG); \
	done

.PHONY: manifest
manifest: push
	docker manifest create --amend $(IMAGE):$(IMAGE_VERSION) $(shell echo $(PLATFORMS) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(IMAGE_VERSION)~g")
	@for arch in $(PLATFORMS); do docker manifest annotate --arch "$${arch##*/}" ${IMAGE}:${IMAGE_VERSION} ${IMAGE}-$${arch}:${IMAGE_VERSION}; done
	docker manifest push --purge $(IMAGE):$(IMAGE_VERSION)

# enable buildx
.PHONY: init-docker-buildx
init-docker-buildx:
	./../../../hack/init-buildx.sh
