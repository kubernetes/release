#!/bin/sh
set -e

docker build -t kubelet-rpm-builder .
echo "Cleaning output directory..."
sudo rm -rf output/*
mkdir -p output
docker run -ti --rm -v $PWD/output/:/root/rpmbuild/RPMS/ kubelet-rpm-builder
sudo chown -R $USER $PWD/output

echo
echo "----------------------------------------"
echo
echo "RPMs written to: $PWD/output/x86_64/"
ls $PWD/output/x86_64/
echo
echo "Yum repodata written to: $PWD/output/x86_64/yum/"
