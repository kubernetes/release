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

package cli

import (
	"os"

	guuid "github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/cip/audit"
)

type AuditOptions struct {
	ProjectID    string
	RepoURL      string
	RepoBranch   string
	ManifestPath string
	UUID         string
}

func RunAuditCmd(opts *AuditOptions) error {
	opts.set()

	if err := validateAuditOptions(opts); err != nil {
		return errors.Wrap(err, "validating audit options")
	}

	auditorContext, err := audit.InitRealServerContext(
		opts.ProjectID,
		opts.RepoURL,
		opts.RepoBranch,
		opts.ManifestPath,
		opts.UUID,
	)
	if err != nil {
		return errors.Wrap(err, "creating auditor context")
	}

	auditorContext.RunAuditor()

	return nil
}

func (o *AuditOptions) set() {
	logrus.Infof("Setting image auditor options...")

	if o.ProjectID == "" {
		o.ProjectID = os.Getenv("CIP_AUDIT_GCP_PROJECT_ID")
	}

	if o.RepoURL == "" {
		o.RepoURL = os.Getenv("CIP_AUDIT_MANIFEST_REPO_URL")
	}

	if o.RepoBranch == "" {
		o.RepoBranch = os.Getenv("CIP_AUDIT_MANIFEST_REPO_BRANCH")
	}

	if o.ManifestPath == "" {
		o.ManifestPath = os.Getenv("CIP_AUDIT_MANIFEST_REPO_MANIFEST_DIR")
	}

	// TODO: Should we allow this to be configurable via the command line?
	o.UUID = os.Getenv("CIP_AUDIT_TESTCASE_UUID")
	if len(o.UUID) > 0 {
		logrus.Infof("Starting auditor in Test Mode (%s)", o.UUID)
	} else {
		o.UUID = guuid.NewString()
		logrus.Infof("Starting auditor in Regular Mode (%s)", o.UUID)
	}

	logrus.Infof(
		// nolint: lll
		"Image auditor options: [GCP project: %s, repo URL: %s, repo branch: %s, path: %s, UUID: %s]",
		o.RepoURL,
		o.RepoBranch,
		o.ManifestPath,
		o.ProjectID,
		o.UUID,
	)
}

func validateAuditOptions(o *AuditOptions) error {
	// TODO: Validate root options
	// TODO: Validate audit options
	return nil
}
