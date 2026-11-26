# krel signing build type

Build type URI: `https://git.k8s.io/release/docs/krel/buildtypes/v2`

This build type describes a Kubernetes stage build, as recorded by Cloud
Build and attested by the signing build of
[`gcb/provenance/cloudbuild.yaml`](/gcb/provenance/cloudbuild.yaml), which
runs `krel provenance` outside of the stage build. Unlike the
[v1](../v1/README.md) provenance, which `krel stage` writes and signs
itself, this provenance is written from the Cloud Build record of the stage
and from digests the signing build computes, and signed by an identity the
stage can't use, for SLSA build level 3. The stage keeps writing its v1
provenance next to it.

## Build definition

### External parameters

The externally controlled inputs of the stage build, from its Cloud Build
record:

| Field           | Description                                                              |
|-----------------|--------------------------------------------------------------------------|
| `releaseType`   | Type of release to build: `alpha`, `beta`, `rc` or `official` (`_TYPE`). |
| `releaseBranch` | Kubernetes branch the release is built from (`_RELEASE_BRANCH`).         |
| `buildVersion`  | Build version to release (`_BUILDVERSION`).                              |
| `nomock`        | Whether the stage staged to the production locations (`_NOMOCK`).        |
| `entryPoint`    | Cloud Build configuration of a stage.                                    |
| `substitutions` | The substitutions the stage build was submitted with, without `_GCP_USER_TAG`, which names who submitted it. |
| `steps`         | The steps of the stage build: their image `name`, `entrypoint` and `args`. |

### Internal parameters

| Field     | Description                              |
|-----------|------------------------------------------|
| `project` | Google Cloud project of the stage build. |

### Resolved dependencies

- Kubernetes at `_K8S_ORG` and `_K8S_REPO`: with the commit Cloud Build
  checked out when that repository is the source of the stage build, and
  otherwise at the `_K8S_REF` the stage checked out itself, with the commit
  only when `_K8S_REF` is one. Stages that clone Kubernetes in a step, as
  `gcb/stage` does today, record the ref, usually `HEAD`, and the commit
  only through `_BUILDVERSION`.
- kubernetes/release at the `_TOOL_ORG`, `_TOOL_REPO` and `_TOOL_REF` the
  stage ran krel from, with the commit when `_TOOL_REF` is one.
- The image of each step, by the SHA-256 digest Cloud Build resolved, once
  per image and digest.

## Run details

- `builder.id`: `https://git.k8s.io/release/gcb/provenance`
- `metadata.invocationId`: the Cloud Build build ID of the stage build.
- `metadata.startedOn` / `metadata.finishedOn`: start and end of the stage
  build.

## Attestations

The signing build writes three statements with this predicate, signed as
`krel-provenance@k8s-releng-prod.iam.gserviceaccount.com` for nomock stages,
or as the untrusted `krel-provenance-mock@k8s-releng-prod.iam.gserviceaccount.com`
for mock stages, into the stage directory in the bucket. Each signer is
pinned to its bucket and registry, the production ones or the mock ones, and
the signing build fails for a stage build whose substitutions name others:

- `provenance-v2.json` is about the staged files, the same subjects as
  `provenance.json` of the stage, with digests the signing build computed.
- `image-provenance-v2.json` is about the pushed container images, the same
  subjects as `image-provenance.json`. Its bundle is also attached to the
  images as an OCI referrer, like the v1 one.
- `<version>/final-provenance-v2.json` is about the files of the release of
  the version, named by their path in the `release` directory of the bucket.
  They are the files the release SBOM lists, and each has to be staged with
  the digest the SBOM claims.

The signature and certificate files `krel release` adds to the stage
directory later (`.sig`, `.cert`) are not subjects, like in `provenance.json`.

The signing build only signs stage builds the stage trigger of their kind
started, production or mock, that ran as
`krel-stage@k8s-release.iam.gserviceaccount.com`: Cloud Build announces every
build of the project, also ones submitted by hand.

Before it signs, the signing build checks that the staged files and images
are the ones the stage build wrote: every staged file has to have been
created and last updated while the stage build ran, from the metadata of its
object, listed after the download, and every image the tags of the stage
resolve to has to have been uploaded to its repository while the stage build
ran. A file replaced or an image tag moved after the stage build fails the
signing build. This doesn't cover writes by another build that runs at the
same time as the stage build.

Running the signing build again for the same stage build keeps a bundle that
verifies as the signer's and holds the same statement, and signs it again
otherwise.

## Example

```json
{
  "buildDefinition": {
    "buildType": "https://git.k8s.io/release/docs/krel/buildtypes/v2",
    "externalParameters": {
      "releaseType": "beta",
      "releaseBranch": "master",
      "buildVersion": "v1.38.0-beta.0.899+9f140ee0061e18",
      "nomock": true,
      "entryPoint": "https://git.k8s.io/release/gcb/stage/cloudbuild.yaml",
      "substitutions": {
        "_BUILDVERSION": "v1.38.0-beta.0.899+9f140ee0061e18",
        "_RELEASE_BRANCH": "master",
        "_TYPE": "beta"
      },
      "steps": [
        {
          "name": "gcr.io/k8s-staging-releng/k8s-cloud-builder:v1.38.0-go1.27.4-trixie.0",
          "entrypoint": "",
          "args": ["bin/krel", "stage", "--nomock"]
        }
      ]
    },
    "internalParameters": {
      "project": "k8s-release"
    },
    "resolvedDependencies": [
      {
        "uri": "git+https://github.com/kubernetes/kubernetes@HEAD"
      },
      {
        "uri": "git+https://github.com/kubernetes/release@9c79ae380875188f191f7d30ee01389772a10f3d",
        "digest": {"gitCommit": "9c79ae380875188f191f7d30ee01389772a10f3d"}
      },
      {
        "name": "gcr.io/k8s-staging-releng/k8s-cloud-builder:v1.38.0-go1.27.4-trixie.0",
        "digest": {"sha256": "8073326e165640d2ed25688ffd1b6eaffc4841c54e4664ed3ba551b6a236884d"}
      }
    ]
  },
  "runDetails": {
    "builder": {
      "id": "https://git.k8s.io/release/gcb/provenance"
    },
    "metadata": {
      "invocationId": "54ef86d9-5a89-4e0e-a008-10904a0f5a34",
      "startedOn": "2026-11-25T15:56:49.360329688Z",
      "finishedOn": "2026-11-25T16:26:50.339390Z"
    }
  }
}
```
