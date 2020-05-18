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

.PHONY:	build push

STAGING_REGISTRY?=gcr.io/k8s-staging-build-image
PROD_REGISTRY?=us.gcr.io/k8s-artifacts-prod/build-image
IMAGE=kube-cross

TAG?=$(shell git describe --tags --always --dirty)
CONFIG?=go1.14
KUBE_CROSS_VERSION?=v1.14.3-1

# Build args
GO_VERSION?=1.14.3
PROTOBUF_VERSION?=3.0.2
ETCD_VERSION?=v3.4.7

# TODO: Support multi-arch kube-cross images
#       ref:
#       - https://github.com/kubernetes/release/pull/1140/commits/e96924f58d9ce2bb73045fbba13f259b4345c0d9#r395056830
#       - https://github.com/kubernetes/kubernetes/issues/78964
#       - https://github.com/kubernetes/kubernetes/issues/75114
ARCH=amd64

all: build push

build:
	docker build \
		-t $(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(TAG)-$(CONFIG) \
		-t $(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(KUBE_CROSS_VERSION) \
		-t $(PROD_REGISTRY)/$(IMAGE)-$(ARCH):$(KUBE_CROSS_VERSION) \
		--build-arg=GO_VERSION=$(GO_VERSION) \
		--build-arg=PROTOBUF_VERSION=$(PROTOBUF_VERSION) \
		--build-arg=ETCD_VERSION=$(ETCD_VERSION) \
		.

push:
	docker push $(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(TAG)-$(CONFIG)
	docker push $(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(KUBE_CROSS_VERSION)
	docker manifest create --amend $(STAGING_REGISTRY)/$(IMAGE):$(KUBE_CROSS_VERSION) \
		$(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(KUBE_CROSS_VERSION)
	docker manifest annotate $(STAGING_REGISTRY)/$(IMAGE):$(KUBE_CROSS_VERSION) \
		$(STAGING_REGISTRY)/$(IMAGE)-$(ARCH):$(KUBE_CROSS_VERSION) \
		--arch $(ARCH)
	docker manifest push -p $(STAGING_REGISTRY)/$(IMAGE):$(KUBE_CROSS_VERSION)
