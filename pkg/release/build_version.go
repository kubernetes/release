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

package release

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/testgrid"
	"k8s.io/release/pkg/util"
)

// SetBuildVersion sets the build version for a branch
// against a set of blocking CI jobs
//
// branch - The branch name.
// jobPath - A local directory to store the copied cache entries.
// exclude_suites - A list of (greedy) patterns to exclude CI jobs from
// 					checking against the primary job.
func SetBuildVersion(
	branch, jobPath string,
	excludeSuites []string,
) error {
	logrus.Infof("Setting build version for branch %q", branch)

	if branch == git.Master {
		branch = "release-master"
		logrus.Infof("Changing %s branch to %q", git.Master, branch)
	}

	allJobs, err := testgrid.New().BlockingTests(branch)
	if err != nil {
		return errors.Wrap(err, "getting all test jobs")
	}
	logrus.Infof("Got testgrid jobs for branch %q: %v", branch, allJobs)

	if len(allJobs) == 0 {
		return errors.Errorf(
			"No sig-%s-blocking list found in the testgrid config.yaml", branch,
		)
	}

	// Filter out excluded suites
	secondaryJobs := []string{}
	for _, job := range allJobs {
		for _, pattern := range excludeSuites {
			matched, err := regexp.MatchString(pattern, job)
			if err != nil {
				return errors.Wrapf(err, "regex comile failed: %s", pattern)
			}
			if matched {
				secondaryJobs = append(secondaryJobs, job)
			}
		}
	}

	// Update main cache
	// We dedup the $main_job's list of successful runs and just run through
	// that unique list. We then leave the full state of secondaries below so
	// we have finer granularity at the Jenkin's job level to determine if a
	// build is ok.
	jobCaches := map[string]JobCache{} // a map of jobs and their caches
	mainJob := allJobs[0]
	jcc := NewJobCacheClient()
	mainJobCache, err := jcc.GetJobCache(mainJob, true)
	if err != nil {
		return errors.Wrap(err, "building job cache for main job")
	}
	jobCaches[mainJob] = mainJobCache

	// Update secondary caches limited by main cache last build number
	for _, job := range secondaryJobs {
		cache, err := jcc.GetJobCache(job, true)
		if err != nil {
			return errors.Wrapf(err, "building job cache for job: %s", job)
		}
		jobCaches[job] = cache
	}

	// TODO: continue port from releaselib.sh::set_build_version

	return errors.New("unimplemented")
}

type JobCacheClient struct {
	gcpClient gcpClient
}

// NewJobCacheClient creates a new job cache retrieval client
func NewJobCacheClient() *JobCacheClient {
	return &JobCacheClient{
		gcpClient: &defaultGcpClient{},
	}
}

func (j *JobCacheClient) SetClient(client gcpClient) {
	j.gcpClient = client
}

//counterfeiter:generate . gcpClient
type gcpClient interface {
	CopyJobCache(string) (string, error)
}

type defaultGcpClient struct{}

func (g *defaultGcpClient) CopyJobCache(job string) (jsonPath string, err error) {
	jsonPath = filepath.Join(os.TempDir(), fmt.Sprintf("job-cache-%s", job))

	const logRoot = "gs://kubernetes-jenkins/logs/"
	if err := gcp.GSUtil(
		"-qm", "cp",
		logRoot+filepath.Join(job, "jobResultsCache.json"),
		jsonPath,
	); err != nil {
		return "", errors.Wrap(err, "copying job results cache")
	}
	return jsonPath, nil
}

// JobCache is a map of build numbers (key) and their versions (value)
type JobCache map[string]string

// GetJobCache pulls Jenkins server job cache from GS and resutns a `JobCache`
//
// job - The Jenkins job name.
// dedup -  dedup git's monotonically increasing (describe) build numbers.
func (j *JobCacheClient) GetJobCache(job string, dedup bool) (JobCache, error) {
	logrus.Infof("Getting %s build results from GCS", job)

	tempJSON, err := j.gcpClient.CopyJobCache(job)
	if err != nil {
		return nil, errors.Wrap(err, "getting GCP job cache")
	}

	if !util.Exists(tempJSON) {
		// If there's no file up on job doesn't exist: Skip it.
		logrus.Infof("Skipping non existing job: %s", job)
		return nil, nil
	}
	defer os.RemoveAll(tempJSON)

	// Additional select on .version is because we have so many empty versions
	// for now 2 passes. First pass sorts by buildnumber, second builds the
	// dictionary.
	out, err := command.New("jq", "-r",
		`.[] | `+
			`select(.result == "SUCCESS") | `+
			`select(.version != null) | `+
			`[.version,.buildnumber] | "\(.[0]|rtrimstr("\n")) \(.[1])"`,
		tempJSON,
	).Pipe("sort", "-rn", "-k2,2").RunSilentSuccessOutput()
	if err != nil {
		return nil, errors.Wrap(err, "filtering job cache")
	}

	lastVersion := ""
	res := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(out.OutputTrimNL()))
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		if len(split) != 2 {
			return nil, errors.Wrapf(err,
				"unexpected string in job results cache %s: %s",
				tempJSON, scanner.Text(),
			)
		}

		version := split[0]
		buildNumber := split[1]

		if dedup && version == lastVersion {
			continue
		}
		lastVersion = version

		if buildNumber != "" && version != "" {
			res[buildNumber] = version
		}
	}

	return res, nil
}
