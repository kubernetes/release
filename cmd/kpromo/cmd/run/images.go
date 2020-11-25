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

package run

import (
	"fmt"
	"os"

	guuid "github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/cip/audit"
	reg "k8s.io/release/pkg/cip/dockerregistry"
	"k8s.io/release/pkg/cip/gcloud"
	"k8s.io/release/pkg/cip/stream"
)

var (
	// GitDescribe is stamped by bazel
	// TODO: These vars were stamped by bazel and need to built by other means now
	GitDescribe string

	// GitCommit is stamped by bazel
	GitCommit string

	// TimestampUtcRfc3339 is stamped by bazel
	TimestampUtcRfc3339 string
)

// imagesCmd represents the subcommand for `kpromo run images`
var imagesCmd = &cobra.Command{
	Use:           "images",
	Short:         "Promote images from a staging registry to production",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Wrap(runImagePromotion(newImagesOpts), "run `kpromo run images`")
	},
}

// TODO: Push these into a package
type imagesOpts struct {
	manifest        string
	thinManifestDir string

	keyFiles string

	snapshot    string
	snapshotTag string

	outputFormat string

	snapshotSvcAcct         string
	manifestBasedSnapshotOf string

	auditManifestRepoURL    string
	auditManifestRepoBranch string
	auditManifestPath       string
	auditGcpProjectID       string

	threads int

	maxImageSize int

	severityThreshold int

	jsonLogSummary bool
	parseOnly      bool
	dryRun         bool

	version bool

	minimalSnapshot bool

	useServiceAcct bool

	audit bool
}

// TODO: Push these into a package
var newImagesOpts = &imagesOpts{
	threads: 10,

	outputFormat: "YAML",

	maxImageSize: 2048,

	severityThreshold: -1,
}

