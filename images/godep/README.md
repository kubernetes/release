# Godep

This build step provides a container based on the golang container that adds in
the [Godep](https://github.com/tools/godep) tool.  Godep is an official golang
tool for dependency management.  Although godep is now deprecated, it is used
heavily throughout the Kubernetes project.  This image exists to be compatible
with existing Make targets that depend on godep.

This image uses godep as the entrypoint and takes in go commands as arguments.
An example use case for Google Container Builder:

```
- name: gcr.io/k8s-image-staging/godep
  dir: "/workspace/src/github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter"
  args: ["go", "test", "./..."]
```

The above example wraps the go test command in godep, which fetches packages
defined in the Godeps directory to satisfy dependencies.
