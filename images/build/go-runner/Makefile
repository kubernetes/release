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
include $(CURDIR)/../../Makefile.common-image

IMGNAME = go-runner
APP_VERSION = $(shell cat VERSION)
GO_MINOR_VERSION ?= 1.16
OS_CODENAME ?= buster
REVISION ?= 0

# Build args
GO_VERSION ?= 1.16.4
DISTROLESS_IMAGE ?= static-debian10

# Configuration
CONFIG = go$(GO_MINOR_VERSION)-$(OS_CODENAME)
BUILD_METADATA = go$(GO_VERSION)-$(OS_CODENAME).$(REVISION)

IMAGE_VERSION = $(APP_VERSION)-$(BUILD_METADATA)

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

BUILD_ARGS = --build-arg=GO_VERSION=$(GO_VERSION) \
             --build-arg=DISTROLESS_IMAGE=$(DISTROLESS_IMAGE)
