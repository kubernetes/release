# Copyright 2021 The Kubernetes Authors.
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

# An attempt to fix issues like:
# ```
# Error while loading /usr/sbin/dpkg-split: No such file or directory
# Error while loading /usr/sbin/dpkg-deb: No such file or directory
# ```
# See: https://github.com/docker/buildx/issues/495
RUN ln -s /usr/bin/dpkg-split /usr/sbin/dpkg-split \
    && ln -s /usr/bin/dpkg-deb /usr/sbin/dpkg-deb \
    && ln -s /bin/tar /usr/sbin/tar \
    && ln -s /bin/rm /usr/sbin/rm \
    && apt-get update \
    && apt-get -y --no-install-recommends install libcap2-bin
