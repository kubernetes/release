# To rebuild and publish this container run:
#   gcloud builds submit --config update_build_container.yaml .

FROM ubuntu:18.04

# Install packages
RUN apt-get -q update && apt-get install -qqy apt-transport-https \
    ca-certificates curl git gnupg2 lsb-release python \
    software-properties-common wget python-setuptools python-dev \
    build-essential jq pandoc gettext-base

# Install Pip packages
RUN easy_install pip
RUN pip install yq

# Packages required by the make in k8s
# localtime
RUN apt-get -q update && apt-get install -qqy tzdata

# install net tools
# required by common.sh
RUN apt-get -q update && apt-get install -qqy grep net-tools rsync

# Install gcloud
# common::set_cloud_binaries() looks for it in this path
RUN mkdir /opt/google
RUN curl -sSL https://sdk.cloud.google.com > /tmp/install.sh && \
    bash /tmp/install.sh --install-dir=/opt/google --disable-prompts


# Install docker stuff
#---------------------
# Based on instructions from:
# https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/#uninstall-old-versions
RUN \
   apt-get -y update && \
   apt-get install -y \
      linux-image-extra-virtual \
      apt-transport-https \
      ca-certificates \
      curl \
      software-properties-common && \
   curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
   add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) \
      stable edge" && \
   apt-get -y update

ARG DOCKER_VERSION=18.06.0~ce~3-0~ubuntu
RUN apt-get install -y docker-ce=${DOCKER_VERSION} unzip
