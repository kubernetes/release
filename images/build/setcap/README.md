### setcap

This image is based on debian-base and installs the libcap2-bin package. The
main use of this image is to apply `CAP_NET_BIND_SERVICE` to the kube-apiserver
binary so that it can a bind to ports less than 1024 and still be run as non
root.

This image is compiled for multiple architectures.

### How to release

If you're editing the Dockerfile or some other thing, please bump the TAG in the Makefile.

Build and push images for all the architectures:

```console
$ make all-push
# ---> staging-k8s.gcr.io/setcap-amd64:TAG
# ---> staging-k8s.gcr.io/setcap-arm:TAG
# ---> staging-k8s.gcr.io/setcap-arm64:TAG
# ---> staging-k8s.gcr.io/setcap-ppc64le:TAG
# ---> staging-k8s.gcr.io/setcap-s390x:TAG
```

If you don't want to push the images, run `make sub-build-{target_arch}` or `make all-build` instead
