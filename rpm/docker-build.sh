#!/bin/sh
set -e

sudo docker build -t kubelet-rpm-builder .
mkdir -p output
sudo docker run -ti --rm -v $PWD:/root/rpmbuild/SPECS -v $PWD/output/:/root/rpmbuild/RPMS/ kubelet-rpm-builder
sudo chown -R $USER $PWD/output

echo
echo "----------------------------------------"
echo
echo "RPMs written to: $PWD/output/x86_64/"
ls $PWD/output/x86_64/
echo
echo "Yum repodata written to: $PWD/output/x86_64/yum/"
