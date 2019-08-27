FROM fedora:30

RUN dnf install -y \
      rpm-build \
      rpmdevtools \
      createrepo \
    && dnf clean all

RUN rpmdev-setuptree

RUN useradd builder -u 1000 -m -G users,wheel

ADD build.sh /home/builder
COPY . /home/builder/rpmbuild/SPECS

RUN chown -R builder:builder /home/builder

USER builder

ENTRYPOINT ["/home/builder/build.sh"]
