This test has a specal case where the folder "c" has a promoter manifest in it,
and also has a folder at "../../images/c/images.yaml" (relative to its
location). This should still fail because we check for images files under the
toplevel "images" folder, not just any folder 2 levels up from the promoter
manifest.
