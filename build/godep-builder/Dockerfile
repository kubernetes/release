# Copyright 2019 The Kubernetes Authors.
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

FROM golang:1.12

LABEL "source.repo"="github/kubernetes/release/build/godep-builder"

ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH

RUN go version && \
    go get github.com/tools/godep && \
    godep version && \
    apt-get update && \
    apt-get install -y build-essential

ENTRYPOINT ["/gopath/bin/godep"]
