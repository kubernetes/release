### distroless-iptables

Serves as the base image for `k8s.gcr.io/kube-proxy-${ARCH}`.

This image is compiled for multiple architectures.

#### How to release

If you're editing the Dockerfile or some other thing, please bump the `TAG` in the Makefile.

```console
Build and  push images for all the architectures
$ make all-push
# ---> staging-k8s.gcr.io/distroless-iptables-amd64:TAG
# ---> staging-k8s.gcr.io/distroless-iptables-arm:TAG
# ---> staging-k8s.gcr.io/distroless-iptables-arm64:TAG
# ---> staging-k8s.gcr.io/distroless-iptables-ppc64le:TAG
# ---> staging-k8s.gcr.io/distroless-iptables-s390x:TAG
```

If you don't want to push the images, run `make build ARCH={target_arch}` or `make all-build` instead.
