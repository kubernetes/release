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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Push pushes the build by taking the internal options into account.
func (bi *Instance) Build() error {
	/*
		def main(args):
				# pylint: disable=too-many-branches
				"""Build and push kubernetes.

				This is a python port of the kubernetes/hack/jenkins/build.sh script.
				"""
				if os.path.split(os.getcwd())[-1] != 'kubernetes':
						print >>sys.stderr, (
								'Scenario should only run from either kubernetes directory!')
						sys.exit(1)
	*/

	buildExists, buildExistsErr := checkBuildExists("foo", "bar", "baz")
	if buildExistsErr != nil {
		return errors.Wrapf(buildExistsErr, "checking if build exists")
	}

	if buildExists {
		logrus.Infof("Build already exists. Exiting...")
		return nil
	}

	/*
		env = {
				# Skip gcloud update checking; do we still need this?
				'CLOUDSDK_COMPONENT_MANAGER_DISABLE_UPDATE_CHECK': 'true',
				# Don't run any unit/integration tests when building
				'KUBE_RELEASE_RUN_TESTS': 'n',
		}
	*/

	/*
		push_build_args = ['--nomock', '--verbose', '--ci']
		if args.suffix:
				push_build_args.append('--gcs-suffix=%s' % args.suffix)
		if args.release:
				push_build_args.append('--bucket=%s' % args.release)
		if args.registry:
				push_build_args.append('--docker-registry=%s' % args.registry)
		if args.extra_version_markers:
				push_build_args.append('--extra-version-markers=%s' % args.extra_version_markers)
		if args.fast:
				push_build_args.append('--fast')
		if args.allow_dup:
				push_build_args.append('--allow-dup')
		if args.skip_update_latest:
				push_build_args.append('--noupdatelatest')
	*/

	/*
		if args.register_gcloud_helper:
				# Configure docker client for gcr.io authentication to allow communication
				# with non-public registries.
				check_no_stdout('gcloud', 'auth', 'configure-docker')
	*/

	/*
		for key, value in env.items():
				os.environ[key] = value
	*/

	/*
		check('make', 'clean')
	*/

	/*
		if args.fast:
				check('make', 'quick-release')
		else:
				check('make', 'release')
	*/

	// Pushing the build
	pushBuildErr := NewInstance(bi.opts).Push()
	if pushBuildErr != nil {
		return errors.Wrapf(pushBuildErr, "pushing build")
	}

	return nil
}

// checkBuildExists check if the target build exists in the specified GCS
// bucket. This allows us to speed up build jobs by not duplicating builds.
func checkBuildExists(gcs, suffix, fast string) (bool, error) {
	/*
		def check_build_exists(gcs, suffix, fast):
			""" check if a k8s build with same version
					already exists in remote path
			"""
			if not os.path.exists('hack/print-workspace-status.sh'):
					print >>sys.stderr, 'hack/print-workspace-status.sh not found, continue'
					return False
	*/

	/*
		version = ''
		try:
				match = re.search(
						r'gitVersion ([^\n]+)',
						check_output('hack/print-workspace-status.sh')
				)
				if match:
						version = match.group(1)
		except subprocess.CalledProcessError as exc:
				# fallback with doing a real build
				print >>sys.stderr, 'Failed to get k8s version, continue: %s' % exc
				return False

		if version:
				if not gcs:
						gcs = 'kubernetes-release-dev'
				gcs = 'gs://' + gcs
				mode = 'ci'
				if fast:
						mode += '/fast'
				if suffix:
						mode += suffix
				gcs = os.path.join(gcs, mode, version)
				try:
						check_no_stdout('gsutil', 'ls', gcs)
						check_no_stdout('gsutil', 'ls', gcs + "/kubernetes.tar.gz")
						check_no_stdout('gsutil', 'ls', gcs + "/bin")
						return True
				except subprocess.CalledProcessError as exc:
						print >>sys.stderr, (
								'gcs path %s (or some files under it) does not exist yet, continue' % gcs)
		return False
	*/

	return true, nil
}
