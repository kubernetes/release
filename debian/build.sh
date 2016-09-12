#!/bin/bash

ARCHITECTURES="amd64 arm arm64"
DISTROS="xenial"
K8S_VERSIONS="1.3.6 1.4.0-beta.0"

# the cni package is named after the latest cni release (0.3.0) plus the k8s revision (1), plus the six first chars from the commit
# this means the kubelet can demand a specific k8s revision by requiring (>= 0.3.0.1)
CNI_VERSION="0.3.0.1-07a8a2"


if [[ -z $(dpkg --print-foreign-architectures | grep armhf) ]]; then
  sudo dpkg --add-architecture armhf
fi
if [[ -z $(dpkg --print-foreign-architectures | grep arm64) ]]; then
  sudo dpkg --add-architecture arm64
fi

if [[ $# == 0 || ! -z $(echo $@ | grep kubernetes-cni) ]]; then

  for arch in ${ARCHITECTURES}; do
    for distro in ${DISTROS}; do
      echo "cni] arch: ${arch} distro: ${distro}"
      go run build.go -arch "${arch}" -distro_name "${distro}" -package kubernetes-cni -version ${CNI_VERSION} -revision 00
    done
  done
fi

if [[ $# == 0 || ! -z $(echo $@ | grep kubectl) ]]; then

  for version in ${K8S_VERSIONS}; do
    for arch in ${ARCHITECTURES}; do
      for distro in ${DISTROS}; do
        echo "kubectl] version: ${version} arch: ${arch} distro: ${distro}"
        go run build.go -arch "${arch}" -distro_name "${distro}" -package kubectl -version "${version}" -revision 02
      done
    done
  done
fi

if [[ $# == 0 || ! -z $(echo $@ | grep kubelet) ]]; then

  for version in ${K8S_VERSIONS}; do
    for arch in ${ARCHITECTURES}; do
      for distro in ${DISTROS}; do
        echo "kubelet] version: ${version} arch: ${arch} distro: ${distro}"
        go run build.go -arch "${arch}" -distro_name "${distro}" -package kubelet -version "${version}" -revision 02
      done
    done
  done
fi
