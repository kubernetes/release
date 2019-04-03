#!/bin/sh
# Entrypoint for the build container to create the rpms and yum repodata:
set -e
REQDARGS=2
USAGE="./entry.sh PACKAGE VERSION [GOARCH/RPMARCH,GOARCH/RPMARCH,....]"
if [ $# -lt ${REQDARGS} ]; then
  echo ${USAGE}
  exit 1
fi

PACKAGE=$1
VERSION=$2
USERARCHS=$3
if [ -n "${USERARCHS}" ]; then
  IFS=','; ARCHS=${USERARCHS}; unset IFS;
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
echo "Building ${PACKAGE} RPMs at ${VERSION}:"

if [ "${PACKAGE}" = "kubernetes" ]; then
  if [ -f "/root/rpmbuild/SPECS/${VERSION}/kubernetes.spec" ]; then
    ln -s /root/rpmbuild/SPECS/${VERSION}/kubernetes.spec /root/rpmbuild/SPECS/kubernetes.spec
  else
    echo "invalid version: ${VERSION}"
    echo "missing ${VERSION}/kubernetes.spec"
    exit 1
  fi
fi

for ARCH in ${ARCHS[@]}; do
  IFS=/ read GOARCH RPMARCH<<< ${ARCH}; unset IFS;
  SRC_PATH="/root/rpmbuild/SOURCES/${RPMARCH}"
  mkdir -p ${SRC_PATH}
  cp -r /root/rpmbuild/SPECS/* ${SRC_PATH}
  echo "Building ${PACKAGE} RPM at ${VERSION} for ${GOARCH}....."
  sed -i "s/\%global ARCH.*/\%global ARCH ${GOARCH}/" ${SRC_PATH}/${PACKAGE}.spec
  # Download sources if not already available
  cd ${SRC_PATH} && spectool -gf ${PACKAGE}.spec
  /usr/bin/rpmbuild --target ${RPMARCH} --define "_sourcedir ${SRC_PATH}" -bb ${SRC_PATH}/${PACKAGE}.spec
  mkdir -p /root/rpmbuild/RPMS/${RPMARCH}
  createrepo -o /root/rpmbuild/RPMS/${RPMARCH}/ /root/rpmbuild/RPMS/${RPMARCH}
done
