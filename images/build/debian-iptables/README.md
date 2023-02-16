### debian-iptables

Serves as the base image for `registry.k8s.io/kube-proxy-${ARCH}` and multiarch (not `amd64`) `registry.k8s.io/flannel-${ARCH}` images.

This image is compiled for multiple architectures.

#### How to release

If you're editing the Dockerfile or some other thing, please bump the `TAG` in the Makefile.

```console
Build and  push images for all the architectures
$ make all-push
# ---> gcr.io/k8s-staging-build-image/debian-iptables-amd64:TAG
# ---> gcr.io/k8s-staging-build-image/debian-iptables-arm:TAG
# ---> gcr.io/k8s-staging-build-image/debian-iptables-arm64:TAG
# ---> gcr.io/k8s-staging-build-image/debian-iptables-ppc64le:TAG
# ---> gcr.io/k8s-staging-build-image/debian-iptables-s390x:TAG
```

If you don't want to push the images, run `make build ARCH={target_arch}` or `make all-build` instead


[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/build/debian-iptables/README.md?pixel)]()
