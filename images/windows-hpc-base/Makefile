# set default shell
SHELL=/bin/bash -o pipefail

REGISTRY ?= gcr.io/k8s-staging-build-image
IMG ?= windows-hpc-base-image
TAG ?=  v1

.PHONY: clean
clean:
	rm -rf build

.PHONY: image_build
image_build:
	./build-hpc-base-image.sh

.PHONY: image_push
image_push: image_build
	docker image import ./build/layer/layer.tar $(REGISTRY)/$(IMG):$(TAG)
	echo "Pushing $(REGISTRY)/$(IMG):$(TAG)"
	docker push $(REGISTRY)/$(IMG):$(TAG)