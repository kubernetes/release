#!/bin/bash

# This is a build script for Golang relnotes (release note collector), see
# README.md for more information.

# At the root directory
cd $GOPATH/src/k8s.io/release

# Run dep to get the dependencies
dep ensure

# Run gazelle to auto-generate BUILD files
bazel run //:gazelle

# Build the program
bazel build toolbox/relnotes:relnotes
