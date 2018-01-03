#!/bin/sh
set -e

# Clean the output directory first, so any RPMs that were built during a previous run are not
# included in the context sent when building the image.
echo "Cleaning output directory"
sudo rm -rf output

# Build the image
docker build -t kubelet-rpm-builder .

# Prep the output directory where the built RPMs go
mkdir -p output

if [[ -z "${KEEP_SOURCES}" ]]; then
  echo "Cleaning sources directory"
  rm -rf sources/*
else
  echo "Using existing sources directory"
fi

# Allow overriding the major.minor.patch-release values.
if [[ -n "${KUBE_MAJOR}" ]]; then
  KUBE_MAJOR_ARG=(-e KUBE_MAJOR=${KUBE_MAJOR})
fi

if [[ -n "${KUBE_MINOR}" ]]; then
  KUBE_MINOR_ARG=(-e KUBE_MINOR=${KUBE_MINOR})
fi

if [[ -n "${KUBE_PATCH}" ]]; then
  KUBE_PATCH_ARG=(-e KUBE_PATCH=${KUBE_PATCH})
fi

if [[ -n "${RPM_RELEASE}" ]]; then
  RPM_RELEASE_ARG=(-e RPM_RELEASE=${RPM_RELEASE})
fi

# Build the RPMs. The 'sources' directory is bind mounted into the container, so you can
# pre-download sources if you wish.
docker run --rm \
  -v "$PWD"/output/:/root/rpmbuild/RPMS/ \
  -v "$PWD"/sources/x86_64/:/root/rpmbuild/SOURCES/x86_64/ \
  -v "$PWD"/sources/armhfp/:/root/rpmbuild/SOURCES/armhfp/ \
  -v "$PWD"/sources/aarch64/:/root/rpmbuild/SOURCES/aarch64/ \
  -v "$PWD"/sources/ppc64le/:/root/rpmbuild/SOURCES/ppc64le/ \
  -v "$PWD"/sources/s390x/:/root/rpmbuild/SOURCES/s390x/ \
  "${KUBE_MAJOR_ARG[@]}" \
  "${KUBE_MINOR_ARG[@]}" \
  "${KUBE_PATCH_ARG[@]}" \
  "${RPM_RELEASE_ARG[@]}" \
  kubelet-rpm-builder \
  "$1"
sudo chown -R "$USER" "$PWD"/output

echo
echo "----------------------------------------"
echo
echo "RPMs written to: "
ls "$PWD"/output/*/
echo
echo "Yum repodata written to: "
ls "$PWD"/output/*/repodata/
