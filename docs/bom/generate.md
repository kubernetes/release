# bom generate

Create SPDX compliant Bill of Materials 

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)

## Summary 

`bom generate` is the subcommand to generate SPDX manifests.
Currently supports creating SBOM for files, images, and docker
archives (images in tarballs). Supports pulling images from
registries.

`bom` can take a deeper look into images using a growing number
of analyzers designed to add more sense to common base images.

## Installation

Simply [install bom](README.md).

##  Usage:

```
  bom generate [flags]
```

### Command Line Flags

```
Flags:
  -a, --analyze-images     go deeper into images using the available analyzers
  -c, --config string      path to yaml SBOM configuration file
  -d, --dirs strings       list of directories to include in the manifest as packages
  -f, --file strings       list of files to include
  -h, --help               help for generate
      --ignore strings     list of regexp patterns to ignore when scanning directories
  -i, --image strings      list of images
  -n, --namespace string   an URI that servers as namespace for the SPDX doc
      --no-gitignore       don't use exclusions from .gitignore files
      --no-gomod           don't perform go.mod analysis, sbom will not include data about go packages
      --no-transient       don't include transient go dependencies, only direct deps from go.mod
  -o, --output string      path to the file where the document will be written (defaults to STDOUT)
  -t, --tarball strings    list of docker archive tarballs to include in the manifest

```