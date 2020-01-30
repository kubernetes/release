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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/log"
	"k8s.io/release/pkg/patch"
)

type opts struct {
	patch.AnnounceOptions
	K8sRepoURL      string
	K8sBranch       string
	ReleaseRepoURL  string
	ReleaseBranch   string
	ProjectID       string
	BuildConfigPath string
	Loglevel        string
	Tail            bool
}

const (
	defaultK8sRepoURL     = "https://github.com/kubernetes/kubernetes"
	defaultK8sBranch      = "master"
	defaultReleaseRepoURL = "https://github.com/kubernetes/release"
	defaultReleaseBranch  = "master"
	defaultGCPorjectID    = "kubernetes-release-test"
	defaultLogLevel       = "info"

	// separator which hopefully never appears in any of our keys/values.
	sep = "\001\002\001"
)

func main() {
	cmd := getCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getCommand() *cobra.Command {
	opts := opts{}

	cmd := &cobra.Command{
		Use:          "patch-announce <flags>",
		Long:         "submits a GCB job to run `krel patch-announce` in the clouds",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&opts.SenderName, "sender-name", "n", "", "email sender's name")
	cmd.Flags().StringVarP(&opts.SenderEmail, "sender-email", "m", "", "email sender's address")
	cmd.Flags().StringVarP(&opts.FreezeDate, "freeze-date", "f", "", "date when no CPs are allowed anymore")
	cmd.Flags().StringVarP(&opts.CutDate, "cut-date", "c", "", "date when the patch release is planned to be cut")
	cmd.Flags().StringVarP(&opts.BuildConfigPath, "config", "C", "", "file path to the patch-announce cloudbuild.yaml")
	cmd.Flags().StringVarP(&opts.ProjectID, "project-id", "p", defaultGCPorjectID, "Google Project ID")
	cmd.Flags().StringVarP(&opts.K8sRepoURL, "kubernetes-repo-url", "r", defaultK8sRepoURL, `git URL for the kubernetes repo ("k/k")`)
	cmd.Flags().StringVarP(&opts.K8sBranch, "kubernetes-branch", "b", defaultK8sBranch, `branch to checkout for the kubernetes repo ("k/k")`)
	cmd.Flags().StringVarP(&opts.ReleaseRepoURL, "release-repo-url", "R", defaultReleaseRepoURL, `git URL for the release repo ("k/release)`)
	cmd.Flags().StringVarP(&opts.ReleaseBranch, "release-branch", "B", defaultReleaseBranch, `branch to checkout for the release repo ("k/release")`)
	cmd.Flags().StringVarP(&opts.Loglevel, "log-level", "l", defaultLogLevel, "log level on the GCB job")
	cmd.Flags().BoolVarP(&opts.Tail, "tail", "t", false, "tail the build")
	cmd.Flags().BoolVar(&opts.Nomock, "nomock", false, `run in nomock (="real") mode or not`)

	cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return log.SetupGlobalLogger(opts.Loglevel)
	}

	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		for _, f := range []string{"sender-name", "sender-email", "freeze-date", "cut-date"} {
			if err := cmd.MarkFlagRequired(f); err != nil {
				return err
			}
		}

		if opts.BuildConfigPath == "" {
			p, err := kReleaseLocalPath()
			if err != nil {
				return err
			}
			opts.BuildConfigPath = filepath.Join(p, "..", "..", "gcb", "patch-announce", "cloudbuild.yaml")
			logrus.Debugf("no config filed specified, defaulting to %q", opts.BuildConfigPath)
		}

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		exeName := "gcloud"
		exe, err := exec.LookPath(exeName)
		if err != nil {
			return err
		}

		args := []string{
			exeName, "builds", "submit",
			"--no-source",
			"--project=" + opts.ProjectID,
			"--config=" + opts.BuildConfigPath,
		}

		if !opts.Tail {
			args = append(args, "--async")
		}

		subst := substitutions{
			"_K8S_GIT_URL":        opts.K8sRepoURL,
			"_K8S_GIT_BRANCH":     opts.K8sBranch,
			"_FREEZE_DATE":        opts.FreezeDate,
			"_CUT_DATE":           opts.CutDate,
			"_SENDER_NAME":        opts.SenderName,
			"_SENDER_EMAIL":       opts.SenderEmail,
			"_RELEASE_GIT_URL":    opts.ReleaseRepoURL,
			"_RELEASE_GIT_BRANCH": opts.ReleaseBranch,
			"_LOG_LEVEL":          opts.Loglevel,
			"_NOMOCK":             fmt.Sprintf("%t", opts.Nomock),
		}
		logrus.Debugf("about to run %q with substitutions [%s]", args, subst.human())

		args = append(args, "--substitutions="+subst.string())

		logrus.Infof("execing %q", exeName)
		return syscall.Exec(exe, args, os.Environ())
	}

	return cmd
}

type substitutions map[string]string

func (s substitutions) join(sep string) string {
	a := []string{}

	for k, v := range s {
		a = append(a, k+"="+v)
	}

	return strings.Join(a, sep)
}

func (s substitutions) string() string {
	// https://cloud.google.com/sdk/gcloud/reference/topic/escaping
	return "^" + sep + "^" + s.join(sep)
}

func (s substitutions) human() string {
	return s.join(", ")
}

func kReleaseLocalPath() (string, error) {
	if _, filename, _, ok := runtime.Caller(0); ok {
		return filepath.Dir(filename), nil
	}
	return "", fmt.Errorf("could not find the local path to k/release")
}
