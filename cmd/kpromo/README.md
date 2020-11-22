# kpromo - Artifact promoter

kpromo is a tool responsible for artifact promotion.

It has two operation modes:

- `run` - Execute a file promotion (formerly "promobot-files") (image promotion coming soon)
- `manifest` - Generate/modify a file manifest to target for promotion (image support coming soon)

Expectations:

- `kpromo run` should only be run in auditable environments
- `kpromo manifest` should primarily be run by contributors

- [File promotion](#file-promotion)
  - [Running the file promoter](#running-the-file-promoter)
  - [Generating a file promotion manifest](#generating-a-file-promotion-manifest)
  - [Consumers](#consumers)

## File promotion

The file promoter copies files from source GCS buckets to one or more
destination buckets, by reading a `Manifest` file (in YAML).

The `Manifest` lists the files and their hashes that should be copied from src to
dest.

Example `Manifest` for files:

```yaml
filestores:
- base: gs://staging/
  src: true
- base: gs://prod/subdir
  service-account: foo@google-containers.iam.gserviceaccount.com
files:
- name: vegetables/artichoke
  sha256: 2d4f26491e0e470236f73a0b8d6828db017eab988cd102fc19afe31f1f56aff7
- name: vegetables/beetroot
  sha256: 160b98e27ec99f77efe01e2996fa386f2b2aec552599f8bd861be0a857e7f29f
```

`filestores` is the equivalent of the container manifest `registries`, and lists
the buckets from which the promoter should read or write files.  `files` is the
equivalent of `images`, and lists the files that should be promoted.

`filestores` supports `service-account`, and it also supports relative paths -
note that the source files in the example above are in the root of the bucket,
but they are copied into a subdirectory of the target bucket.

`files` is a list of files to be copied.  The `name` is appended to the base of
the filestore, and then the files are copied.  If the source file does not have
the matching sha256, it will not be copied.

When errors are encountered building the list of files to be copied, no files
will be copied.  When errors are encountered while copying files, we will still
attempt to copy remaining files, but the process will report the error.

Currently only Google Cloud Storage (GCS) buckets supported, with a prefix of
`gs://`

### Running the file promoter

```console
$ kpromo run files --help

Promote files from a staging object store to production

Usage:
  kpromo run files [flags]

Flags:
      --dry-run                 test run promotion without modifying any filestore (default true)
      --files files             path to the files manifest
      --filestores filestores   path to the filestores promoter manifest
  -h, --help                    help for files
      --use-service-account     allow service account usage with gcloud calls

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
```

### Generating a file promotion manifest

This tool will generate a manifest fragment for uploading a set of
files, located in the specified path.

It takes a single argument `--src`, which is the base of the directory
tree; all files under that directory (recursively) are hashed and
output into the `files` section of a manifest.

The manifest is written to stdout.

```console
$ kpromo manifest files --help
Promote files from a staging object store to production

Usage:
  kpromo manifest files [flags]

Flags:
  -h, --help            help for files
      --prefix string   only export files starting with the provided prefix
      --src string      the base directory to copy from

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
```

### Consumers

- [`kOps`][kops-release-process]

[kops-release-process]: https://kops.sigs.k8s.io/development/release/
