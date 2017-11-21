# Release Notes Collector

This is a Golang implementation of existing release note collector
[relnotes](https://github.com/kubernetes/release/blob/master/relnotes).

Golang requires this repo being placed within $GOPATH, explicitly at
$GOPATH/src/k8s.io/release.

This tool also uses [dep](https://github.com/golang/dep) to manage Golang
dependencies. To install dep, follow the instructions in [dep's Github
page](https://github.com/golang/dep).

**To bulid, run the build script:**

`./build.sh`

**Or do it manually:**

`cd $GOPATH/src/k8s.io/release`

`dep ensure`

`bazel run //:gazelle`

`bazel build toolbox/relnotes:relnotes`

**Some example command gathering release notes for Kubernetes (assume currently in
a kubernetes repo):**

* (On branch release-1.7:)

`../release/bazel-bin/toolbox/relnotes/relnotes --preview --htmlize-md
--html-file /tmp/release-note-html-testfile
--release-tars=_output/release-tars v1.7.0..v1.7.2`

* (On branch release-1.7:)

`../release/bazel-bin/toolbox/relnotes/relnotes --preview --html-file
/tmp/release-note-html-testfile --release-tars=_output/release-tars
v1.7.0..v1.7.0`

* (On branch release-1.6.3:)

`../release/bazel-bin/toolbox/relnotes/relnotes --html-file
/tmp/release-note-html-testfile --full`
