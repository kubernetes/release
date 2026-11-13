### distroless-iptables

Serves as the base image for `registry.k8s.io/kube-proxy-${ARCH}`.

This image is compiled for multiple architectures.

The distroless-trixie build is used for kubernetes 1.38 and later, and
includes only iptables, ipset, and nft.

The distroless-bookworm version is used for older kubernetes releases,
and additionally contains the iptables-wrapper script, plus various
other binaries that iptables-wrapper needs. (Kubernetes 1.38 and later
don't need iptables-wrapper because their golang iptables code handles
the iptables-legacy/iptables-nft detection itself.)

We build our own nft 1.0.6.1 binary rather than using the older
version in bookworm or the newer version in trixie, because 1.0.6.1
has the best compatibility with both newer and older nft binaries that
might be in use by the rest of the system.

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
