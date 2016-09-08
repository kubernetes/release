#!/bin/bash

ARCHITECTURES="amd64 arm arm64"
DISTROS="xenial"
K8S_VERSIONS="1.3.6 1.4.0-beta.0"
CNI_VERSION="0.4.0-alpha-07a8a2"


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
        go run build.go -arch "${arch}" -distro_name "${distro}" -package kubectl -version "${version}" -revision 01
      done
    done
  done
fi

if [[ $# == 0 || ! -z $(echo $@ | grep kubelet) ]]; then

  for version in ${K8S_VERSIONS}; do
    for arch in ${ARCHITECTURES}; do
      for distro in ${DISTROS}; do
        echo "kubelet] version: ${version} arch: ${arch} distro: ${distro}"
        go run build.go -arch "${arch}" -distro_name "${distro}" -package kubelet -version "${version}" -revision 01
      done
    done
  done
fi
