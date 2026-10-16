# Copyright 2021 The Kubernetes Authors.
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
include $(CURDIR)/../../Makefile.common-image

IMGNAME = setcap
IMAGE_VERSION ?= bookworm-v1.0.9
CONFIG ?= bookworm
DEBIAN_BASE_VERSION ?= bookworm-v1.0.9

# Build args
BASE_REGISTRY ?= registry.k8s.io/build-image
BASEIMAGE ?= $(BASE_REGISTRY)/debian-base:$(DEBIAN_BASE_VERSION)

PLATFORMS ?= linux/amd64 linux/arm64 linux/ppc64le linux/s390x

BUILD_ARGS = --pull \
             --build-arg=BASEIMAGE=$(BASEIMAGE)
