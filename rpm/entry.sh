#!/bin/sh
# Entrypoint for the build container to create the rpms and yum repodata:
# Usage: ./entry.sh GOARCH/RPMARCH,GOARCH/RPMARCH,....

set -e

declare -a ARCHS

if [ $# -gt 0 ]; then
  IFS=','; ARCHS=($1); unset IFS;
else
  #GOARCH/RPMARCH
  ARCHS=(
    amd64/x86_64
    arm/armhfp
    arm64/aarch64
    ppc64le/ppc64le
    s390x/s390x
  )
fi

if [[ -n "${KUBE_MAJOR}" ]]; then
  KUBE_MAJOR_ARG=(--define "KUBE_MAJOR ${KUBE_MAJOR}")
fi

if [[ -n "${KUBE_MINOR}" ]]; then
  KUBE_MINOR_ARG=(--define "KUBE_MINOR ${KUBE_MINOR}")
fi

if [[ -n "${KUBE_PATCH}" ]]; then
  KUBE_PATCH_ARG=(--define "KUBE_PATCH ${KUBE_PATCH}")
fi

if [[ -n "${RPM_RELEASE}" ]]; then
  RPM_RELEASE_ARG=(--define "RPM_RELEASE ${RPM_RELEASE}")
fi

for ARCH in "${ARCHS[@]}"; do
  IFS=/ read GOARCH RPMARCH<<< ${ARCH}; unset IFS;
  SRC_PATH="/root/rpmbuild/SOURCES/${RPMARCH}"
  mkdir -p "${SRC_PATH}"
  cp -r /root/common-sources/* "${SRC_PATH}"
  echo "Building RPM's for ${GOARCH}....."
  # Download sources if not already available
  cd "${SRC_PATH}"
  spectool \
    "${KUBE_MAJOR_ARG[@]}" \
    "${KUBE_MINOR_ARG[@]}" \
    "${KUBE_PATCH_ARG[@]}" \
    "${RPM_RELEASE_ARG[@]}" \
    -gf \
    /root/rpmbuild/SPECS/kubelet.spec

  /usr/bin/rpmbuild --target "${RPMARCH}" \
                    --define "ARCH ${GOARCH}" \
                    --define "_sourcedir ${SRC_PATH}" \
                    "${KUBE_MAJOR_ARG[@]}" \
                    "${KUBE_MINOR_ARG[@]}" \
                    "${KUBE_PATCH_ARG[@]}" \
                    "${RPM_RELEASE_ARG[@]}" \
                    -bb \
                    /root/rpmbuild/SPECS/kubelet.spec

  mkdir -p /root/rpmbuild/RPMS/"${RPMARCH}"
  createrepo -o /root/rpmbuild/RPMS/"${RPMARCH}"/ /root/rpmbuild/RPMS/"${RPMARCH}"
done
