# Kubernetes RPM Builder

This directory contains a spec file for Kubernetes, and a Dockerfile for building the RPMs without requiring RPM specific tooling on the host system.

cd to this directory and run ./build-docker.sh.

A Docker container will be built, and run to build the packages and generate yum repo metadata. Output will be in output/x86_64/ directory.

RPMs are built by downloading upstream published binaries rather than from source.
