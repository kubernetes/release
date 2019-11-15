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

package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubepkg",
	Short: "kubepkg",
}

type rootOptions struct {
}

var rootOpts = &rootOptions{}

type ChannelType string

const (
	ChannelRelease ChannelType = "release"
	ChannelTesting ChannelType = "testing"
	ChannelNightly ChannelType = "nightly"

	minimumKubernetesVersion = "1.13.0"
	minimumCNIVersion        = "0.7.5"
	pre117CNIVersion         = "0.7.5"

	defaultRevision = "0"

	packagesRootDir = "packages"

	kubeadmConf = "10-kubeadm.conf"
)

var (
	minimumCRIToolsVersion = minimumKubernetesVersion
	latestPackagesDir      = fmt.Sprintf("%s/%s", packagesRootDir, "latest")
)

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

type build struct {
	Package     string
	Definitions []packageDefinition
}

type packageDefinition struct {
	Name     string
	Version  string
	Revision string

	Channel           ChannelType
	KubernetesVersion string
	KubeletCNIVersion string

	DownloadLinkBase         string
	KubeadmKubeletConfigFile string

	CNIDownloadLink string
}

type cfg struct {
	*packageDefinition
	Arch         string
	DebArch      string
	Package      string
	Dependencies string
}

type stringList []string

func (ss *stringList) String() string {
	return strings.Join(*ss, ",")
}
func (ss *stringList) Set(v string) error {
	*ss = strings.Split(v, ",")
	return nil
}

var (
	revision        string
	kubeVersion     string
	cniVersion      string
	criToolsVersion string

	packages      = stringList{"kubelet", "kubectl", "kubeadm", "kubernetes-cni", "cri-tools"}
	channels      = stringList{"release", "testing", "nightly"}
	architectures = stringList{"amd64", "arm", "arm64", "ppc64le", "s390x"}

	releaseDownloadLinkBase = "https://dl.k8s.io"

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep-tmp", false, "keep tmp dir after build")
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().Var(&packages, "packages", "packages to build")
	rootCmd.PersistentFlags().Var(&channels, "channels", "channels to build for")
	rootCmd.PersistentFlags().Var(&architectures, "arch", "architectures to build for")
	rootCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "", "Kubernetes version to build")
	rootCmd.PersistentFlags().StringVar(&revision, "revision", defaultRevision, "deb package revision.")
	rootCmd.PersistentFlags().StringVar(&cniVersion, "cni-version", "", "CNI version to build")
	rootCmd.PersistentFlags().StringVar(&criToolsVersion, "cri-tools-version", "", "CRI tools version to build")
	rootCmd.PersistentFlags().StringVar(&releaseDownloadLinkBase, "release-download-link-base", "https://dl.k8s.io", "release download link base.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
