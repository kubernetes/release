# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG BASEIMAGE

FROM ${BASEIMAGE}

# Install iptables and ebtables packages from buster-backports
RUN ln -s /bin/rm /usr/sbin/rm \
    && echo deb http://deb.debian.org/debian buster-backports main >> /etc/apt/sources.list \
    && apt-get update \
    && apt-get -t buster-backports -y --no-install-recommends install \
        iptables \
        ebtables

# Install other dependencies and then clean up apt caches
RUN clean-install \
    conntrack \
    ipset \
    kmod \
    netbase

# Install iptables wrapper scripts to detect the correct iptables mode
# the first time any of them is run
ADD https://raw.githubusercontent.com/kubernetes-sigs/iptables-wrappers/master/iptables-wrapper-installer.sh .
RUN chmod +x ./iptables-wrapper-installer.sh \
    && ./iptables-wrapper-installer.sh
