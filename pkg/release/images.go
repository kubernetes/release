/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package release

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/command"
)

// Images is a wrapper around container image related functionality
type Images struct {
	client commandClient
}

// NewImages creates a new Images instance
func NewImages() *Images {
	return &Images{&defaultCommandClient{}}
}

// SetClient can be used to set the internal command client
func (i *Images) SetClient(client commandClient) {
	i.client = client
}

// commandClient is a client for working with Docker
//counterfeiter:generate . commandClient
type commandClient interface {
	Execute(cmd string, args ...string) error
	RepoTagFromTarball(path string) (string, error)
}

type defaultCommandClient struct{}

func (*defaultCommandClient) Execute(cmd string, args ...string) error {
	return command.Execute(cmd, args...)
}

func (*defaultCommandClient) RepoTagFromTarball(path string) (string, error) {
	tagOutput, err := command.
		New("tar", "xf", path, "manifest.json", "-O").
		Pipe("jq", "-r", ".[0].RepoTags[0]").
		RunSilentSuccessOutput()
	if err != nil {
		return "", err
	}
	return tagOutput.Output(), nil
}

var tagRegex = regexp.MustCompile(`^.+/(.+):.+$`)

// PublishImages relases container images to the provided target registry
// was in releaselib.sh: release::docker::release
func (i *Images) Publish(registry, version, buildPath string) error {
	releaseImagesPath := filepath.Join(buildPath, ImagesPath)
	logrus.Infof(
		"Pushing container images from %s to registry %s",
		releaseImagesPath, registry,
	)

	manifestImages := make(map[string][]string)

	archPaths, err := ioutil.ReadDir(releaseImagesPath)
	if err != nil {
		return errors.Wrapf(err, "read images path %s", releaseImagesPath)
	}
	for _, archPath := range archPaths {
		arch := archPath.Name()
		if !archPath.IsDir() {
			logrus.Infof("Skipping %s because it's not a directory", arch)
			continue
		}

		if err := filepath.Walk(
			filepath.Join(releaseImagesPath, arch),
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}

				fileName := info.Name()
				if !strings.HasSuffix(fileName, ".tar") {
					logrus.Infof("Skipping non-tarball %s", fileName)
					return nil
				}

				origTag, err := i.client.RepoTagFromTarball(path)
				if err != nil {
					return errors.Wrap(err, "getting repo tags for tarball")
				}

				tagMatches := tagRegex.FindStringSubmatch(origTag)
				if len(tagMatches) != 2 {
					return errors.Errorf(
						"malformed tag %s in %s", origTag, fileName,
					)
				}

				binary := tagMatches[1]
				newTag := filepath.Join(
					registry,
					strings.TrimSuffix(binary, "-"+arch),
				)
				newTagWithArch := fmt.Sprintf("%s-%s:%s", newTag, arch, version)
				manifestImages[newTag] = append(manifestImages[newTag], arch)

				if err := i.client.Execute(
					"docker", "load", "-qi", path,
				); err != nil {
					return errors.Wrap(err, "load container image")
				}

				if err := i.client.Execute(
					"docker", "tag", origTag, newTagWithArch,
				); err != nil {
					return errors.Wrap(err, "tag container image")
				}

				logrus.Infof("Pushing %s", newTagWithArch)

				if err := i.client.Execute(
					"gcloud", "docker", "--", "push", newTagWithArch,
				); err != nil {
					return errors.Wrap(err, "push container image")
				}

				if err := i.client.Execute(
					"docker", "rmi", origTag, newTagWithArch,
				); err != nil {
					return errors.Wrap(err, "remove local container image")
				}

				return nil
			},
		); err != nil {
			return errors.Wrap(err, "traversing path")
		}
	}

	for image, arches := range manifestImages {
		imageVersion := fmt.Sprintf("%s:%s", image, version)
		logrus.Infof("Creating manifest image %s", imageVersion)

		manifests := []string{}
		for _, arch := range arches {
			manifests = append(manifests,
				fmt.Sprintf("%s-%s:%s", image, arch, version),
			)
		}
		if err := i.client.Execute("docker", append(
			[]string{"manifest", "create", "--amend", imageVersion},
			manifests...,
		)...); err != nil {
			return errors.Wrap(err, "create manifest")
		}

		for _, arch := range arches {
			logrus.Infof(
				"Annotating %s-%s:%s with --arch %s",
				image, arch, version, arch,
			)
			if err := i.client.Execute(
				"docker", "manifest", "annotate", "--arch", arch,
				imageVersion, fmt.Sprintf("%s-%s:%s", image, arch, version),
			); err != nil {
				return errors.Wrap(err, "annotate manifest with arch")
			}
		}

		logrus.Infof("Pushing manifest image %s", imageVersion)
		if err := i.client.Execute(
			"docker", "manifest", "push", imageVersion, "--purge",
		); err != nil {
			return errors.Wrap(err, "push manifest")
		}
	}

	return nil
}
