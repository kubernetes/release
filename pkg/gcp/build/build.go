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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"sigs.k8s.io/yaml"
)

const (
	gcsSourceDir = "/source"
	gcsLogsDir   = "/logs"
)

// TODO: Pull some of these options in cmd/gcbuilder, so they don't have to be public.
type Options struct {
	BuildDir       string
	ConfigDir      string
	CloudbuildFile string
	LogDir         string
	ScratchBucket  string
	Project        string
	AllowDirty     bool
	NoSource       bool
	Variant        string
	EnvPassthrough string
}

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getVersion() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	t := time.Now().Format("20060102")
	return fmt.Sprintf("v%s-%s", t, strings.TrimSpace(string(output))), nil
}

func (o *Options) ValidateConfigDir() error {
	configDir := o.ConfigDir
	dirInfo, err := os.Stat(o.ConfigDir)
	if os.IsNotExist(err) {
		log.Fatalf("Config directory (%s) does not exist", configDir)
	}

	if !dirInfo.IsDir() {
		log.Fatalf("Config directory (%s) is not actually a directory", configDir)
	}

	_, err = os.Stat(o.CloudbuildFile)
	if os.IsNotExist(err) {
		log.Fatalf("%s does not exist", o.CloudbuildFile)
	}

	return nil
}

func (o *Options) uploadBuildDir(targetBucket string) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	name := f.Name()
	_ = f.Close()
	defer os.Remove(name)

	log.Printf("Creating source tarball at %s...\n", name)
	if err := runCmd("tar", "--exclude", ".git", "-czf", name, "."); err != nil {
		return "", fmt.Errorf("failed to tar files: %s", err)
	}

	u := uuid.New()
	uploaded := fmt.Sprintf("%s/%s.tgz", targetBucket, u.String())
	log.Printf("Uploading %s to %s...\n", name, uploaded)
	if err := runCmd("gsutil", "cp", name, uploaded); err != nil {
		return "", fmt.Errorf("failed to upload files: %s", err)
	}

	return uploaded, nil
}

func getExtraSubs(o Options) map[string]string {
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

func runSingleJob(o Options, jobName, uploaded, version string, subs map[string]string) error {
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

	if o.ScratchBucket != "" {
		args = append(args, "--gcs-log-dir", o.ScratchBucket+gcsLogsDir)
		args = append(args, "--gcs-source-staging-dir", o.ScratchBucket+gcsSourceDir)
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

	cmd := exec.Command("gcloud", args...)

	if o.LogDir != "" {
		p := path.Join(o.LogDir, strings.Replace(jobName, "/", "-", -1)+".log")
		f, err := os.Create(p)

		if err != nil {
			return fmt.Errorf("couldn't create %s: %v", p, err)
		}

		defer f.Close()

		cmd.Stdout = f
		cmd.Stderr = f
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %s: %v", cmd.Args, err)
	}

	return nil
}

type variants map[string]map[string]string

func getVariants(o Options) (variants, error) {
	content, err := ioutil.ReadFile(path.Join(o.ConfigDir, "variants.yaml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load variants.yaml: %v", err)
		}
		if o.Variant != "" {
			return nil, fmt.Errorf("no variants.yaml found, but a build variant (%q) was specified", o.Variant)
		}
		return nil, nil
	}
	v := struct {
		Variants variants `json:"variants"`
	}{}
	if err := yaml.UnmarshalStrict(content, &v); err != nil {
		return nil, fmt.Errorf("failed to read variants.yaml: %v", err)
	}
	if o.Variant != "" {
		va, ok := v.Variants[o.Variant]
		if !ok {
			return nil, fmt.Errorf("requested variant %q, which is not present in variants.yaml", o.Variant)
		}
		return variants{o.Variant: va}, nil
	}
	return v.Variants, nil
}

func RunBuildJobs(o Options) []error {
	var uploaded string
	if o.ScratchBucket != "" {
		if !o.NoSource {
			var err error
			uploaded, err = o.uploadBuildDir(o.ScratchBucket + gcsSourceDir)
			if err != nil {
				return []error{fmt.Errorf("failed to upload source: %v", err)}
			}
		}
	} else {
		log.Println("Skipping advance upload and relying on gcloud...")
	}

	log.Println("Running build jobs...")
	tag, err := getVersion()
	if err != nil {
		return []error{fmt.Errorf("failed to get current tag: %v", err)}
	}

	if !o.AllowDirty && strings.HasSuffix(tag, "-dirty") {
		return []error{fmt.Errorf("the working copy is dirty")}
	}

	vs, err := getVariants(o)
	if err != nil {
		return []error{err}
	}
	if len(vs) == 0 {
		log.Println("No variants.yaml, starting single build job...")
		if err := runSingleJob(o, "build", uploaded, tag, getExtraSubs(o)); err != nil {
			return []error{err}
		}
		return nil
	}

	log.Printf("Found variants.yaml, starting %d build jobs...\n", len(vs))

	w := sync.WaitGroup{}
	w.Add(len(vs))
	var errors []error
	extraSubs := getExtraSubs(o)
	for k, v := range vs {
		go func(job string, vc map[string]string) {
			defer w.Done()
			log.Printf("Starting job %q...\n", job)
			if err := runSingleJob(o, job, uploaded, tag, mergeMaps(extraSubs, vc)); err != nil {
				errors = append(errors, fmt.Errorf("job %q failed: %v", job, err))
				log.Printf("Job %q failed: %v\n", job, err)
			} else {
				log.Printf("Job %q completed.\n", job)
			}
		}(k, v)
	}
	w.Wait()
	return errors
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
