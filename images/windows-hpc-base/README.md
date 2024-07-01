# Windows HostProcessContainer Base Image

This image can be used as the base image for Windows HostProcessContainer images and is based on https://github.com/microsoft/windows-host-process-containers-base-image/.

This image is built in and published to resources owned by the Kubernetes community so that other Kubernetes images can build upon it ant not rely on external resources such as the Microsoft Container Registry.

To build the image tarball:

```sh
make image_build
```

To publish the image to a registry:

```sh
REGISTRY=gcr.io/k8s-staging-releng VERSION=v0.1.0 make image_publish
```
