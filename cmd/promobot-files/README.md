# File Promoter

The file promoter (aka "promobot-files") copies files from source GCS buckets to
one or more destination buckets, by reading a Manifest file (in YAML).

The Manifest lists the files and their hashes that should be copied from src to
dest.

Example Manifest for files:

```
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
