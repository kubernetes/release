FROM golang:1.12 AS builder

ENV GO111MODULE=on

WORKDIR /workspace

COPY build.go go.mod go.sum ./

RUN go build ./...


FROM debian:buster

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -y \
    && apt-get -yy -q install --no-install-recommends --no-install-suggests --fix-missing \
      dpkg-dev \
      build-essential \
      ca-certificates \
      curl \
      debhelper \
      dh-systemd \
      fakeroot \
    && apt-get upgrade -y \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN useradd builder -u 9000 -m -s /bin/false

USER builder

RUN mkdir /home/builder/workspace

WORKDIR /home/builder/workspace

RUN mkdir ./bin

COPY --chown=builder:builder bionic ./bionic
COPY --chown=builder:builder jessie ./jessie
COPY --chown=builder:builder sid ./sid
COPY --chown=builder:builder stretch ./stretch
COPY --chown=builder:builder trusty ./trusty
COPY --chown=builder:builder xenial ./xenial

COPY --from=builder --chown=builder:builder /workspace/deb-builder .

ENTRYPOINT ["./deb-builder"]
