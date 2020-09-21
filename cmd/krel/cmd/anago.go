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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/release"
)

// TODO: This is an initial draft stub of anago's functionality.
//       Once we get further along here, we should break this out into two
//       distinct, constrained commands: 'stage' and 'release'.

var anagoOpts = &release.Options{} // TODO: Commenting these packages/commands out since they fail in CI.
//       These can be fixed by changing the CI test image to one that includes the packages.
//nolint:gocritic
/*
	requiredPackages = []string{
		"jq",
		"git",
		"bsdmainutils",
	}

	// TODO: Do we really need this if we use the Google Cloud SDK instead?
	requiredCommands = []string{
		"gsutil",
		"gcloud",
	}
*/ // anagoCmd is a krel subcommand which invokes runAnago()
var anagoCmd = &cobra.Command{
	Use:           "anago",
	Short:         "Run anago",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnago(anagoOpts)
	},
}

func init() {
	anagoCmd.PersistentFlags().StringVar(
		&anagoOpts.ReleaseType,
		"release-type",
		"",
		fmt.Sprintf("release type, must be one of: '%s'",
			strings.Join([]string{
				release.ReleaseTypeAlpha,
				release.ReleaseTypeBeta,
				release.ReleaseTypeRC,
				release.ReleaseTypeOfficial,
			}, "', '"),
		),
	)

	anagoCmd.PersistentFlags().StringVar(
		&anagoOpts.BaseDir,
		"base-dir",
		release.DefaultBaseDir,
		"",
	)

	for _, f := range []string{
		"release-type",
	} {
		if err := anagoCmd.MarkPersistentFlagRequired(f); err != nil {
			logrus.Fatalf("Unable to set %q flag as required: %v", f, err)
		}
	}

	rootCmd.AddCommand(anagoCmd)
}

