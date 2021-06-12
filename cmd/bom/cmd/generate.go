/*
Copyright 2021 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/spdx"
	"sigs.k8s.io/release-utils/util"
)

var genOpts = &generateOptions{}

var generateCmd = &cobra.Command{
	Short: "bom generate → Create SPDX manifests",
	Long: `bom → Create SPDX manifests

generate is the bom subcommand to generate SPDX manifests.
Currently supports creating SBOM for files, images, and docker
archives (images in tarballs). Supports pulling images from
registries.

bom can take a deeper look into images using a growing number
of analyzers designed to add more sense to common base images.

`,
	Use:               "generate",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		for i, arg := range args {
			if util.Exists(arg) {
				file, err := os.Open(arg)
				if err != nil {
					return errors.Wrapf(err, "checking argument %d", i)
				}
				fileInfo, err := file.Stat()
				if err != nil {
					return errors.Wrapf(err, "calling stat on argument %d", i)
				}
				if fileInfo.IsDir() {
					genOpts.directories = append(genOpts.directories, arg)
				}
			}
		}
		return generateBOM(genOpts)
	},
}

type generateOptions struct {
	analyze        bool
	noGitignore    bool
	noGoModules    bool
	noGoTransient  bool
	namespace      string
	outputFile     string
	configFile     string
	images         []string
	tarballs       []string
	files          []string
	directories    []string
	ignorePatterns []string
}

// Validate verify options consistency
func (opts *generateOptions) Validate() error {
	if opts.configFile == "" &&
		len(opts.images) == 0 &&
		len(opts.files) == 0 &&
		len(opts.tarballs) == 0 &&
		len(opts.directories) == 0 {
		return errors.New("to generate a SPDX BOM you have to provide at least one image or file")
	}

	// A namespace URL is required
	if opts.configFile == "" && opts.namespace == "" {
		msg := "Error. No namespace defined\n\n"
		msg += "You did not specify a namespace for your document. This is an error.\n"
		msg += "To produce a valid SPDX SBOM, the document has to have an URI as its\n"
		msg += "namespace.\n\nIf you are testing, you can use http://example.com/ for now but your\n"
		msg += "final document must have a namespace URI pointing to the location where\n"
		msg += "your SBOM will be referenced in the future.\n\n"
		msg += "For more details, check the SPDX documentation here:\n"
		msg += "https://spdx.github.io/spdx-spec/2-document-creation-information/#25-spdx-document-namespace\n\n"
		msg += "Hint: --namespace is your friend here\n\n"
		logrus.Info(msg)

		return errors.New("A namespace URI must be defined to have a compliant SPDX BOM")
	}

	// Check namespace is a valid URL
	if _, err := url.Parse(opts.namespace); err != nil {
		return errors.Wrap(err, "parsing the namespace URL")
	}

	return nil
}

func init() {
	generateCmd.PersistentFlags().StringSliceVarP(
		&genOpts.images,
		"image",
		"i",
		[]string{},
		"list of images",
	)

	generateCmd.PersistentFlags().StringSliceVarP(
		&genOpts.files,
		"file",
		"f",
		[]string{},
		"list of files to include",
	)

	generateCmd.PersistentFlags().StringSliceVarP(
		&genOpts.tarballs,
		"tarball",
		"t",
		[]string{},
		"list of docker archive tarballs to include in the manifest",
	)

	generateCmd.PersistentFlags().StringSliceVarP(
		&genOpts.directories,
		"dirs",
		"d",
		[]string{},
		"list of directories to include in the manifest as packages",
	)

	generateCmd.PersistentFlags().StringSliceVar(
		&genOpts.ignorePatterns,
		"ignore",
		[]string{},
		"list of regexp patterns to ignore when scanning directories",
	)

	generateCmd.PersistentFlags().BoolVar(
		&genOpts.noGitignore,
		"no-gitignore",
		false,
		"don't use exclusions from .gitignore files",
	)

	generateCmd.PersistentFlags().BoolVar(
		&genOpts.noGoModules,
		"no-gomod",
		false,
		"don't perform go.mod analysis, sbom will not include data about go packages",
	)

	generateCmd.PersistentFlags().BoolVar(
		&genOpts.noGoTransient,
		"no-transient",
		false,
		"don't include transient go dependencies, only direct deps from go.mod",
	)

	generateCmd.PersistentFlags().StringVarP(
		&genOpts.namespace,
		"namespace",
		"n",
		"",
		"an URI that servers as namespace for the SPDX doc",
	)

	generateCmd.PersistentFlags().StringVarP(
		&genOpts.outputFile,
		"output",
		"o",
		"",
		"path to the file where the document will be written (defaults to STDOUT)",
	)

	generateCmd.PersistentFlags().BoolVarP(
		&genOpts.analyze,
		"analyze-images",
		"a",
		false,
		"go deeper into images using the available analyzers",
	)

	generateCmd.PersistentFlags().StringVarP(
		&genOpts.configFile,
		"config",
		"c",
		"",
		"path to yaml SBOM configuration file",
	)
}

func generateBOM(opts *generateOptions) error {
	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "validating command line options")
	}
	logrus.Info("Generating SPDX Bill of Materials")

	builder := spdx.NewDocBuilder()
	builderOpts := &spdx.DocGenerateOptions{
		Tarballs:         opts.tarballs,
		Files:            opts.files,
		Images:           opts.images,
		Directories:      opts.directories,
		OutputFile:       opts.outputFile,
		Namespace:        opts.namespace,
		AnalyseLayers:    opts.analyze,
		ProcessGoModules: !opts.noGoModules,
		OnlyDirectDeps:   !opts.noGoTransient,
		ConfigFile:       opts.configFile,
	}

	// We only replace the ignore patterns one or more where defined
	if len(opts.ignorePatterns) > 0 {
		builderOpts.IgnorePatterns = opts.ignorePatterns
	}
	doc, err := builder.Generate(builderOpts)
	if err != nil {
		return errors.Wrap(err, "generating doc")
	}

	if opts.outputFile == "" {
		markup, err := doc.Render()
		if err != nil {
			return errors.Wrap(err, "rendering document")
		}
		fmt.Println(markup)
	}
	return nil
}
