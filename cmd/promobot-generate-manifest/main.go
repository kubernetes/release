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
	"path/filepath"

	"golang.org/x/xerrors"

	"k8s.io/klog"
	"k8s.io/release/pkg/promobot"
	"sigs.k8s.io/yaml"
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

	var opt promobot.GenerateManifestOptions
	opt.PopulateDefaults()

	src := ""
	flag.StringVar(
		&src,
		"src",
		src,
		"the base directory to copy from")

	flag.StringVar(
		&opt.Prefix,
		"prefix",
		opt.Prefix,
		"restrict the exported files; only export those starting with the provided prefix")

	flag.Parse()

	if src == "" {
		return xerrors.New("must specify --src")
	}

	s, err := filepath.Abs(src)
	if err != nil {
		return xerrors.Errorf("cannot resolve %q to absolute path: %w", src, err)
	}
	opt.BaseDir = s

	manifest, err := promobot.GenerateManifest(ctx, opt)
	if err != nil {
		return err
	}

	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return xerrors.Errorf("error serializing manifest: %w", err)
	}

	if _, err := os.Stdout.Write(manifestYAML); err != nil {
		return err
	}

	return nil
}
