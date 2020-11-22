This test has a specal case where the folder "manifest" has a promoter manifest
in it. This should succeed because we check for manifest files under the
toplevel "manifests" folder (with in "s"), and ignore other folders.
