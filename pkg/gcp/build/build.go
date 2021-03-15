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

package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/object"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/tar"
	"sigs.k8s.io/yaml"
)

const (
	gcsSourceDir = "/source"
	gcsLogsDir   = "/logs"

	DefaultCloudbuildFile = "cloudbuild.yaml"
)

// TODO: Pull some of these options in cmd/gcbuilder, so they don't have to be public.
type Options struct {
	objStore       object.Store
	BuildDir       string
	ConfigDir      string
	CloudbuildFile string
	LogDir         string
	ScratchBucket  string
	Project        string
	AllowDirty     bool
	NoSource       bool
	Async          bool
	DiskSize       string
	Variant        string
	EnvPassthrough string
}

// NewDefaultOptions returns a new default `*Options` instance.
func NewDefaultOptions() *Options {
	return &Options{
		objStore:       object.NewGCS(),
		Project:        release.DefaultKubernetesStagingProject,
		CloudbuildFile: DefaultCloudbuildFile,
	}
}

func PrepareBuilds(o *Options) error {
	if o.ConfigDir == "" {
		return errors.New("expected a config directory to be provided")
	}

	if bazelWorkspace := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); bazelWorkspace != "" {
		if err := os.Chdir(bazelWorkspace); err != nil {
			return errors.Wrapf(err, "failed to chdir to bazel workspace (%s)", bazelWorkspace)
		}
	}

	if o.BuildDir == "" {
		o.BuildDir = o.ConfigDir
	}

	logrus.Infof("Build directory: %s", o.BuildDir)

	// Canonicalize the config directory to be an absolute path.
	// As we're about to cd into the build directory, we need a consistent way to reference the config files
	// when the config directory is not the same as the build directory.
	absConfigDir, absErr := filepath.Abs(o.ConfigDir)
	if absErr != nil {
		return errors.Wrapf(absErr, "could not resolve absolute path for config directory")
	}

	o.ConfigDir = absConfigDir
	o.CloudbuildFile = path.Join(o.ConfigDir, o.CloudbuildFile)

	configDirErr := o.ValidateConfigDir()
	if configDirErr != nil {
		return errors.Wrapf(configDirErr, "could not validate config directory")
	}

	logrus.Infof("Config directory: %s", o.ConfigDir)

	logrus.Infof("Changing to build directory: %s", o.BuildDir)
	if err := os.Chdir(o.BuildDir); err != nil {
		return errors.Wrapf(err, "failed to chdir to build directory (%s)", o.BuildDir)
	}

	return nil
}

func (o *Options) ValidateConfigDir() error {
	configDir := o.ConfigDir
	dirInfo, err := os.Stat(o.ConfigDir)
	if os.IsNotExist(err) {
		logrus.Infof("Config directory (%s) does not exist", configDir)
		return err
	}

	if !dirInfo.IsDir() {
		logrus.Infof("Config directory (%s) is not actually a directory", configDir)
		return err
	}

	_, err = os.Stat(o.CloudbuildFile)
	if os.IsNotExist(err) {
		logrus.Infof("%s does not exist", o.CloudbuildFile)
		return err
	}

	return nil
}

func (o *Options) uploadBuildDir(targetBucket string) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temp file")
	}
	name := f.Name()
	_ = f.Close()
	defer os.Remove(name)

	logrus.Infof("Creating source tarball at %s...", name)
	if err := tar.Compress(name, ".", regexp.MustCompile(".git")); err != nil {
		return "", errors.Wrap(err, "create tarball")
	}

	u := uuid.New()
	uploaded := fmt.Sprintf("%s/%s.tgz", targetBucket, u.String())
	logrus.Infof("Uploading %s to %s...", name, uploaded)
	if err := o.objStore.CopyToRemote(
		name,
		uploaded,
	); err != nil {
		return "", errors.Wrap(err, "upload files to GCS")
	}

	return uploaded, nil
}

func getExtraSubs(o *Options) map[string]string {
	envs := strings.Split(o.EnvPassthrough, ",")
	subs := map[string]string{}
	for _, e := range envs {
		e = strings.TrimSpace(e)
		if e != "" {
			subs[e] = os.Getenv(e)
		}
	}
	return subs
}

