FROM fedora:30

RUN dnf install -y \
      rpm-build \
      rpmdevtools \
      createrepo \
    && dnf clean all

RUN useradd builder -u 9000 -m -s /bin/false

USER builder

RUN mkdir /home/builder/workspace

WORKDIR /home/builder/workspace

RUN rpmdev-setuptree

ADD --chown=builder:builder build.sh .
COPY --chown=builder:builder . ./SPECS

ENTRYPOINT ["./build.sh"]
