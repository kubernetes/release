# cip-mm

This tool **m**odifies promoter **m**anifests. For now it dumps some filtered
subset of a staging GCR and merges those contents back into a given promoter
manifest.

## Examples

- Add all images with a matching digest from staging repo
  `gcr.io/k8s-staging-artifact-promoter` to a manifest, using the name and tags
  already existing in the staging repo:

```
cip-mm \
  --base_dir=$HOME/go/src/github.com/kubernetes/k8s.io/k8s.gcr.io  \
  --staging_repo=gcr.io/k8s-staging-artifact-promoter \
  --filter_digest=sha256:7594278deaf6eeaa35caedec81796d103e3c83a26d7beab091a5d25a9ba6aa16
```

- Add a single image named "foo" and tagged "1.0" from staging repo
  `gcr.io/k8s-staging-artifact-promoter` to a manifest:

```
cip-mm \
  --base_dir=$HOME/go/src/github.com/kubernetes/k8s.io/k8s.gcr.io  \
  --staging_repo=gcr.io/k8s-staging-artifact-promoter \
  --filter_image=cip \
  --filter_tag=1.0
```

- Add all images tagged `1.0` from staging repo
  `gcr.io/k8s-staging-artifact-promoter` to a manifest:

```
cip-mm \
  --base_dir=$HOME/go/src/github.com/kubernetes/k8s.io/k8s.gcr.io  \
  --staging_repo=gcr.io/k8s-staging-artifact-promoter \
  --filter_image=cip \
  --filter_tag=1.0
```
