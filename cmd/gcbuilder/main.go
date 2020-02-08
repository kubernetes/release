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
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/release/pkg/gcp/build"
)

func parseFlags() build.Options {
	o := build.Options{}
	flag.StringVar(&o.BuildDir, "build-dir", "", "If provided, this directory will be uploaded as the source for the Google Cloud Build run.")
	flag.StringVar(&o.CloudbuildFile, "gcb-config", "cloudbuild.yaml", "If provided, this will be used as the name of the Google Cloud Build config file.")
	flag.StringVar(&o.LogDir, "log-dir", "", "If provided, build logs will be sent to files in this directory instead of to stdout/stderr.")
	flag.StringVar(&o.ScratchBucket, "scratch-bucket", "", "The complete GCS path for Cloud Build to store scratch files (sources, logs).")
	flag.StringVar(&o.Project, "project", "", "If specified, use a non-default GCP project.")
	flag.BoolVar(&o.AllowDirty, "allow-dirty", false, "If true, allow pushing dirty builds.")
	flag.BoolVar(&o.NoSource, "no-source", false, "If true, no source will be uploaded with this build.")
	flag.StringVar(&o.Variant, "variant", "", "If specified, build only the given variant. An error if no variants are defined.")
	flag.StringVar(&o.EnvPassthrough, "env-passthrough", "", "Comma-separated list of specified environment variables to be passed to GCB as substitutions with an _ prefix. If the variable doesn't exist, the substitution will exist but be empty.")

	flag.Parse()

	if flag.NArg() < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "expected a config directory to be provided")
		os.Exit(1)
	}

	o.ConfigDir = strings.TrimSuffix(flag.Arg(0), "/")

	return o
}

func main() {
	o := parseFlags()

	if bazelWorkspace := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); bazelWorkspace != "" {
		if err := os.Chdir(bazelWorkspace); err != nil {
			log.Fatalf("Failed to chdir to bazel workspace (%s): %v", bazelWorkspace, err)
		}
	}

	if o.BuildDir == "" {
		o.BuildDir = o.ConfigDir
	}

	log.Printf("Build directory: %s\n", o.BuildDir)

	// Canonicalize the config directory to be an absolute path.
	// As we're about to cd into the build directory, we need a consistent way to reference the config files
	// when the config directory is not the same as the build directory.
	absConfigDir, absErr := filepath.Abs(o.ConfigDir)
	if absErr != nil {
		log.Fatalf("Could not resolve absolute path for config directory: %v", absErr)
	}

	o.ConfigDir = absConfigDir
	o.CloudbuildFile = path.Join(o.ConfigDir, o.CloudbuildFile)

	configDirErr := o.ValidateConfigDir()
	if configDirErr != nil {
		log.Fatalf("Could not validate config directory: %v", configDirErr)
	}

	log.Printf("Config directory: %s\n", o.ConfigDir)

	log.Printf("cd-ing to build directory: %s\n", o.BuildDir)
	if err := os.Chdir(o.BuildDir); err != nil {
		log.Fatalf("Failed to chdir to build directory (%s): %v", o.BuildDir, err)
	}

	errors := build.RunBuildJobs(o)
	if len(errors) != 0 {
		log.Fatalf("Failed to run some build jobs: %v", errors)
	}
	log.Println("Finished.")
}
