FROM fedora:27
MAINTAINER Devan Goodwin <dgoodwin@redhat.com>

RUN dnf install -y rpm-build rpmdevtools createrepo && dnf clean all

RUN rpmdev-setuptree

USER root
ADD entry.sh /root/

# Only put the spec file in SPECS
COPY kubelet.spec /root/rpmbuild/SPECS

# Put sources that don't vary per arch in a common directory. The entry.sh script copies them to the
# appropriate source directory based on the arch being built.
COPY 10-kubeadm-post-1.8.conf 10-kubeadm-pre-1.8.conf kubelet.service /root/common-sources/

ENTRYPOINT ["/root/entry.sh"]
