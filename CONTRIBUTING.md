# Contributing

Thanks for taking the time to join our community and start contributing!

The [Contributor Guide](https://github.com/kubernetes/community/blob/master/contributors/guide/README.md)
provides detailed instructions on how to get your ideas and bug fixes seen and accepted.

Please remember to sign the [CNCF CLA](https://github.com/kubernetes/community/blob/master/CLA.md) and
read and observe the [Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Autogenerating Bazel Configs

For Go code, this repository is currently set up with [Gazelle](https://github.com/bazelbuild/bazel-gazelle), which is a tool that can be used to generate Bazel `BUILD` files.

If you add Go code which includes new dependencies, you have to update the [Dep](https://github.com/golang/dep) configs and then use Gazelle to generate the appropriate Bazel configs:

```bash
# install dep
go get -u github.com/golang/dep/cmd/dep

# update Gopkg.lock
dep ensure

# generate the go_repository stanzas in WORKSPACE
bazel run //:gazelle -- update-repos -from_file=Gopkg.lock

# generate all of the BUILD files
bazel run //:gazelle
```

If you add new Go files but do not add any dependencies, the following should be sufficient:

```bash
# generate all of the BUILD files
bazel run //:gazelle
```
