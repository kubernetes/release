#!/bin/bash -eux

kube_release_tag="v1.4.0-beta.7"

source_image_prefix="gcr.io/google_containers"

# TODO discovery
kube_source_images=(
  hyperkube
  kube-apiserver
  kube-controller-manager
  kube-scheduler
  kube-proxy
)

etcd_release_tag="2.2.5"
kubedns_release_tag="1.7"
kube_dnsmasq_release_tag="1.3"
exechealthz_release_tag="1.1"

architectures=(
  amd64
  arm64
  arm
)

multiarch_prefix="docker.io/multiarchkubernetes"

head="image: \"${multiarch_prefix}/%s\"\nmanifests:\n"
item="%s {image: \"${multiarch_prefix}/%s\", platform: {architecture: %s, os: linux}}\n"

for i in "${kube_source_images[@]}" ; do
  printf "${head}" "${i}:${kube_release_tag}" > "${i}.yaml"
  for a in "${architectures[@]}" ; do
    image="${i}-${a}:${kube_release_tag}"
    printf "${item}" "-" "${image}" "${a}" >> "${i}.yaml"
    docker pull "${source_image_prefix}/${image}"
    docker tag "${source_image_prefix}/${image}" "${multiarch_prefix}/${image}"
    docker push "${multiarch_prefix}/${image}"
  done
done


for i in "etcd" "kubedns" "kube-dnsmasq" "exechealthz" ; do
  v="${i/-/_}_release_tag"
  release_tag="${!v}"
  printf "${head}" "${i}:${release_tag}" > "${i}.yaml"
  for a in "${architectures[@]}" ; do
    image="${i}-${a}:${release_tag}"
    printf "${item}" "-" "${image}" "${a}" >> "${i}.yaml"
    docker pull "${source_image_prefix}/${image}"
    docker tag "${source_image_prefix}/${image}" "${multiarch_prefix}/${image}"
    docker push "${multiarch_prefix}/${image}"
  done
done