func RunSingleJob(o *Options, jobName, uploaded, version string, subs map[string]string) error {
	s := make([]string, 0, len(subs)+1)
	for k, v := range subs {
		s = append(s, fmt.Sprintf("_%s=%s", k, v))
	}

	s = append(s, "_GIT_TAG="+version)
	args := []string{
		"builds", "submit",
		"--verbosity", "info",
		"--config", o.CloudbuildFile,
		"--substitutions", strings.Join(s, ","),
	}

	if o.Project != "" {
		args = append(args, "--project", o.Project)
	}

	if o.Async {
		args = append(args, "--async")
	}

	if o.ScratchBucket != "" {
		args = append(
			args,
			"--gcs-log-dir",
			o.ScratchBucket+gcsLogsDir,
			"--gcs-source-staging-dir",
			o.ScratchBucket+gcsSourceDir,
		)
	}

	if uploaded != "" {
		args = append(args, uploaded)
	} else {
		if o.NoSource {
			args = append(args, "--no-source")
		} else {
			args = append(args, ".")
		}
	}

	if o.DiskSize != "" {
		diskSizeInt, intErr := strconv.Atoi(o.DiskSize)
		if intErr != nil {
			return intErr
		}

		if diskSizeInt > 1000 {
			return errors.New("Selected disk size must be no greater than 1000 GB")
		} else if diskSizeInt <= 0 {
			return errors.New("Selected disk size must be greater than 0 GB")
		}

		diskSizeArg := fmt.Sprintf("--disk-size=%s", o.DiskSize)
		args = append(args, diskSizeArg)
	}

	cmd := command.New(gcp.GCloudExecutable, args...)

	if o.LogDir != "" {
		p := path.Join(o.LogDir, strings.ReplaceAll(jobName, "/", "-")+".log")
		f, err := os.Create(p)
		if err != nil {
			return errors.Wrapf(err, "couldn't create %s", p)
		}

		defer f.Close()
		cmd.AddWriter(f)
	}

	logrus.Infof("cloudbuild command to send to gcp: %s", cmd.String())
	if err := cmd.RunSuccess(); err != nil {
		return errors.Wrapf(err, "error running %s", cmd.String())
	}

	return nil
}

type variants map[string]map[string]string

func getVariants(o *Options) (variants, error) {
	content, err := os.ReadFile(path.Join(o.ConfigDir, "variants.yaml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to load variants.yaml")
		}
		if o.Variant != "" {
			return nil, errors.Errorf("no variants.yaml found, but a build variant (%q) was specified", o.Variant)
		}
		return nil, nil
	}
	v := struct {
		Variants variants `json:"variants"`
	}{}
	if err := yaml.UnmarshalStrict(content, &v); err != nil {
		return nil, errors.Wrapf(err, "failed to read variants.yaml")
	}
	if o.Variant != "" {
		va, ok := v.Variants[o.Variant]
		if !ok {
			return nil, errors.Errorf("requested variant %q, which is not present in variants.yaml", o.Variant)
		}
		return variants{o.Variant: va}, nil
	}
	return v.Variants, nil
}

func RunBuildJobs(o *Options) []error {
	var uploaded string
	if o.ScratchBucket != "" {
		if !o.NoSource {
			var err error
			uploaded, err = o.uploadBuildDir(o.ScratchBucket + gcsSourceDir)
			if err != nil {
				return []error{errors.Wrapf(err, "failed to upload source")}
			}
		}
	} else {
		logrus.Info("Skipping advance upload and relying on gcloud...")
	}

	logrus.Info("Running build jobs...")
	repo := release.NewRepo()
	err := repo.Open()
	if err != nil {
		return []error{err}
	}

	tag, err := repo.GetTag()
	if err != nil {
		return []error{err}
	}

	if !o.AllowDirty && strings.HasSuffix(tag, "-dirty") {
		return []error{errors.New("the working copy is dirty")}
	}

	vs, err := getVariants(o)
	if err != nil {
		return []error{err}
	}
	if len(vs) == 0 {
		logrus.Info("No variants.yaml, starting single build job...")
		if err := RunSingleJob(o, "build", uploaded, tag, getExtraSubs(o)); err != nil {
			return []error{err}
		}
		return nil
	}

	logrus.Infof("Found variants.yaml, starting %d build jobs...", len(vs))

	w := sync.WaitGroup{}
	w.Add(len(vs))
	var jobErrors []error
	extraSubs := getExtraSubs(o)
	for k, v := range vs {
		go func(job string, vc map[string]string) {
			defer w.Done()
			logrus.Infof("Starting job %q...", job)
			if err := RunSingleJob(o, job, uploaded, tag, mergeMaps(extraSubs, vc)); err != nil {
				logrus.Infof("Job %q failed: %v", job, err)
				jobErrors = append(jobErrors, errors.Wrapf(err, "job %q failed", job))
			} else {
				logrus.Infof("Job %q completed", job)
			}
		}(k, v)
	}
	w.Wait()
	return jobErrors
}

func mergeMaps(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func ListJobs(project string, lastJobs int64) error {
	ctx := context.Background()
	opts := option.WithCredentialsFile("")

	service, err := cloudbuild.NewService(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "failed to fetching gcloud credentials... try running \"gcloud auth application-default login\"")
	}

	req, err := service.Projects.Builds.List(project).PageSize(lastJobs).Do()
	if err != nil {
		return errors.Wrap(err, "failed to listing the builds")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Start Time", "Finish Time", "Status", "Console Logs"})

	for _, build := range req.Builds {
		table.Append([]string{strings.TrimSpace(build.StartTime), strings.TrimSpace(build.FinishTime), strings.TrimSpace(build.Status), strings.TrimSpace(build.LogUrl)})
	}
	table.Render()

	return nil
}

func GetJobsByTag(project, tagsFilter string) ([]*cloudbuild.Build, error) {
	ctx := context.Background()
	opts := option.WithCredentialsFile("")

	service, err := cloudbuild.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetching gcloud credentials... try running \"gcloud auth application-default login\"")
	}

	req, err := service.Projects.Builds.List(project).Filter(tagsFilter).PageSize(50).Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to listing the builds")
	}

	return req.Builds, nil
}
