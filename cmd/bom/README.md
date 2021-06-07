# bom (Bill of Materials)
## A utility to generate SPDX compliant Bill of Materials manifests.

`bom` is a tiny utility that leverages the code written for the Kubernetes
Bill of Materials project. It enables software authors to generate an 
SBOM for their projects in a simple, yet powerful way.

![terminal demo](../../docs/bom/cast.svg "Terminal demo")


`bom` is a general-purpose tool that can generate SPDX packages from
directories, container images, single files, and other sources. The utility
has a built-in license classifier that recognizes the 400+ licenses in 
the SPDX catalog.

Other features include Golang dependency analysis and full `.gitignore`
support when scanning git repositories.

## Generate your own Bill of Materials

If you are looking for a way to create a bill of materials for your project, we
have created a 
[HOWTO guide to generating an SBOM](../../docs/bom/create-a-bill-of-materials.md).

The guide includes information about
[what a Bill of Materials is](../../docs/bom/create-a-bill-of-materials.md#what-is-a-bill-of-materials),
[the SPDX standard](../../docs/bom/create-a-bill-of-materials.md#spdx-software-package-data-exchange),
and instructions to add files, images, directories, and
other sources to your BOM.

## Compiling bom

To compile bom, clone the Kubernetes Release Engineering repository and
run the `compile-tools` script:

```
git clone git@github.com:kubernetes/release.git
cd release
./compile-release-tools
```

## Examples

The following examples show how bom can process different sources to generate
an SPDX Bill of Materials. Multiple sources can be combined to get a document
describing different packages.

### Generate an SBOM from the Current Directory:

To process a directory as a source for your SBOM, use the `-d` flag or simply pass
the path as the first argument to `bom`:

```bash
bom generate -n http://example.com/ .
```

### Process a Container Image

This example pulls the kube-apiserver image, analyzes it, and describes in the
SBOM. Each of its layers are then expressed as a subpackage in the resulting
document:

```
bom generate -n http://example.com/ --image k8s.gcr.io/kube-apiserver:v1.21.0 
```

### Generate a BOM to describe files

You can create an SBOM with just files in the manifest. For that, use `-f`:

```
bom generate -n http://example.com/ \
    -f Makefile \
    -f file1.exe \
    -f document.md \
    -f other/file.txt 
```