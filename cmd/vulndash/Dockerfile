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
ARG DISTROLESS_IMAGE
FROM golang:${GO_VERSION} as builder

WORKDIR /go/src/k8s.io/release

# Copy the sources
ENV package="./cmd/vulndash"
COPY ../../go.mod ../../go.sum ./
COPY ../../pkg ./pkg/
COPY ../../cmd/vulndash ${package}/

RUN go mod download

# Build
ARG ARCH

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${ARCH}

RUN go build -ldflags '-s -w -buildid= -extldflags "-static"' \
    -o vulndash ${package}

# Production image
FROM gcr.io/distroless/${DISTROLESS_IMAGE}:latest

LABEL maintainers="Kubernetes Authors"
LABEL description="A vulnerability dashboard for GCR container images"

WORKDIR /
COPY --from=builder /go/src/k8s.io/release/vulndash .

ENTRYPOINT ["/vulndash"]
