# Generating a Bill of Materials for Your Project

 <!-- toc -->
- [Introduction](#introduction)
  - [What is a Bill of Materials?](#what-is-a-bill-of-materials)
  - [bom: A tool to generate SBOMs for your Project](#bom-a-tool-to-generate-sboms-for-your-project)
- [SPDX: Software Package Data Exchange](#spdx-software-package-data-exchange)
  - [Files and Packages](#files-and-packages)
  - [Relationships](#relationships)
  - [Licensing Information](#licensing-information)
- [Create your Bill of Materials](#create-your-bill-of-materials)
  - [Thinking about your Release](#thinking-about-your-release)
  - [Generating the SBOM](#generating-the-sbom)
    - [Namespace](#namespace)
    - [Simplest Use Case: One Package](#simplest-use-case-one-package)
    - [Adding Additional Sources](#adding-additional-sources)
    - [Example: Generate an SBOM for etcd](#example-generate-an-sbom-for-etcd)
<!-- /toc -->

## Introduction

To generate a Bill of Materials for your project, download `bom`, our utility 
that leverages the code found in this repo and point it to your project:

```
cd gitrepo/
bom -n 'http://mybom.com/' . 
```

All of the tool's options are explained on its page. Keep reading
for more information about our tools, SBOMs, and the SPDX standard. 

### What is a Bill of Materials?

A _Software Bill of Materials_ (often BOM or SBOM for short) is a
manifest that lists everything included in a software release.
"Everything" can take different meanings: software packages or images,
documentation, tarballs, single files. These pieces can be components
or source code, variants of the same artifact (eg a binary for different
platforms).

An SBOM can also provide visibility on the dependencies of your project.
There are many types of dependencies and many reasons that consumers of
a project need to know them: security, compliance, compatibility.

Finally, a Bill of Materials allows software creators to express licensing
information for their project as a whole, but also for individual pieces
and it's dependencies. You can release your project under the Apache 2.0
license but have its documentation published under Creative Commons. Then,
there are all of your dependencies' original licenses. A well-written
SBOM can express all of them in the same document. 

### bom: A tool to generate SBOMs for your Project

As part of the effort to produce a bill of materials for Kubernetes, SIG
Release developed a set of libraries to produce fully compliant SPDX SBOMs.
Our tools support license scanning, image layer analyzers, processing of
golang dependencies, and other features. These libraries are available for
other projects to automate the production of their own Bills of Materials.

For simpler use cases, all of our SBOM automation is also available in
a general-purpose tool called `bom`. You can find all the options that
`bom` supports in its README.md.

`bom` supports generating Bills of Materials in SPDX compliant tag-value
format. It can process directories, single files, read the contents of
container images (both from container tar archives and directly from registries),
and tarred sources.

In addition `bom` will scan your code to find licensing information. Its
classifier supports detection of all the SPDX recognized licenses.

## SPDX: Software Package Data Exchange

[SPDX](https://spdx.dev/) or _Software Package Data Exchange_ is an open standard to create
bills of materials. It has been in the works for 10+ years, coordinated
by the SPDX Workgroup, a project of the Linux Foundation. 

As of June 2020, [the SPDX specification](https://spdx.github.io/spdx-spec/) is
in version 2.2. The current version allows software authors to include metadata about
their project describing its contents, relationships among them, and other
components and licensing. 

### Files and Packages

There are two main building blocks in an SPDX manifest: Files and Packages.

[Files](https://spdx.github.io/spdx-spec/4-file-information/) are what you
would expect: an individual item in a filesystem tree. The data about a file
in an SPDX SBOM includes its name, checksums, license, file type, copyright
data, and other attributes.

[Packages](https://spdx.github.io/spdx-spec/3-package-information/) are a
non-specific element in SPDX representing anything that can group other elements.
An `.rpm` or `.deb` package can be an SPDX package, but so can be a container
image or a tarball. Packages contain files, but can also contain other files
or a mix of both. An image, for example, can be viewed as a package, which
contains other packages (its layers), and those, in turn, a set of files.

SPDX metadata about packages is similar to metadata about files but it also
includes data about its version, where it came from, who wrote it, and an
important one: the package verification code.

To provide a mechanism to ensure the integrity of its contents, the
package construct defines a checksum
[verification code](https://spdx.github.io/spdx-spec/3-package-information/#39-package-verification-code).
This is a SHA1 sum derived from concatenating a hash of each item in the package.

### Relationships

The most useful feature of SPDX is the ability to express relationships among
elements. For example, a `Package` `CONTAINS` a `File`. A SPDX `Document`
`DESCRIBES` `Package`s and `File`s. A `File` is `GENERATED_FROM` a source
`Package` and so on.

The SPDX spec defines a [rich relationship vocabulary](https://spdx.github.io/spdx-spec/7-relationships-between-SPDX-elements/)
which enables developers to describe very complex interaction among components,
source code, the artifacts its build produces but also its dependencies and
build tools.

### Licensing Information

Independently of source code and artifacts, software licensing is a complex
problem in itself. SPDX has been thought from the ground up to express licensing
of each element in the document.

The pertinence of a license over a file or package can come from different
sources: it can be expressed by the file itself, it can be inherited from a package,
enforced by its dependencies, or perhaps it can be inferred by an automated tool.
SPDX makes no attempt to make any determination about the licensing of elements
but it has many different ways to allow authors to express a license and where the
licensing determination came from.

SPDX maintains a large list of open source licenses. All licenses have a tag that
represents them in a document. For example, the tag for the Mozilla Public License
2.0 is `MPL-2.0`, `MIT-0` is MIT without attribution. The SPDX licenses are published
in a public repository and are available in machine-readable formats such as JSON and XML.

## Create your Bill of Materials

There are a couple of factors to take into consideration whe drafting your Bill
of Materials.

### Thinking about your Release

The first thing you need to consider when planning your Bill of Materials is your release
structure. What does it look like? What are the main artifacts I want to list in my BOM?
Is your source code available?

But the main focus should be the consumers of the BOM. How is your document going to be
used? Is it for checking the completeness of your release? Is someone trying to check
for vulnerabilities in your dependencies? Perhaps your compliance person needs to check
the licenses that interact with your project. Think about these and other use cases and
create one or more SPDX documents which are useful for your consumers. 

As the name implies, open source software releases include a snapshot of the source code
in time: the state of your repo when a git tag was cut, for example. Do you want to include
the source code in the same document?When we were testing for the Kubernetes SBOM, the file
produced was over 11 MB long, so we decided to split the source data to its own SPDX file.

### Generating the SBOM

When you are ready to generate the BOM for your project, make sure everything you want to
list in the SPDX document is available. `bom` can read container images remotely from their
registry but everything else has to be available locally.

#### Namespace

Every SPDX document has to declare its namespace. The namespace is a URI, it must be unique
for the document you are generating. The purpose of the namespace is to have an anchor point
to reference your release in the SPDX world. Other software components which rely on your
project will use the URI to declare they are using your thing.

#### Simplest Use Case: One Package

In the simplest case, you can feed `bom` a source and build a single package SBOM.
For example, to generate SBOM from your git repository run the following (note the
dot at the end):

```
bom generate -n http://example.com/ .
```
This command will traverse your repository directory structure, listing everything it finds,
scanning license files. If your repository is a Go module, it will process the dependencies.
`bom` will use your `.gitignore` file and skip any patterns listed in it.

After bom runs, all your source code will be expressed as `File`s in an SPDX `Package`. `bom`
will do some determinations to complete the data it needs to produce the document such as
generating names for packages and files.
 
#### Adding Additional Sources

Generally, an SPDX bill of materials will include more than one package. you can pass `bom`
more sources to add to the document. These can be container images, other directories, container
archives, etc. When you add other sources, bom will add them as top-level packages in the
document. Some of these will include sub-packages: layers of images, dependencies, etc.

Here is a sample of other command line flags you can pass to `bom generate` to add more elements
to your bill of materials:

| Short | Long Flag | Description |
| --- | --- | --- |
| -d | --dirs | List of directories to include in the manifest as packages |
| -f | --file | List of files to include |
| -i | --image | List of image references |
| -t | --tarball | List of container archive tarballs to include in the manifest

#### Example: Generate an SBOM for etcd

Let us say you want to generate a bill of materials for etcd, which is at version v3.4.16
as I write this. If you only want to build an SBOM describing only the source in the
repository, do the following:

```
git clone https://github.com/etcd-io/etcd
cd etcd
bom generate -n https://etcd.io/etcd-v3.4.16.spdx -o etcd-v3.4.16.spdx \
   --directory=.
```

This will produce a manifest describing the repo and its golang dependencied 
in `etcd-v3.4.16.spdx`.

Now, to make your SBOM more complete, you may want to include a container image.
To do that run the same invocation, but this time adding the image with the
`--image` flag:

```
bom generate -n https://etcd.io/etcd-v3.4.16.spdx -o etcd-v3.4.16.spdx \
   --directory=.\
   --image=quay.io/coreos/etcd:v3.4.16
```
This command will fetch the container image from the coreos repo and add it as a
package. At this point, your bom will contain two top-level Packages: the directory
and the image. If you inspect it, you will see the image's layers as subpackages too.

Finally, perhaps you want to add a binary distribution file. Download the compressed
artifact from Github and add it to the SBOM:

```
curl -L https://github.com/etcd-io/etcd/releases/download/v3.4.16/etcd-v3.4.16-darwin-amd64.zip \
     -O /tmp/etcd-v3.4.16-darwin-amd64.zip

bom generate -n https://etcd.io/etcd-v3.4.16.spdx -o etcd-v3.4.16.spdx \
    --dirs=.  \
    --image=quay.io/coreos/etcd:v3.4.16 \
    --file=/tmp/etcd-v3.4.16-darwin-amd64.zip
``` 

The resulting sbom from the last invocation will include at the top level of the SBOM
three things: two `Package`s (the directory and the image) and one `File`: the binary
distribution. Note that when listing the zip file as a single file, `bom` did not perform
any special treatment to it. You can see a copy of the resulting file: `etcd-v3.4.16.spdx`