func init() {
	// TODO: Move this into a default options function in pkg/promobot
	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.manifest,
		"manifest",
		newImagesOpts.manifest,
		"the manifest file to load",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.thinManifestDir,
		"thin-manifest-dir",
		newImagesOpts.thinManifestDir,
		"recursively read in all manifests within a folder, but all manifests MUST be 'thin' manifests named 'promoter-manifest.yaml', which are like regular manifests but instead of defining the 'images: ...' field directly, the 'imagesPath' field must be defined that points to another YAML file containing the 'images: ...' contents",
	)

	imagesCmd.PersistentFlags().IntVar(
		&newImagesOpts.threads,
		"threads",
		newImagesOpts.threads,
		"number of concurrent goroutines to use when talking to GCR",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.jsonLogSummary,
		"json-log-summary",
		newImagesOpts.jsonLogSummary,
		"only log a json summary of important errors",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.parseOnly,
		"parse-only",
		newImagesOpts.parseOnly,
		"only check that the given manifest file is parseable as a Manifest",
	)

	// TODO: Consider moving this to the root command
	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.dryRun,
		"dry-run",
		newImagesOpts.dryRun,
		"test run promotion without modifying any registry",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.keyFiles,
		"key-files",
		newImagesOpts.keyFiles,
		"CSV of service account key files that must be activated for the promotion (<json-key-file-path>,...)",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.version,
		"version",
		newImagesOpts.version,
		"print version",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.snapshot,
		"snapshot",
		newImagesOpts.snapshot,
		"read all images in a repository and print to stdout",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.snapshotTag,
		"snapshot-tag",
		newImagesOpts.snapshotTag,
		"only snapshot images with the given tag",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.minimalSnapshot,
		"minimal-snapshot",
		newImagesOpts.minimalSnapshot,
		"(only works with -snapshot/-manifest-based-snapshot-of) discard tagless images from snapshot output if they are referenced by a manifest list",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.outputFormat,
		"output-format",
		newImagesOpts.outputFormat,
		"(only works with -snapshot/-manifest-based-snapshot-of) choose output format of the snapshot (default: YAML; allowed values: 'YAML' or 'CSV')",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.snapshotSvcAcct,
		"snapshot-service-account",
		newImagesOpts.snapshotSvcAcct,
		"service account to use for -snapshot",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.manifestBasedSnapshotOf,
		"manifest-based-snapshot-of",
		newImagesOpts.manifestBasedSnapshotOf,
		"read all images in either -manifest or -thin-manifest-dir and print all images that should be promoted to the given registry (assuming the given registry is empty); this is like -snapshot, but instead of reading over the network from a registry, it reads from the local manifests only",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.useServiceAcct,
		"use-service-account",
		newImagesOpts.useServiceAcct,
		"pass '--account=...' to all gcloud calls (default: false)",
	)

	imagesCmd.PersistentFlags().BoolVar(
		&newImagesOpts.audit,
		"audit",
		newImagesOpts.audit,
		"stand up an HTTP server that responds to Pub/Sub push events for auditing",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.auditManifestRepoURL,
		"audit-manifest-repo-url",
		// TODO: Set this in a function instead
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_URL"),
		"https://... address of the repository that holds the promoter manifests",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.auditManifestRepoBranch,
		"audit-manifest-repo-branch",
		// TODO: Set this in a function instead
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_BRANCH"),
		"Git branch to check out (use) for -audit-manifest-repo",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.auditManifestPath,
		"audit-manifest-path",
		// TODO: Set this in a function instead
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_MANIFEST_DIR"),
		"path (relative to the root of -audit-manifest-repo) to the manifests directory",
	)

	imagesCmd.PersistentFlags().StringVar(
		&newImagesOpts.auditGcpProjectID,
		"audit-gcp-project-id",
		// TODO: Set this in a function instead
		os.Getenv("CIP_AUDIT_GCP_PROJECT_ID"),
		"GCP project ID (name); used for labeling error reporting logs to GCP",
	)

	imagesCmd.PersistentFlags().IntVar(
		&newImagesOpts.maxImageSize,
		"max-image-size",
		newImagesOpts.maxImageSize,
		"The maximum image size (MiB) allowed for promotion and must be a positive value (otherwise set to the default value of 2048 MiB)",
	)

	// TODO: Set this in a function instead
	if newImagesOpts.maxImageSize <= 0 {
		newImagesOpts.maxImageSize = 2048
	}

	imagesCmd.PersistentFlags().IntVar(
		&newImagesOpts.severityThreshold,
		"vuln-severity-threshold",
		newImagesOpts.severityThreshold,
		"Using this flag will cause the promoter to only run the vulnerability check. Found vulnerabilities at or above this threshold will result in the vulnerability check failing [severity levels between 0 and 5; 0 - UNSPECIFIED, 1 - MINIMAL, 2 - LOW, 3 - MEDIUM, 4 - HIGH, 5 - CRITICAL]",
	)

	RunCmd.AddCommand(imagesCmd)
}

