### distroless-iptables

Serves as the base image for `registry.k8s.io/kube-proxy-${ARCH}`.

This image is compiled for multiple architectures.

#### How to release

If you're editing the Dockerfile or some other thing, please bump the `IMAGE_VERSION` in the Makefile.

```console
Build and push images for all the architectures, then push the manifest list
$ make manifest REGISTRY=us-central1-docker.pkg.dev/k8s-staging-images/build-image
# ---> us-central1-docker.pkg.dev/k8s-staging-images/build-image/distroless-iptables-amd64:IMAGE_VERSION
# ---> us-central1-docker.pkg.dev/k8s-staging-images/build-image/distroless-iptables-arm64:IMAGE_VERSION
# ---> us-central1-docker.pkg.dev/k8s-staging-images/build-image/distroless-iptables-ppc64le:IMAGE_VERSION
# ---> us-central1-docker.pkg.dev/k8s-staging-images/build-image/distroless-iptables-s390x:IMAGE_VERSION
# ---> us-central1-docker.pkg.dev/k8s-staging-images/build-image/distroless-iptables:IMAGE_VERSION
```

If you don't want to push the images, run `make container REGISTRY=... [PLATFORMS=linux/{target_arch}]` instead.
