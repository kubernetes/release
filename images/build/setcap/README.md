### setcap

This image is based on debian-base and installs the libcap2-bin package. The
main use of this image is to apply `CAP_NET_BIND_SERVICE` to the kube-apiserver
binary so that it can a bind to ports less than 1024 and still be run as non
root.

This image is compiled for multiple architectures.

### How to release

If you're editing the Dockerfile or some other thing, amend the variants.yaml and bump the versions in dependencies.yaml.

Build and push images for all the architectures, then push the manifest list:

```console
$ make manifest REGISTRY=<registry>
# ---> <registry>/setcap-amd64:TAG
# ---> <registry>/setcap-arm64:TAG
# ---> <registry>/setcap-ppc64le:TAG
# ---> <registry>/setcap-s390x:TAG
# ---> <registry>/setcap:TAG
```

If you don't want to push the images, run `make container REGISTRY=<registry>` instead. Use `PLATFORMS=linux/<arch>` to limit the platforms built.
