#!/bin/sh
# Entrypoint for the build container to create the rpms and yum repodata:

set -e

/usr/bin/rpmbuild --define "_sourcedir /root/rpmbuild/SPECS/" -bb /root/rpmbuild/SPECS/kubelet.spec

mkdir -p /root/rpmbuild/RPMS/x86_64/
createrepo -o /root/rpmbuild/RPMS/x86_64/ /root/rpmbuild/RPMS/x86_64
