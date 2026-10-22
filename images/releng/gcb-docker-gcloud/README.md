# gcb-docker-gcloud image

This image is available for Google Cloud Build jobs that want to use a
combination of `docker`, `gcloud`, and `go` all in the same build step

## contents

- base:
  - golang:alpine
- languages:
  - `go`
- tools:
  - `docker`
  - `docker-buildx` and `tonistiigi/binfmt` for multi-arch support
  - `gcloud` via rapid channel, with default components

You can find the image at `registry.k8s.io/releng/gcb-docker-gcloud`.
