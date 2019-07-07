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

##@ Package

.PHONY: build-debs build-rpms verify-published-debs verify-published-rpms

build-debs: ## Build debs
	PACKAGE_TYPE="debs" ./build/package-in-docker.sh

build-rpms: ## Build rpms
	PACKAGE_TYPE="rpms" ./build/package-in-docker.sh
	
verify-published-debs: ## Ensure debs have been published
	./hack/packages/verify-published.sh debs

verify-published-rpms: ## Ensure rpms have been published
	./hack/packages/verify-published.sh rpms

##@ Verify

.PHONY: verify verify-shellcheck

# TODO: Uncomment verify-shellcheck once we finish shellchecking the repo.
#       ref: https://github.com/kubernetes/release/issues/726
verify: #verify-shellcheck ## Runs verification scripts to ensure correct execution

verify-shellcheck: ## Runs shellcheck
	./hack/verify-shellcheck.sh

##@ Tests

.PHONY: test
test: test-go test-sh ## Runs unit tests to ensure correct execution

.PHONY: test-go 
test-go: ## Run all golang tests
	./hack/test-go.sh
	
.PHONY: test-sh
test-sh: ## Runs all shellscript tests
	./hack/test-sh.sh
##@ Helpers

.PHONY: help

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
