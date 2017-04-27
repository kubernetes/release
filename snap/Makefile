KUBE_VERSION=$(shell curl -L https://dl.k8s.io/release/stable.txt)

ifndef VERBOSE
	MAKEFLAGS += --no-print-directory
endif

targets = kubectl kube-apiserver kube-controller-manager kube-scheduler kubelet kube-proxy kubeadm kubefed kubernetes-test

build = ./build-scripts/build

.PHONY: $(targets)

default:
	@KUBE_VERSION=${KUBE_VERSION} ${build} $(targets)

clean:
	@rm -rf build

kubectl:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kubectl

kubeadm:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kubeadm

kubefed:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kubefed

kube-apiserver:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kube-apiserver

kube-controller-manager:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kube-controller-manager

kube-scheduler:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kube-scheduler

kubelet:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kubelet

kube-proxy:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kube-proxy

kubernetes-test:
	@KUBE_VERSION=${KUBE_VERSION} ${build} kubernetes-test
