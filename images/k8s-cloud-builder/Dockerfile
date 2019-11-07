FROM k8s.gcr.io/kube-cross:v1.13.4-1

ARG DEBIAN_FRONTEND=noninteractive

# Install packages
RUN apt-get -q update \
    && apt-get install -qqy \
        apt-transport-https \
        build-essential \
        ca-certificates \
        curl \
        gettext-base \
        git \
        gnupg2 \
        jq \
        lsb-release \
        pandoc \
        python \
        python-dev \
        python-setuptools \
        software-properties-common \
        wget

# Install Pip packages
RUN python /usr/lib/python2.7/dist-packages/easy_install.py pip \
    && pip install --no-cache-dir yq

# Packages required by the make in k8s
# localtime
RUN apt-get -q update \
    && apt-get install -qqy tzdata

# install net tools
# required by common.sh
RUN apt-get -q update \
    && apt-get install -qqy \
        grep \
        net-tools \
        rsync

# Install gcloud
# common::set_cloud_binaries() looks for it in this path
RUN mkdir /opt/google
RUN apt-get -y update \
    && apt-get -y install \
        gcc \
        python2.7 \
        python-dev \
        python-setuptools \
        wget \
        ca-certificates \
        # These are necessary for add-apt-respository
        software-properties-common \
    # Install Git >2.0.1
    && apt-get -y update \
    && apt-get -y install git \
    # Setup Google Cloud SDK (latest)
    && curl -sSL https://sdk.cloud.google.com > /tmp/install.sh \
    && bash /tmp/install.sh --install-dir=/opt/google --disable-prompts \
    # install crcmod: https://cloud.google.com/storage/docs/gsutil/addlhelp/CRC32CandInstallingcrcmod
    && pip install -U crcmod \
    # Clean up
    && apt-get -y remove \
        gcc \
        python-dev \
        python-setuptools \
        wget \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf ~/.config/gcloud

ENV PATH=/opt/google/google-cloud-sdk/bin:$PATH

# Install docker stuff
#---------------------
# Based on instructions from:
# https://docs.docker.com/install/linux/docker-ce/debian/
RUN apt-get -y update \
    && apt-get -y install \
        apt-transport-https \
        ca-certificates \
        curl \
        make \
        software-properties-common \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - \
    && apt-key fingerprint 0EBFCD88 \
    && add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/debian \
      $(lsb_release -cs) \
      stable" \
    && apt-get -y update \
    && apt-get -y install \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        unzip

# Cleanup a bit
RUN rm -rf /var/lib/apt/lists/*

ENTRYPOINT []
