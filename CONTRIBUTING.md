# Contributing

Thanks for taking the time to join our community and start contributing!

The [Contributor Guide](https://github.com/kubernetes/community/blob/master/contributors/guide/README.md)
provides detailed instructions on how to get your ideas and bug fixes seen and accepted.

Please remember to sign the [CNCF CLA](https://github.com/kubernetes/community/blob/master/CLA.md) and
read and observe the [Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Autogenerating Bazel Configs

If you add or modify Go code, make sure to generate the necessary Bazel (`BUILD.bazel`) files:

```bash
./hack/update-all.sh
```

## Submitting a Pull Request

Before submitting a pull request, please make sure to verify that all tests are passing:

```bash
./hack/verify-all.sh
```