// nolint: gocyclo
func runImagePromotion(opts *imagesOpts) error {
	if opts.version {
		printVersion()
		return nil
	}

	if err := validateImageOptions(opts); err != nil {
		return errors.Wrap(err, "validating image options")
	}

	if opts.audit {
		uuid := os.Getenv("CIP_AUDIT_TESTCASE_UUID")
		if len(uuid) > 0 {
			logrus.Infof("Starting auditor in Test Mode (%s)", uuid)
		} else {
			uuid = guuid.New().String()
			logrus.Infof("Starting auditor in Regular Mode (%s)", uuid)
		}

		auditServerContext, err := audit.InitRealServerContext(
			opts.auditGcpProjectID,
			opts.auditManifestRepoURL,
			opts.auditManifestRepoBranch,
			opts.auditManifestPath,
			uuid,
		)
		if err != nil {
			return errors.Wrap(err, "creating auditor context")
		}

		auditServerContext.RunAuditor()
	}

	// Activate service accounts.
	if opts.useServiceAcct && opts.keyFiles != "" {
		if err := gcloud.ActivateServiceAccounts(opts.keyFiles); err != nil {
			return errors.Wrap(err, "activating service accounts")
		}
	}

	var (
		mfest       reg.Manifest
		srcRegistry *reg.RegistryContext
		err         error
		mfests      []reg.Manifest
	)

	promotionEdges := make(map[reg.PromotionEdge]interface{})
	sc := reg.SyncContext{}
	mi := make(reg.MasterInventory)

	if opts.manifest == "" && opts.thinManifestDir == "" {
		logrus.Fatal(fmt.Errorf("one of -manifest or -thin-manifest-dir is required"))
	}

	if opts.snapshot != "" {
		srcRegistry = &reg.RegistryContext{
			Name:           reg.RegistryName(opts.snapshot),
			ServiceAccount: opts.snapshotSvcAcct,
			Src:            true,
		}
	} else {
		srcRegistry = &reg.RegistryContext{
			Name:           reg.RegistryName(opts.manifestBasedSnapshotOf),
			ServiceAccount: opts.snapshotSvcAcct,
			Src:            true,
		}
	}

	mfests = []reg.Manifest{
		{
			Registries: []reg.RegistryContext{
				*srcRegistry,
			},
			Images: []reg.Image{},
		},
	}

	doingPromotion := false
	if opts.manifest != "" {
		mfest, err = reg.ParseManifestFromFile(opts.manifest)
		if err != nil {
			logrus.Fatal(err)
		}

		mfests = append(mfests, mfest)
		for _, registry := range mfest.Registries {
			mi[registry.Name] = nil
		}

		sc, err = reg.MakeSyncContext(
			mfests,
			opts.threads,
			opts.dryRun,
			opts.useServiceAcct,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		doingPromotion = true
	} else if opts.thinManifestDir != "" {
		mfests, err = reg.ParseThinManifestsFromDir(opts.thinManifestDir)
		if err != nil {
			return errors.Wrap(err, "parsing thin manifest directory")
		}

		sc, err = reg.MakeSyncContext(
			mfests,
			opts.threads,
			opts.dryRun,
			opts.useServiceAcct)
		if err != nil {
			logrus.Fatal(err)
		}

		doingPromotion = true
	}

	if opts.parseOnly {
		return nil
	}

	// If there are no images in the manifest, it may be a stub manifest file
	// (such as for brand new registries that would be watched by the promoter
	// for the very first time).
	if doingPromotion && opts.manifestBasedSnapshotOf == "" {
		promotionEdges, err = reg.ToPromotionEdges(mfests)
		if err != nil {
			return errors.Wrap(err, "converting list of manifests to edges for promotion")
		}

		imagesInManifests := false
		for _, mfest := range mfests {
			if len(mfest.Images) > 0 {
				imagesInManifests = true
				break
			}
		}
		if !imagesInManifests {
			logrus.Info("No images in manifest(s) --- nothing to do.")
			return nil
		}

		// Print version to make Prow logs more self-explanatory.
		printVersion()

		if opts.severityThreshold >= 0 {
			logrus.Info("********** START (VULN CHECK) **********")
			logrus.Info("DISCLAIMER: Vulnerabilities are found as issues with " +
				"package binaries within image layers, not necessarily " +
				"with the image layers themselves. So a \"fixable\" " +
				"vulnerability may not necessarily be immediately" +
				"actionable. For example, even though a fixed version " +
				"of the binary is available, it doesn't necessarily mean " +
				"that a new version of the image layer is available.")
		} else if opts.dryRun {
			logrus.Info("********** START (DRY RUN) **********")
		} else {
			logrus.Info("********** START **********")
		}
	}

	if len(opts.snapshot) > 0 || len(opts.manifestBasedSnapshotOf) > 0 {
		rii := make(reg.RegInvImage)
		if len(opts.manifestBasedSnapshotOf) > 0 {
			promotionEdges, err = reg.ToPromotionEdges(mfests)
			if err != nil {
				return errors.Wrap(err, "converting list of manifests to edges for promotion")
			}

			rii = reg.EdgesToRegInvImage(
				promotionEdges,
				opts.manifestBasedSnapshotOf,
			)

			if opts.minimalSnapshot {
				sc.ReadRegistries(
					[]reg.RegistryContext{*srcRegistry},
					true,
					reg.MkReadRepositoryCmdReal,
				)

				sc.ReadGCRManifestLists(reg.MkReadManifestListCmdReal)
				rii = sc.RemoveChildDigestEntries(rii)
			}
		} else {
			sc, err = reg.MakeSyncContext(
				mfests,
				opts.threads,
				opts.dryRun,
				opts.useServiceAcct,
			)
			if err != nil {
				logrus.Fatal(err)
			}

			sc.ReadRegistries(
				[]reg.RegistryContext{*srcRegistry},
				// Read all registries recursively, because we want to produce a
				// complete snapshot.
				true,
				reg.MkReadRepositoryCmdReal,
			)

			rii = sc.Inv[mfests[0].Registries[0].Name]
			if opts.snapshotTag != "" {
				rii = reg.FilterByTag(rii, opts.snapshotTag)
			}

			if opts.minimalSnapshot {
				logrus.Info("-minimal-snapshot specified; removing tagless child digests of manifest lists")
				sc.ReadGCRManifestLists(reg.MkReadManifestListCmdReal)
				rii = sc.RemoveChildDigestEntries(rii)
			}
		}

		var snapshot string
		switch opts.outputFormat {
		case "CSV":
			snapshot = rii.ToCSV()
		case "YAML":
			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		default:
			logrus.Errorf("invalid value %s for -output-format; defaulting to YAML", opts.outputFormat)
			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		}

		fmt.Print(snapshot)
		return nil
	}

	if opts.jsonLogSummary {
		defer sc.LogJSONSummary()
	}

	// Check the pull request
	if opts.dryRun {
		err = sc.RunChecks([]reg.PreCheck{})
		if err != nil {
			return errors.Wrap(err, "running prechecks before promotion")
		}
	}

	// Promote.
	mkProducer := func(
		srcRegistry reg.RegistryName,
		srcImageName reg.ImageName,
		destRC reg.RegistryContext,
		imageName reg.ImageName,
		digest reg.Digest, tag reg.Tag, tp reg.TagOp,
	) stream.Producer {
		var sp stream.Subprocess
		sp.CmdInvocation = reg.GetWriteCmd(
			destRC,
			sc.UseServiceAccount,
			srcRegistry,
			srcImageName,
			imageName,
			digest,
			tag,
			tp,
		)

		return &sp
	}

	promotionEdges, ok := sc.FilterPromotionEdges(promotionEdges, true)
	// If any funny business was detected during a comparison of the manifests
	// with the state of the registries, then exit immediately.
	if !ok {
		return errors.New("encountered errors during edge filtering")
	}

	if opts.severityThreshold >= 0 {
		err = sc.RunChecks(
			[]reg.PreCheck{
				reg.MKImageVulnCheck(sc, promotionEdges, opts.severityThreshold, nil),
			},
		)
		if err != nil {
			return errors.Wrap(err, "checking image vulnerabilities")
		}
	} else {
		err = sc.Promote(promotionEdges, mkProducer, nil)
		if err != nil {
			return errors.Wrap(err, "promoting images")
		}
	}

	if opts.severityThreshold >= 0 {
		logrus.Info("********** FINISHED (VULN CHECK) **********")
	} else if opts.dryRun {
		logrus.Info("********** FINISHED (DRY RUN) **********")
	} else {
		logrus.Info("********** FINISHED **********")
	}

	return nil
}

func validateImageOptions(o *imagesOpts) error {
	// TODO: Validate options
	return nil
}

func printVersion() {
	fmt.Printf("Built:   %s\n", TimestampUtcRfc3339)
	fmt.Printf("Version: %s\n", GitDescribe)
	fmt.Printf("Commit:  %s\n", GitCommit)
}
