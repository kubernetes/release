# Release tooling images

By convention, every image we maintain has a subdirectory in this location. Each image directory contains a minimum of:

- a Dockerfile named `Dockerfile`
- a cloud builder config named `cloudbuild.yaml`

An image is built using [cloud-build](https://cloud.google.com/cloud-build) as follows:

```sh:
IMG='k8s-cloud-builder'
gcloud builds submit --config "./${IMG}/cloudbuild.yaml" "./${IMG}"
```

Alternatively, these images can also be built/updated using the [project Makefile](../Makefile):

```sh
# update all images
make update-images

# update a specific image, in this example the k8s-cloud-builder
# make image-<image-name>
make image-k8s-cloud-builder
```

The build/update process performs the following operations:

- uploads the image directory content to cloud-build
- uses it as a build context
- creates an image as per the Dockerfile

If the operations succeeds and if `cloudbuild.yaml` specifies the `images` target, images will
be pushed to the specified registries with the specified tags. You can get more details in the [cloud-build docs][gcb_images].

## Currently used images

| Image | used in/by |
| --- | --- |
| [k8s-cloud-builder](./k8s-cloud-builder/) | The "main" image, [`krel stage/release`](../cmd/krel) runs with on cloud-build |
| [releng-ci](./releng/ci) | The image used for CI testing |

[gcb_images]: https://cloud.google.com/cloud-build/docs/configuring-builds/store-images-artifacts#storing_images_in
