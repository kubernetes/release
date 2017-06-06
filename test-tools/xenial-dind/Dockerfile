FROM ubuntu:xenial

STOPSIGNAL SIGRTMIN+3

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get update \
    && apt-get -y upgrade \
    && apt-get -y install systemd \
    && cd /lib/systemd/system/ \
    && ls sysinit.target.wants/* multi-user.target.wants/* local-fs.target.wants/* | grep -v systemd-tmpfiles-setup.service \
        | xargs -n 1 rm -f  -v \
    && ls sockets.target.wants/*initctl* \
        | xargs -n 1 rm -f -v \
    && ls graphical.target.wants/* \
        | xargs -n 1 rm -f -v \
    && ls /etc/systemd/system/*.wants/* \
        | xargs -n 1 rm -f -v \
    && ln -v -f /lib/systemd/system/multi-user.target /lib/systemd/system/default.target

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install apt-transport-https ca-certificates \
    && apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D \
    && echo deb https://apt.dockerproject.org/repo ubuntu-xenial main > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get -y install docker-engine=1.12.1-0~xenial \
    && apt-get -y install socat \
    && apt-get -y autoremove \
    && apt-get clean \
    && rm -r -f /var/lib/apt/lists/* /tmp/* /var/tmp/*

CMD [ "/sbin/init" ]
