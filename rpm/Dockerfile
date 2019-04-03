FROM fedora:29
MAINTAINER kubernetes-sig-release@googlegroups.com

RUN dnf install -y rpm-build rpmdevtools createrepo && dnf clean all

RUN rpmdev-setuptree

USER root
ADD entry.sh /root/
COPY ./ /root/rpmbuild/SPECS
ENTRYPOINT ["/root/entry.sh"]
