# Copyright 2020 The Kubernetes Authors.
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

# Build the manager binary
ARG DISTROLESS_IMAGE
ARG BUILDER_IMAGE
FROM ${BUILDER_IMAGE} as builder
WORKDIR /workspace

# Copy the module files which define the dependency for
# the external kube-log-runner binary.
COPY ./go.* ./

# Allow fallback to 'direct' for GOPROXY
#
# The GOPROXY environment variable now supports skipping proxies that return
# errors. Proxy URLs may now be separated with either commas (,) or pipe
# characters (|). If a proxy URL is followed by a comma, the go command will
# only try the next proxy in the list after a 404 or 410 HTTP response. If a
# proxy URL is followed by a pipe character, the go command will try the next
# proxy in the list after any error. Note that the default value of GOPROXY
# remains https://proxy.golang.org,direct, which does not fall back to direct
# in case of errors.
#
# ref: https://golang.org/doc/go1.15#go-command
ENV GOPROXY="https://proxy.golang.org|direct"

# Build
ARG package=k8s.io/component-base/logs/kube-log-runner
ARG ARCH

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${ARCH}

RUN go env
RUN go version

RUN GOBIN=`pwd` go install -ldflags '-s -w -buildid= -extldflags "-static"' \
    ${package}

# Production image
FROM ${DISTROLESS_IMAGE}
LABEL maintainers="Kubernetes Authors"
LABEL description="go based runner for distroless scenarios"
WORKDIR /

# The image uses the traditional "go-runner" name because that is part of its
# API.
COPY --from=builder /workspace/kube-log-runner go-runner
ENTRYPOINT ["/go-runner"]
