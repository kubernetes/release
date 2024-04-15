# Copyright 2019 The Kubernetes Authors.
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

# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

.DEFAULT_GOAL:=help
SHELL:=/usr/bin/env bash

COLOR:=\\033[36m
NOCOLOR:=\\033[0m

##@ Verify

.PHONY: verify verify-boilerplate verify-build verify-dependencies verify-go-mod

verify: release-tools verify-boilerplate verify-build verify-dependencies verify-go-mod ## Runs verification scripts to ensure correct execution

verify-boilerplate: ## Runs the file header check
	./hack/verify-boilerplate.sh

verify-build: ## Builds the project for a chosen set of platforms
	./hack/verify-build.sh

verify-dependencies: ## Runs zeitgeist to verify dependency versions
	./hack/verify-dependencies.sh

verify-go-mod: ## Runs the go module linter
	./hack/verify-go-mod.sh

##@ Tests

.PHONY: test
test: test-go-unit test-sh ## Runs unit tests to ensure correct execution

.PHONY: test-go-unit
test-go-unit: ## Runs golang unit tests
	./hack/test-go.sh

.PHONY: test-go-integration
test-go-integration: ## Runs golang integration tests
	./hack/test-go-integration.sh

.PHONY: test-sh
test-sh: ## Runs all shellscript tests
	./hack/test-sh.sh

##@ Tools

RELEASE_TOOLS ?=

.PHONY: release-tools

release-tools: ## Compiles a set of release tools, specified by $RELEASE_TOOLS
	./compile-release-tools $(RELEASE_TOOLS)

##@ Images

.PHONY: update-images

images := \
	k8s-cloud-builder

update-images: $(addprefix image-,$(images)) ## Update all images in ./images/
image-%:
	$(eval img := $(subst image-,,$@))
	gcloud builds submit --config './images/$(img)/cloudbuild.yaml' './images/$(img)'

RUNTIME ?= docker
LOCALIMAGE_NAME := k8s-cloud-builder

.PHONY: local-image-build
local-image-build: ## Build a local image to use the tools of this repository on non Debian/Ubuntu/Fedora distributions
	$(RUNTIME) build \
		-f images/k8s-cloud-builder/Dockerfile \
		-t $(LOCALIMAGE_NAME)

.PHONY: local-image-run
local-image-run: ## Run a locally build image to use the tools of this repository on non Debian/Ubuntu/Fedora distributions
	$(RUNTIME) run -it \
		-v $$HOME/.config/gcloud:/root/.config/gcloud \
		-v $(shell pwd):/go/src/k8s.io/release \
		-w /go/src/k8s.io/release \
		$(LOCALIMAGE_NAME) bash

##@ Dependencies

.SILENT: update-deps update-deps-go update-mocks
.PHONY:  update-deps update-deps-go update-mocks

update-deps: update-deps-go ## Update all dependencies for this repo
	echo -e "${COLOR}Commit/PR the following changes:${NOCOLOR}"
	git status --short

update-deps-go: GO111MODULE=on
update-deps-go: ## Update all golang dependencies for this repo
	go get -u -t ./...
	go mod tidy
	go mod verify
	$(MAKE) test-go-unit
	./hack/update-all.sh

update-mocks: ## Update all generated mocks
	go generate ./...
	for f in $(shell find . -name fake_*.go); do \
		if ! grep -q "^`cat hack/boilerplate/boilerplate.generatego.txt`" $$f; then \
			cp hack/boilerplate/boilerplate.generatego.txt tmp ;\
			cat $$f >> tmp ;\
			mv tmp $$f ;\
		fi \
	done

##@ Helpers

.PHONY: help

help:  ## Display this help
	@awk \
		-v "col=${COLOR}" -v "nocol=${NOCOLOR}" \
		' \
			BEGIN { \
				FS = ":.*##" ; \
				printf "\nUsage:\n  make %s<target>%s\n", col, nocol \
			} \
			/^[a-zA-Z_-]+:.*?##/ { \
				printf "  %s%-15s%s %s\n", col, $$1, nocol, $$2 \
			} \
			/^##@/ { \
				printf "\n%s%s%s\n", col, substr($$0, 5), nocol \
			} \
		' $(MAKEFILE_LIST)
