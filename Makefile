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

##@ Package

.PHONY: verify-published-debs verify-published-rpms

verify-published-debs: ## Ensure debs have been published
	./hack/packages/verify-published.sh debs

verify-published-rpms: ## Ensure rpms have been published
	./hack/packages/verify-published.sh rpms

##@ Verify

.PHONY: verify verify-shellcheck

# TODO: Uncomment verify-shellcheck once we finish shellchecking the repo.
#       ref: https://github.com/kubernetes/release/issues/726
verify: #verify-shellcheck ## Runs verification scripts to ensure correct execution
	@echo consider make verify-bazel as well

verify-shellcheck: ## Runs shellcheck
	./hack/verify-shellcheck.sh

verify-bazel:
	bazel test //...

##@ Tests

.PHONY: test
test: test-go test-sh ## Runs unit tests to ensure correct execution

.PHONY: test-go
test-go: ## Runs all golang tests
	./hack/test-go.sh

.PHONY: test-sh
test-sh: ## Runs all shellscript tests
	./hack/test-sh.sh

##@ Images

.PHONY: update-images

images := \
	k8s-cloud-builder

update-images: $(addprefix image-,$(images)) ## Update all images in ./images/
image-%:
	$(eval img := $(subst image-,,$@))
	gcloud builds submit --config './images/$(img)/cloudbuild.yaml' './images/$(img)'

##@ Dependencies

.SILENT: update-deps update-deps-go
.PHONY:  update-deps update-deps-go

update-deps: update-deps-go ## Update all dependencies for this repo
	echo -e "${COLOR}Commit/PR the following changes:${NOCOLOR}"
	git status --short

update-deps-go: GO111MODULE=on
update-deps-go: ## Update all golang dependencies for this repo
	go get -u -t ./...
	go mod tidy
	go mod verify
	$(MAKE) test-go
	./hack/update-all.sh

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
