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
ARG GO_VERSION
ARG OS_CODENAME
# TODO(codename): Consider parameterizing in Makefile based on codename
ARG DISTROLESS_IMAGE
FROM golang:${GO_VERSION}-${OS_CODENAME} as builder
WORKDIR /workspace

# Copy the sources
COPY ./go-runner.go ./
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
ARG package=.
ARG ARCH

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${ARCH}

RUN go env

RUN go build -ldflags '-s -w -buildid= -extldflags "-static"' \
    -o go-runner ${package}

# Production image
FROM gcr.io/distroless/${DISTROLESS_IMAGE}
LABEL maintainers="Kubernetes Authors"
LABEL description="go based runner for distroless scenarios"
WORKDIR /
COPY --from=builder /workspace/go-runner .
ENTRYPOINT ["/go-runner"]
