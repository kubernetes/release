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

package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kpkg "k8s.io/release/pkg/kubepkg"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kubepkg",
	Short:   "kubepkg",
	PreRunE: initLogging,
}

type rootOptions struct {
	revision        string
	kubeVersion     string
	cniVersion      string
	criToolsVersion string

	packages      []string
	channels      []string
	architectures []string

	releaseDownloadLinkBase string

	templateDir string
	specOnly    bool

	logLevel string
}

var rootOpts = &rootOptions{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringArrayVar(
		&rootOpts.packages,
		"packages",
		kpkg.DefaultPackages,
		"packages to build",
	)
	rootCmd.PersistentFlags().StringArrayVar(
		&rootOpts.channels,
		"channels",
		kpkg.DefaultChannels,
		"channels to build for",
	)
	rootCmd.PersistentFlags().StringArrayVar(
		&rootOpts.architectures,
		"arch",
		kpkg.DefaultArchitectures,
		"architectures to build for",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.kubeVersion,
		"kube-version",
		"",
		"Kubernetes version to build",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.revision,
		"revision",
		kpkg.DefaultRevision,
		"deb package revision.",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.cniVersion,
		"cni-version",
		"",
		"CNI version to build",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.criToolsVersion,
		"cri-tools-version",
		"",
		"CRI tools version to build",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.releaseDownloadLinkBase,
		"release-download-link-base",
		kpkg.DefaultReleaseDownloadLinkBase,
		"release download link base",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.templateDir,
		"template-dir",
		kpkg.LatestTemplateDir,
		"template directory",
	)
	rootCmd.PersistentFlags().BoolVar(
		&rootOpts.specOnly,
		"spec-only",
		false,
		"only create specs instead of building packages",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.logLevel,
		"log-level",
		"info",
		"the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace'",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}

func initLogging(*cobra.Command, []string) error {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	lvl, err := logrus.ParseLevel(rootOpts.logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	logrus.Debugf("Using log level %q", lvl)
	return nil
}
