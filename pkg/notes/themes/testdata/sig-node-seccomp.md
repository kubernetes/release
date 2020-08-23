---
title: seccomp graduates to General Availability
sig: node
---
The seccomp (secure computing mode) support for Kubernetes has graduated to General Availability (GA). This feature can be used to increase the workload security by restricting the system calls for a Pod (applies to all containers) or single containers.

Technically this means that a first class `seccompProfile` field has been added to the Pod and Container `securityContext` objects:

```yaml
securityContext:
seccompProfile:
    type: RuntimeDefault|Localhost|Unconfined # choose one of the three
    localhostProfile: my-profiles/profile-allow.json # only necessary if type == Localhost
```

The support for `seccomp.security.alpha.kubernetes.io/pod` and `container.seccomp.security.alpha.kubernetes.io/...` annotations are now deprecated, and will be removed in Kubernetes v1.22.0. Right now, an automatic version skew handling will convert the new field into the annotations and vice versa. This means there is no action required for converting existing workloads in a cluster.

You can find more information about how to restrict container system calls with seccomp in the new [documentation page on Kubernetes.io][seccomp-docs]

[seccomp-docs]: https://kubernetes.io/docs/tutorials/clusters/seccomp/
