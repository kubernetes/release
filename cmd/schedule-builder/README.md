# Schedule Builder

This simple tool has the objective to parse the yaml file located in [SIG-Release](https://github.com/kubernetes/sig-release/blob/master/releases/), which shows the scheduled and past patch releases of the current Kubernetes Release cycle in machine readable format.

## Install

The simplest way to install the `schedule-builder` CLI is via `go get`:

```
$ go get k8s.io/release/cmd/schedule-builder
```

This will install `schedule-builder` to `$(go env GOPATH)/bin/schedule-builder`.

Also if you have the `kubernetes/release` cloned you can run the `make release-tools` to build all the tools.

## Usage

To run this tool you can just do, assuming you have cloned both `SIG-Release` and `Release` repositories, like

```
.
+-- kubernetes
|   +-- sig-release
|   +-- release
```

```bash
$ schedule-builder --config-path ../sig-release/releases/schedule.yaml
```

The output will be something similar to this

```
### Timeline

### 1.18

Next patch release is **1.18.4**

End of Life for **1.18** is **TBD**

| PATCH RELEASE | CHERRY PICK DEADLINE  | TARGET DATE |
|---------------|-----------------------|-------------|
| 1.18.4        | 2020-06-12            | 2020-06-17  |
| 1.18.3        | 2020-05-15            | 2020-05-20  |
| 1.18.2        | 2020-04-13            | 2020-04-16  |
| 1.18.1        | 2020-04-06            | 2020-04-08  |

### 1.17

Next patch release is **1.17.7**

End of Life for **1.17** is **TBD**

| PATCH RELEASE |                                 CHERRY PICK DEADLINE                                  | TARGET DATE |
|---------------|---------------------------------------------------------------------------------------|-------------|
| 1.17.7        | 2020-06-12                                                                            | 2020-06-17  |
| 1.17.6        | 2020-05-15                                                                            | 2020-05-20  |
| 1.17.5        | 2020-04-13                                                                            | 2020-04-16  |
| 1.17.4        | 2020-03-09                                                                            | 2020-03-12  |
| 1.17.3        | 2020-02-07                                                                            | 2020-02-11  |
| 1.17.2        | No-op release https://groups.google.com/d/topic/kubernetes-dev/Mhpx-loSBns/discussion | 2020-01-21  |
| 1.17.1        | 2020-01-10                                                                            | 2020-01-14  |

### 1.16

Next patch release is **1.16.11**

End of Life for **1.16** is **TBD**

| PATCH RELEASE |                                 CHERRY PICK DEADLINE                                  | TARGET DATE |
|---------------|---------------------------------------------------------------------------------------|-------------|
| 1.16.11       | 2020-06-12                                                                            | 2020-06-17  |
| 1.16.10       | 2020-05-15                                                                            | 2020-05-20  |
| 1.16.9        | 2020-04-13                                                                            | 2020-04-16  |
| 1.16.8        | 2020-03-09                                                                            | 2020-03-12  |
| 1.16.7        | 2020-02-07                                                                            | 2020-02-11  |
| 1.16.6        | No-op release https://groups.google.com/d/topic/kubernetes-dev/Mhpx-loSBns/discussion | 2020-01-21  |
| 1.16.5        | 2020-01-10                                                                            | 2020-01-14  |
| 1.16.4        | 2019-12-06                                                                            | 2019-12-11  |
| 1.16.3        | 2019-11-08                                                                            | 2019-11-13  |
| 1.16.2        | 2019-10-11                                                                            | 2019-10-15  |
| 1.16.1        | 2019-09-27                                                                            | 2019-10-02  |
```

Also can save the schedule in a file, to do that, you can set the `--output-file` flag together with the filename.

```
$ schedule-builder --config-path ../sig-release/releases/schedule.yaml --output-file my-schedule.md
```
