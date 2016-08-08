FROM fedora:24
MAINTAINER Devan Goodwin <dgoodwin@redhat.com>

RUN dnf install -y rpm-build rpmdevtools createrepo && dnf clean all

RUN rpmdev-setuptree

USER root
ADD entry.sh /root/
CMD ["/root/entry.sh"]

