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

ARG GO_VERSION
ARG DISTROLESS_IMAGE
ARG BUILDER_IMAGE

# Get Go from official Alpine image (use native platform to avoid emulation)
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine as go-source

# Build the manager binary (use native platform for build, cross-compile to target)
FROM --platform=$BUILDPLATFORM ${BUILDER_IMAGE} as builder
WORKDIR /workspace

# Copy Go installation from official image
COPY --from=go-source /usr/local/go /usr/local/go

# Set up Go environment
ENV PATH=/usr/local/go/bin:$PATH
ENV GOROOT=/usr/local/go
ENV GOLANG_VERSION=${GO_VERSION}
ENV GOTOOLCHAIN=auto
ENV GOPATH=/workspace

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

ENV CGO_ENABLED=0

# Determine target architecture from build platform
ARG TARGETARCH
ARG TARGETOS
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN go env

# Use cross-compilation to avoid QEMU emulation issues with Go 1.26.0 runtime
RUN go build -ldflags '-s -w -buildid= -extldflags "-static"' \
    -o go-runner ${package}

# Production image
FROM ${DISTROLESS_IMAGE}
LABEL maintainers="Kubernetes Authors"
LABEL description="go based runner for distroless scenarios"
WORKDIR /
COPY --from=builder /workspace/go-runner .
ENTRYPOINT ["/go-runner"]
