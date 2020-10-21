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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"k8s.io/klog"
	reg "sigs.k8s.io/k8s-container-image-promoter/pkg/dockerregistry"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		// nolint[gomnd]
		os.Exit(1)
	}
}

// nolint[lll]
func run(ctx context.Context) error {
	klog.InitFlags(nil)

	baseDir := ""
	flag.StringVar(
		&baseDir,
		"base_dir",
		baseDir,
		"the manifest directory to look at and modify")
	stagingRepo := ""
	flag.StringVar(
		&stagingRepo,
		"staging_repo",
		stagingRepo,
		"the staging repo which we want to read from")
	filterImage := ""
	flag.StringVar(
		&filterImage,
		"filter_image",
		filterImage,
		"filter staging repo by this image name")
	filterDigest := ""
	flag.StringVar(
		&filterDigest,
		"filter_digest",
		filterDigest,
		"filter images by this digest")
	filterTag := ""
	flag.StringVar(
		&filterTag,
		"filter_tag",
		filterTag,
		"filter images by this tag")

	flag.Parse()

	opt := reg.GrowManifestOptions{}
	if err := opt.Populate(
		baseDir,
		stagingRepo,
		filterImage,
		filterDigest,
		filterTag); err != nil {
		return err
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	return reg.GrowManifest(ctx, opt)
}
