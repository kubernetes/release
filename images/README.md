# Images for the release tooling

By convention, every image we maintain, has a subdirectory here. This directory as a minimum contains
- a Dockerfile named `Dockerfile`
- a cloud builder config named `cloudbuild.yaml`

An image can therefore be built via cloud-build with
```sh
IMG='k8s-cloud-builder'
gcloud builds submit --config "./${IMG}/cloudbuild.yaml" "./${IMG}"
```

Alternatively those images can also be built/updated via the [top level Makefile](../Makefile):
```sh
# update all images
make update-images

# update a specific image, in this example the k8s-cloud-builder
# make image-<image-name>
make image-k8s-cloud-builder
```

By doing so, the image directorie's content will uploaded to cloud-build & used as a
build context and an image according to the Dockerfile will be created. If that
succeeds and the `cloudbuild.yaml` specifies `images` targets, this image will
be pushed to the specified registries with the specified tags. You can read a
bit more on that [in the cloud-build docs][gcb_images].


## Currently used images

Image                                     | used in/by
:---:                                     | --
[k8s-cloud-builder](./k8s-cloud-builder/) | The "main" image, [`anago`](../anago) runs with on cloud-build (submitted via [`gcbmgr`](../gcbmgr))

[gcb_images]: https://cloud.google.com/cloud-build/docs/configuring-builds/store-images-artifacts#storing_images_in
