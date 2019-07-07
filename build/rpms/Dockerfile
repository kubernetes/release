FROM fedora:24

RUN dnf install -y \
      rpm-build \
      rpmdevtools \
      createrepo \
    && dnf clean all

RUN rpmdev-setuptree

USER root

ADD build.sh /root/
COPY . /root/rpmbuild/SPECS

ENTRYPOINT ["/root/build.sh"]
