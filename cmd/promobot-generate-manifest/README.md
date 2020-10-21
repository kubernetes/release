# promobot-generate-manifest

This tool will generate a manifest fragment for uploading a set of
files, located in the specified path.

It takes a single argument `--src`, which is the base of the directory
tree; all files under that directory (recursively) are hashed and
output into the `files` section of a manifest.

The manifest is written to stdout.
