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

# Install dependencies for building Go from source
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and build Go from source
ARG GO_VERSION

ARG TARGETARCH
RUN echo "Determining bootstrap Go version for ${GO_VERSION}" && \
    BOOTSTRAP_GO_VERSION=$(curl -fsSL "https://raw.githubusercontent.com/golang/go/go${GO_VERSION}/src/make.bash" | grep "^bootgo=" | cut -d= -f2) && \
    echo "Installing bootstrap Go version: ${BOOTSTRAP_GO_VERSION} (${TARGETARCH})" && \
    GO_ARCH="${TARGETARCH}"; \
    if [ "${TARGETARCH}" = "arm" ]; then GO_ARCH="armv6l"; fi && \
    curl -fsSL "https://go.dev/dl/go${BOOTSTRAP_GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go-bootstrap.tar.gz && \
    tar -xzf /tmp/go-bootstrap.tar.gz -C /usr/local && \
    mv /usr/local/go /usr/local/go-bootstrap && \
    rm /tmp/go-bootstrap.tar.gz

# Download Go source
RUN echo "Downloading Go source version: ${GO_VERSION}" && \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.src.tar.gz" -o /tmp/go.src.tar.gz && \
    tar -xzf /tmp/go.src.tar.gz -C /usr/local && \
    rm /tmp/go.src.tar.gz

# Build Go from source with bootstrap
WORKDIR /usr/local/go/src
ENV GOROOT_BOOTSTRAP=/usr/local/go-bootstrap
RUN ./make.bash

# Set up Go environment
ENV PATH=/usr/local/go/bin:$PATH
ENV GOROOT=/usr/local/go
ENV GOPATH=/workspace/gopath

# Copy the sources
WORKDIR /workspace
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

RUN go env

RUN go build -ldflags '-s -w -buildid= -extldflags "-static"' \
    -o go-runner ${package}

# Production image
FROM ${DISTROLESS_IMAGE}
LABEL maintainers="Kubernetes Authors"
LABEL description="go based runner for distroless scenarios"
WORKDIR /
COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /workspace/go-runner .
ENTRYPOINT ["/go-runner"]
