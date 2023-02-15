### distroless-iptables

Serves as the base image for `registry.k8s.io/kube-proxy-${ARCH}`.

This image is compiled for multiple architectures.

#### How to release

If you're editing the Dockerfile or some other thing, please bump the `TAG` in the Makefile.

```console
Build and  push images for all the architectures
$ make all-push
# ---> gcr.io/k8s-staging-build-image/distroless-iptables-amd64:TAG
# ---> gcr.io/k8s-staging-build-image/distroless-iptables-arm:TAG
# ---> gcr.io/k8s-staging-build-image/distroless-iptables-arm64:TAG
# ---> gcr.io/k8s-staging-build-image/distroless-iptables-ppc64le:TAG
# ---> gcr.io/k8s-staging-build-image/distroless-iptables-s390x:TAG
```

If you don't want to push the images, run `make build ARCH={target_arch}` or `make all-build` instead.
