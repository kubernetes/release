FROM ruby

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get update -y \
    && apt-get -yy -q install --no-install-recommends --no-install-suggests --fix-missing \
      python3 \
      dpkg-dev \
      build-essential \
      debhelper \
      rpm \
      bsdtar \
    && apt-get upgrade -y \
    && apt-get autoremove -y \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && gem install fpm package_cloud

ADD . /src

WORKDIR /src

ENV PATH "/usr/local/bundle/bin:${PATH}"

CMD [ "make" ]