// runAnago is the function invoked by 'krel anago', responsible for submitting release jobs to GCB
func runAnago(anagoOpts *release.Options) error {
	/*
	   # Clear or validate run state
	   if ((FLAGS_clean)) && [[ -f $PROGSTATE ]]; then
	     logrun mv $PROGSTATE $PROGSTATE.last
	   fi
	   common::validate_command_line "$ORIG_CMDLINE" || common::exit 1 "Exiting..."

	   # Validate command-line
	   common::argc_validate 1 || common::exit 1 "Exiting..."

	   # Set positional args
	   RELEASE_BRANCH=${POSITIONAL_ARGV[0]}

	   # Check branch format
	   [[ $RELEASE_BRANCH =~ $BRANCH_REGEX ]] \
	    || common::exit 1 "Invalid branch name!"

	   # Check arg conflicts
	   if ((FLAGS_rc)) || ((FLAGS_official)); then
	     if [[ "$RELEASE_BRANCH" == "master" ]]; then
	       common::exit 1 "Can't do release candidate or official releases on master!"
	     fi
	   fi

	   if ((FLAGS_rc)) && ((FLAGS_official)); then
	     common::exit 1 "Can't do release candidate and official release!"
	   fi
	*/

	// TODO: Handle nomock
	/*
		# Default mode is a mocked release workflow
		: ${FLAGS_nomock:=0}
		: ${FLAGS_rc:=0}
		: ${FLAGS_official:=0}
	*/

	// Create base and bin directories
	baseDir := anagoOpts.BaseDir
	binDir := filepath.Join(baseDir, "bin")

	if err := os.Setenv("BASEDIR", baseDir); err != nil {
		return err
	}
	if err := os.Setenv("BINDIR", binDir); err != nil {
		return err
	}

	if err := os.MkdirAll(binDir, os.FileMode(0755)); err != nil {
		return err
	}

	// Add BINDIR to path
	currentPath := os.Getenv("PATH")
	path := fmt.Sprintf("%s:%s", currentPath, binDir)
	if err := os.Setenv("PATH", path); err != nil {
		return err
	}

	logrus.Infof("Current PATH is: %s", os.Getenv("PATH"))

	/*
		##############################################################################
		# Initialize logs
		##############################################################################
		# Initialize and save up to 10 (rotated logs)
		if ((FLAGS_stage)); then
			LOGFILE=$TMPDIR/$PROG-stage.log
		else
			LOGFILE=$TMPDIR/$PROG.log
		fi
		common::logfileinit $LOGFILE 10
		# BEGIN script
		common::timestamp begin
	*/

	if !anagoOpts.GCB {
		return errors.New("cannot continue; releases must be run from GCB")
	}

	/*
		# Order workflow based on conditions
		((FLAGS_stage)) || common::stepindex "gitlib::github_acls"
		common::stepindex "check_prerequisites" "get_build_candidate" \
		 "prepare_workspace" "common::disk_space_check"
		common::stepindex "prepare_tree"
		 common::stepindex "local_kube_cross"
		if ((FLAGS_buildonly)); then
			common::stepindex "make_cross"
		elif ! ((FLAGS_prebuild)); then
			common::stepindex "make_cross"
			common::stepindex "build_tree" "generate_release_notes"
			if ((FLAGS_stage)); then
				common::stepindex "stage_source_tree"
			else
				common::stepindex "push_git_objects"
			fi
			common::stepindex "push_all_artifacts"
			if ! ((FLAGS_stage)); then
				common::stepindex "announce"
				common::stepindex "update_github_release"
			fi
		fi

		# Show the workflow order and completed steps
		common::stepheader "WORKFLOW STEPS"
		common::stepindex --toc $PROGSTATE
		logecho

		# Set cloud binaries
		if ! ((FLAGS_buildonly)) && ! common::set_cloud_binaries; then
			logecho "Releasing Kubernetes requires gsutil and gcloud. Please download,"
			logecho "install and authorize through the Google Cloud SDK:"
			logecho
			logecho "https://developers.google.com/cloud/sdk/"
			common::exit 1 "Exiting..."
		fi

		# Set the majorify of global values
		# Moved here b/c now depends on gcloud
		release::set_globals

		((FLAGS_stage)) || common::run_stateful gitlib::github_acls

		common::run_stateful "check_prerequisites"

		# Call run_stateful and store these globals in the $PROGSTATE
		common::run_stateful get_build_candidate JENKINS_BUILD_VERSION \
																						 PARENT_BRANCH BRANCH_POINT
		# Computes a few things and creates a globals and dictionaries - not stateful
		# Always Computed from JENKINS_BUILD_VERSION
		set_release_values || common::exit 1 "Exiting..."

		# Set values based on derived/computed values above
		# WORK/BUILD area
		# For --stage, it's JENKINS_BUILD_VERSION-based
		if ((FLAGS_stage)); then
			WORKDIR=$BASEDIR/$PROG-$JENKINS_BUILD_VERSION
		else
			WORKDIR=$BASEDIR/$PROG-$RELEASE_VERSION_PRIME
		fi

		if [[ $RELEASE_VERSION_PRIME =~ ${VER_REGEX[release]} ]]; then
			CHANGELOG_FILE="CHANGELOG-${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.md"
			CHANGELOG_FILEPATH="$CHANGELOG_DIR/$CHANGELOG_FILE"
		else
			common::exit 1 "Unable to set CHANGELOG file!"
		fi

		# Go tools expect the kubernetes src to be under $GOPATH
		export GOPATH="$WORKDIR"
		logecho "GOPATH: $GOPATH"

		# TOOL_ROOT is release/
		# TREE_ROOT is working branch/tree
		TREE_ROOT=$WORKDIR/src/k8s.io/kubernetes
		# Same as Makefile variable name
		OUT_DIR="_output"
		BUILD_OUTPUT=$TREE_ROOT/$OUT_DIR
		RELEASE_NOTES_HTML=$WORKDIR/src/release-notes.html

		# Ensure the WORKDIR exists
		logrun mkdir -p $WORKDIR

		# Display top pending PRs for branch releases (convenience only)
		if [[ $RELEASE_BRANCH =~ release- ]] &&
			 gitlib::branch_exists $RELEASE_BRANCH; then
			common::stepheader "PENDING PRS ON THE $RELEASE_BRANCH BRANCH"
			gitlib::pending_prs $RELEASE_BRANCH
		fi

		# No need to pre-check this for mock or staging runs.  Overwriting OK.
		if ((FLAGS_nomock)) && ! ((FLAGS_stage)); then
			common::stepheader "GCS TARGET CHECK"
			# Ensure GCS destinations are clear before continuing
			for v in ${RELEASE_VERSION[@]}; do
				release::gcs::destination_empty \
				 gs://$RELEASE_BUCKET/$BUCKET_TYPE/$v || common::exit 1 "Exiting..."
			done
		fi

		common::stepheader "SESSION VALUES"
		# Show versions and ask for confirmation to continue
		# Pass in the indexed RELEASE_VERSION dict key by key
		ALL_RELEASE_VERSIONS=($(for key in ${!RELEASE_VERSION[@]}; do
														 echo RELEASE_VERSION[$key]; done))

		# Depending on the type of operation being performed one of these will be set
		if [[ -n $BRANCH_POINT ]]; then
			DISPLAY_VERSION="BRANCH_POINT"
		else
			DISPLAY_VERSION="JENKINS_BUILD_VERSION"
		fi
		[[ -n $PARENT_BRANCH ]] && DISPLAY_PARENT_BRANCH="PARENT_BRANCH"

		common::printvars -p WORKDIR WORKDIR TREE_ROOT $DISPLAY_PARENT_BRANCH \
												 $DISPLAY_VERSION \
												 RELEASE_VERSION_PRIME ${ALL_RELEASE_VERSIONS[@]} \
												 RELEASE_BRANCH GCRIO_PATH RELEASE_BUCKET BUCKET_TYPE \
												 CHANGELOG_FILEPATH \
												 FLAGS_nomock FLAGS_rc FLAGS_official \
												 LOGFILE

		if [[ -n "$PARENT_BRANCH" ]]; then
			logecho
			logecho "$ATTENTION: $RELEASE_BRANCH is a NEW branch off $PARENT_BRANCH!"
		fi

		if ! ((FLAGS_nomock)); then
			logecho
			logecho "$ATTENTION: This is a mock (--mock) run." \
							"Publishing will be based on the above values for" \
							"RELEASE_BUCKET, BUCKET_TYPE and GCRIO_PATH."
		fi

		logecho
		((FLAGS_yes)) || common::askyorn -e "Do these values look ok for a release" \
		 || common::exit 1 "Exiting..."

		logecho
		logecho -r "${TPUT[BOLD]}>>>>>>>>${TPUT[OFF]}" \
							 "View detailed session output with:  tailf $LOGFILE"
		logecho -r "${TPUT[BOLD]}>>>>>>>>${TPUT[OFF]}" \
							 "(Previous logs can be found in $LOGFILE.{1..10})"

		common::run_stateful prepare_workspace

		# Store RELEASE_GB key=value in $PROGSTATE
		common::run_stateful --strip-args \
		 "common::disk_space_check $BASEDIR $(($RELEASE_GB*${#RELEASE_VERSION[*]}))" \
		 RELEASE_GB

		# Everything happens in the TREE_ROOT context
		logrun cd $TREE_ROOT

		if ! ((FLAGS_stage)); then
			# Check or store STAGED_LOCATION and STAGED_BUCKET global bools
			# to ensure re-entrant non-stage runs cache this important state
			if ! common::check_state STAGED_LOCATION; then
				 common::check_state -a STAGED_LOCATION STAGED_LOCATION=$STAGED_LOCATION
			fi
			if ! common::check_state STAGED_BUCKET ; then
				common::check_state -a STAGED_BUCKET STAGED_BUCKET=$STAGED_BUCKET
			fi
		fi

		if [[ -z $STAGED_LOCATION ]]; then
			logecho "STAGED_LOCATION:=$STAGED_LOCATION, beginning prebuild"
			# Iterate over session release versions for setup, tagging and building
			for label in "${ORDERED_RELEASE_KEYS[@]}"; do
				common::run_stateful "prepare_tree $label" RELEASE_VERSION[$label]
				# This doesn't have to be done per label, but does need to be inserted
				# after an initial prepare_tree(), so just let the statefulness of it
				# ignore a second iteration/call.
				common::run_stateful local_kube_cross
				# --prebuild for GCB, skip actual builds. Do everything else
				if ! ((FLAGS_prebuild)); then
					# Do only make cross in this case
					common::run_stateful "make_cross $label" RELEASE_VERSION[$label]
				fi
			done

			# Stop here
			if ((FLAGS_prebuild)) || ((FLAGS_buildonly)); then
				logecho
				logecho "--prebuild/--buildonly complete"
				logecho
				logecho "Finish this session with:"
				logecho
				logecho "$PROG $(sed -n 's,^CMDLINE: ,,p' $PROGSTATE)"
				common::exit 0 "Exiting..."
			fi

			# Complete the "build" for the gcb case
			if ((FLAGS_stage)); then
				for label in "${ORDERED_RELEASE_KEYS[@]}"; do
					common::run_stateful "build_tree $label" RELEASE_VERSION[$label]
				done
			fi

			# No release notes for X.Y.Z-beta.0 releases
			[[ -z "$PARENT_BRANCH" ]] && common::run_stateful generate_release_notes
		else
			logecho "STAGED_LOCATION undefined, attempting completion"
			# Force complete for these three stages
			for label in "${ORDERED_RELEASE_KEYS[@]}"; do
				SKIP_STEPS+=(prepare_tree+$label build_tree+$label)
				SKIP_STEPS+=(local_kube_cross make_cross+$label)
			done
			SKIP_STEPS+=(generate_release_notes)
			for entry in ${SKIP_STEPS[*]}; do
				 # Check and add for re-entrancy
				 if ! common::check_state $entry; then
					 logecho "$ATTENTION: Skipping $entry step executed during staging"
					 common::check_state -a $entry
				 fi
			done
		fi

		if ((FLAGS_stage)); then
			common::run_stateful stage_source_tree
		else
			common::run_stateful push_git_objects
		fi

		# Push for each release version of this session
		for label in "${ORDERED_RELEASE_KEYS[@]}"; do
			common::run_stateful "push_all_artifacts $label" RELEASE_VERSION[$label]
		done

		# if --stage, we're done
		if ((FLAGS_stage)); then
			((FLAGS_nomock)) && EXTRA_FLAGS+=("--nomock")
			((FLAGS_official)) && EXTRA_FLAGS+=("--official")
			((FLAGS_rc)) && EXTRA_FLAGS+=("--rc")
			logecho
			logecho "To release this staged build, run:"
			logecho
			logecho -n "$ gcbmgr release ${EXTRA_FLAGS[*]} $RELEASE_BRANCH" \
								 "--buildversion=$JENKINS_BUILD_VERSION"
			logecho
			logecho
			logecho "-OR-"
			logecho
			logecho -n "$ anago ${EXTRA_FLAGS[*]} $RELEASE_BRANCH" \
								 "--buildversion=$JENKINS_BUILD_VERSION"
			logecho
			logecho

			# Move PROGSTATE
			logecho -n "Moving $PROGSTATE to $PROGSTATE.last: "
			logrun -s mv $PROGSTATE $PROGSTATE.last

			# Keep only the last staged build
			# clean up here on success
			# Delete everything that's not this build in the vX.Y. namespace
			# || true to catch non-zero exit when list is empty
			if [[ $JENKINS_BUILD_VERSION =~ (v[0-9]+\.[0-9]+\.) ]]; then
				logecho "Cleaning up old staged builds from" \
								"gs://$RELEASE_BUCKET/$BUCKET_TYPE/${BASH_REMATCH[1]}..."
				$GSUTIL ls -d gs://$RELEASE_BUCKET/$BUCKET_TYPE/${BASH_REMATCH[1]}* |\
					grep -v $JENKINS_BUILD_VERSION |xargs $GSUTIL -mq rm -r || true
			fi

			common::exit 0
		fi

		if [[ -n "$PARENT_BRANCH" ]]; then
			# TODO: This needs the gcb create/update release issue treatment
			common::run_stateful "announce --branch"
		else
			common::run_stateful announce
			common::run_stateful update_github_release
			common::stepheader "ARCHIVE RELEASE ON GS"
			common::runstep archive_release
		fi

		# Move PROGSTATE
		logecho -n "Moving $PROGSTATE to $PROGSTATE.last: "
		logrun -s mv $PROGSTATE $PROGSTATE.last
	*/

	return release.Anago(anagoOpts)
}
