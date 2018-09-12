# Kubernetes Snaps

## Build everything

Run `make` here and check for the results in the `build/` directory.

## Build one snap

To build a specific snap, run `make` with the name of the snap, e.g., for
kubectl:

```sh
$ make kubectl
```

The result will again be in the `build/` directory.

## Build a specific version

Set KUBE_VERSION to build the snaps with a particular Kubernetes version, e.g.,

```sh
$ make KUBE_VERSION=v1.5.5
```

## Cleaning up

Simply run `make clean` to remove everything except downloaded resources:

```sh
$ make clean
```
