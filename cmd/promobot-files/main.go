/*
Copyright 2019 The Kubernetes Authors.

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

	// TODO: Use k/release/pkg/log instead
	"k8s.io/klog/v2"
	"k8s.io/release/pkg/promobot"
)

func main() {
	klog.InitFlags(nil)

	var options promobot.PromoteFilesOptions
	options.PopulateDefaults()

	flag.StringVar(
		&options.FilestoresPath,
		"filestores",
		options.FilestoresPath,
		"the manifest of filestores (REQUIRED)")
	flag.StringVar(
		&options.FilesPath,
		"files",
		options.FilesPath,
		"path to the files manifest (REQUIRED).  A directory can be specified.")
	flag.BoolVar(
		&options.DryRun,
		"dry-run",
		options.DryRun,
		"print what would have happened by running this tool;"+
			" do not actually modify any registry")

	flag.BoolVar(
		&options.UseServiceAccount,
		"use-service-account",
		options.UseServiceAccount,
		"allow service account usage with gcloud calls"+
			" (default: false)")

	flag.Parse()

	ctx := context.Background()
	if err := promobot.RunPromoteFiles(ctx, options); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		// nolint[gomnd]
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
