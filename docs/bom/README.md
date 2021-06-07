# bom (Bill of Materials)

Create SPDX compliant Bill of Materials 

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)

## Summary 

bom is a little utility that lets software authors generate
SPDX manifests to describe the contents of a release. The
SPDX manifests provide a way to list and verify all items
contained in packages, images, and individual files while
packing the data along with licensing information.

bom is still in its early stages and it is an effort to open
the libraries developed for the Kubernetes SBOM for other 
projects to use.

For a more in depth instructions on how to create an SBOM see
[Generating a Bill of Materials for Your Project](create-a-bill-of-materials.md)

## Installation

To use bom generate, compile the release engineering tools:

```
git clone git@github.com:kubernetes/release.git
cd release
./compile-release-tools bom
```

## Usage
```
  bom [subcommand]
```

### Available Commands
```
  generate    bom generate → Create SPDX manifests
  help        Help about any command
```

### Command line flags

```
Flags:
  -h, --help               help for bom
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
```
