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
	"flag"
	"fmt"
	"os"

	guuid "github.com/google/uuid"

	"k8s.io/klog"
	"k8s.io/release/pkg/cip/audit"
	reg "k8s.io/release/pkg/cip/dockerregistry"
	"k8s.io/release/pkg/cip/gcloud"
	"k8s.io/release/pkg/cip/stream"
)

// GitDescribe is stamped by bazel.
var GitDescribe string

// GitCommit is stamped by bazel.
var GitCommit string

// TimestampUtcRfc3339 is stamped by bazel.
var TimestampUtcRfc3339 string

// nolint[gocyclo]
func main() {
	// klog uses the "v" flag in order to set the verbosity level
	klog.InitFlags(nil)

	manifestPtr := flag.String(
		"manifest", "", "the manifest file to load")
	thinManifestDirPtr := flag.String(
		"thin-manifest-dir",
		"",
		"recursively read in all manifests within a folder, but all manifests MUST be 'thin' manifests named 'promoter-manifest.yaml', which are like regular manifests but instead of defining the 'images: ...' field directly, the 'imagesPath' field must be defined that points to another YAML file containing the 'images: ...' contents")
	threadsPtr := flag.Int(
		"threads",
		10, "number of concurrent goroutines to use when talking to GCR")
	jsonLogSummaryPtr := flag.Bool(
		"json-log-summary",
		false,
		"only log a json summary of important errors")
	parseOnlyPtr := flag.Bool(
		"parse-only",
		false,
		"only check that the given manifest file is parseable as a Manifest"+
			" (default: false)")
	dryRunPtr := flag.Bool(
		"dry-run",
		true,
		"print what would have happened by running this tool;"+
			" do not actually modify any registry")
	keyFilesPtr := flag.String(
		"key-files",
		"",
		"CSV of service account key files that must be activated for the promotion (<json-key-file-path>,...)")
	// Add in help flag information, because Go's "flag" package automatically
	// adds it, but for whatever reason does not show it as part of available
	// options.
	helpPtr := flag.Bool(
		"help",
		false,
		"print help")
	versionPtr := flag.Bool(
		"version",
		false,
		"print version")
	snapshotPtr := flag.String(
		"snapshot",
		"",
		"read all images in a repository and print to stdout")
	snapshotTag := ""
	flag.StringVar(&snapshotTag, "snapshot-tag", snapshotTag, "only snapshot images with the given tag")
	minimalSnapshotPtr := flag.Bool(
		"minimal-snapshot",
		false,
		"(only works with -snapshot/-manifest-based-snapshot-of) discard tagless images from snapshot output if they are referenced by a manifest list")
	outputFormatPtr := flag.String(
		"output-format",
		"YAML",
		"(only works with -snapshot/-manifest-based-snapshot-of) choose output format of the snapshot (default: YAML; allowed values: 'YAML' or 'CSV')")
	snapshotSvcAccPtr := flag.String(
		"snapshot-service-account",
		"",
		"service account to use for -snapshot")
	manifestBasedSnapshotOf := flag.String(
		"manifest-based-snapshot-of",
		"",
		"read all images in either -manifest or -thin-manifest-dir and print all images that should be promoted to the given registry (assuming the given registry is empty); this is like -snapshot, but instead of reading over the network from a registry, it reads from the local manifests only")
	useServiceAccount := false
	flag.BoolVar(&useServiceAccount, "use-service-account", false,
		"pass '--account=...' to all gcloud calls (default: false)")
	auditorPtr := flag.Bool(
		"audit",
		false,
		"stand up an HTTP server that responds to Pub/Sub push events for auditing")
	auditManifestRepoUrlPtr := flag.String(
		"audit-manifest-repo-url",
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_URL"),
		"https://... address of the repository that holds the promoter manifests")
	auditManifestRepoBranchPtr := flag.String(
		"audit-manifest-repo-branch",
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_BRANCH"),
		"Git branch to check out (use) for -audit-manifest-repo")
	auditManifestPathPtr := flag.String(
		"audit-manifest-path",
		os.Getenv("CIP_AUDIT_MANIFEST_REPO_MANIFEST_DIR"),
		"path (relative to the root of -audit-manifest-repo) to the manifests directory")
	auditGcpProjectID := flag.String(
		"audit-gcp-project-id",
		os.Getenv("CIP_AUDIT_GCP_PROJECT_ID"),
		"GCP project ID (name); used for labeling error reporting logs to GCP")
	maxImageSizePtr := flag.Int(
		"max-image-size",
		2048,
		"The maximum image size (MiB) allowed for promotion and must be a positive value (othwerise set to the default value of 2048 MiB)")
	if *maxImageSizePtr <= 0 {
		*maxImageSizePtr = 2048
	}
	severityThresholdPtr := flag.Int(
		"vuln-severity-threshold",
		-1,
		"Using this flag will cause the promoter to only run the vulnerability check. Found vulnerabilities at or above this threshold will result in the vulnerability check failing [severity levels beteen 0 and 5; 0 - UNSPECIFIED, 1 - MINIMAL, 2 - LOW, 3 - MEDIUM, 4 - HIGH, 5 - CRITICAL]")
	flag.Parse()

	if len(os.Args) == 1 {
		printVersion()
		printUsage()
		os.Exit(0)
	}

	if *helpPtr {
		printUsage()
		os.Exit(0)
	}

	if *versionPtr {
		printVersion()
		os.Exit(0)
	}

	if *auditorPtr {
		uuid := os.Getenv("CIP_AUDIT_TESTCASE_UUID")
		if len(uuid) > 0 {
			klog.Infof("Starting auditor in Test Mode (%s)", uuid)
		} else {
			uuid = guuid.New().String()
			klog.Infof("Starting auditor in Regular Mode (%s)", uuid)
		}

		auditServerContext, err := audit.InitRealServerContext(
			*auditGcpProjectID,
			*auditManifestRepoUrlPtr,
			*auditManifestRepoBranchPtr,
			*auditManifestPathPtr,
			uuid)
		if err != nil {
			klog.Exitln(err)
		}
		auditServerContext.RunAuditor()
	}

	// Activate service accounts.
	if useServiceAccount && len(*keyFilesPtr) > 0 {
		if err := gcloud.ActivateServiceAccounts(*keyFilesPtr); err != nil {
			klog.Exitln(err)
		}
	}

	var mfest reg.Manifest
	var srcRegistry *reg.RegistryContext
	var err error
	var mfests []reg.Manifest
	promotionEdges := make(map[reg.PromotionEdge]interface{})
	sc := reg.SyncContext{}
	mi := make(reg.MasterInventory)

	if len(*snapshotPtr) > 0 || len(*manifestBasedSnapshotOf) > 0 {
		if len(*snapshotPtr) > 0 {
			srcRegistry = &reg.RegistryContext{
				Name:           reg.RegistryName(*snapshotPtr),
				ServiceAccount: *snapshotSvcAccPtr,
				Src:            true,
			}
		} else {
			srcRegistry = &reg.RegistryContext{
				Name:           reg.RegistryName(*manifestBasedSnapshotOf),
				ServiceAccount: *snapshotSvcAccPtr,
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
	} else {
		if *manifestPtr == "" && *thinManifestDirPtr == "" {
			klog.Fatal(fmt.Errorf("one of -manifest or -thin-manifest-dir is required"))
		}
	}

	doingPromotion := false
	if *manifestPtr != "" {
		mfest, err = reg.ParseManifestFromFile(*manifestPtr)
		if err != nil {
			klog.Fatal(err)
		}
		mfests = append(mfests, mfest)
		for _, registry := range mfest.Registries {
			mi[registry.Name] = nil
		}
		sc, err = reg.MakeSyncContext(
			mfests,
			*threadsPtr,
			*dryRunPtr,
			useServiceAccount)
		if err != nil {
			klog.Fatal(err)
		}
		doingPromotion = true
	} else if *thinManifestDirPtr != "" {
		mfests, err = reg.ParseThinManifestsFromDir(*thinManifestDirPtr)
		if err != nil {
			klog.Exitln(err)
		}

		sc, err = reg.MakeSyncContext(
			mfests,
			*threadsPtr,
			*dryRunPtr,
			useServiceAccount)
		if err != nil {
			klog.Fatal(err)
		}
		doingPromotion = true
	}

	if *parseOnlyPtr {
		os.Exit(0)
	}

	// If there are no images in the manifest, it may be a stub manifest file
	// (such as for brand new registries that would be watched by the promoter
	// for the very first time).
	if doingPromotion && len(*manifestBasedSnapshotOf) == 0 {
		promotionEdges, err = reg.ToPromotionEdges(mfests)
		if err != nil {
			klog.Exitln(err)
		}

		imagesInManifests := false
		for _, mfest := range mfests {
			if len(mfest.Images) > 0 {
				imagesInManifests = true
				break
			}
		}
		if !imagesInManifests {
			klog.Info("No images in manifest(s) --- nothing to do.")
			os.Exit(0)
		}

		// Print version to make Prow logs more self-explanatory.
		printVersion()

		if *severityThresholdPtr >= 0 {
			klog.Info("********** START (VULN CHECK) **********")
			klog.Info("DISCLAIMER: Vulnerabilities are found as issues with " +
				"package binaries within image layers, not necessarily " +
				"with the image layers themselves. So a \"fixable\" " +
				"vulnerability may not necessarily be immediately" +
				"actionable. For example, even though a fixed version " +
				"of the binary is available, it doesn't necessarily mean " +
				"that a new version of the image layer is available.")
		} else if *dryRunPtr {
			klog.Info("********** START (DRY RUN) **********")
		} else {
			klog.Info("********** START **********")
		}
	}

	if len(*snapshotPtr) > 0 || len(*manifestBasedSnapshotOf) > 0 {
		rii := make(reg.RegInvImage)
		if len(*manifestBasedSnapshotOf) > 0 {
			promotionEdges, err = reg.ToPromotionEdges(mfests)
			if err != nil {
				klog.Exitln(err)
			}
			rii = reg.EdgesToRegInvImage(promotionEdges,
				*manifestBasedSnapshotOf)

			if *minimalSnapshotPtr {
				sc.ReadRegistries(
					[]reg.RegistryContext{*srcRegistry},
					true,
					reg.MkReadRepositoryCmdReal)
				sc.ReadGCRManifestLists(reg.MkReadManifestListCmdReal)
				rii = sc.RemoveChildDigestEntries(rii)
			}
		} else {
			sc, err = reg.MakeSyncContext(
				mfests,
				*threadsPtr,
				*dryRunPtr,
				useServiceAccount)
			if err != nil {
				klog.Fatal(err)
			}
			sc.ReadRegistries(
				[]reg.RegistryContext{*srcRegistry},
				// Read all registries recursively, because we want to produce a
				// complete snapshot.
				true,
				reg.MkReadRepositoryCmdReal)

			rii = sc.Inv[mfests[0].Registries[0].Name]
			if snapshotTag != "" {
				rii = reg.FilterByTag(rii, snapshotTag)
			}
			if *minimalSnapshotPtr {
				klog.Info("-minimal-snapshot specifed; removing tagless child digests of manifest lists")
				sc.ReadGCRManifestLists(reg.MkReadManifestListCmdReal)
				rii = sc.RemoveChildDigestEntries(rii)
			}
		}

		var snapshot string
		switch *outputFormatPtr {
		case "CSV":
			snapshot = rii.ToCSV()
		case "YAML":
			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		default:
			klog.Errorf("invalid value %s for -output-format; defaulting to YAML", *outputFormatPtr)
			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		}
		fmt.Print(snapshot)
		os.Exit(0)
	}

	if *jsonLogSummaryPtr {
		defer sc.LogJSONSummary()
	}

	// Check the pull request
	if *dryRunPtr {
		err = sc.RunChecks([]reg.PreCheck{})
		if err != nil {
			klog.Exitln(err)
		}
	}

	// Promote.
	mkProducer := func(
		srcRegistry reg.RegistryName,
		srcImageName reg.ImageName,
		destRC reg.RegistryContext,
		imageName reg.ImageName,
		digest reg.Digest, tag reg.Tag, tp reg.TagOp) stream.Producer {
		var sp stream.Subprocess
		sp.CmdInvocation = reg.GetWriteCmd(
			destRC,
			sc.UseServiceAccount,
			srcRegistry,
			srcImageName,
			imageName,
			digest,
			tag,
			tp)
		return &sp
	}
	promotionEdges, ok := sc.FilterPromotionEdges(promotionEdges, true)
	// If any funny business was detected during a comparison of the manifests
	// with the state of the registries, then exit immediately.
	if !ok {
		klog.Exitln("encountered errors during edge filtering")
	}

	if *severityThresholdPtr >= 0 {
		err = sc.RunChecks([]reg.PreCheck{
			reg.MKImageVulnCheck(sc, promotionEdges,
				*severityThresholdPtr, nil),
		})
		if err != nil {
			klog.Exitln(err)
		}
	} else {
		err = sc.Promote(promotionEdges, mkProducer, nil)
		if err != nil {
			klog.Exitln(err)
		}
	}

	if *severityThresholdPtr >= 0 {
		klog.Info("********** FINISHED (VULN CHECK) **********")
	} else if *dryRunPtr {
		klog.Info("********** FINISHED (DRY RUN) **********")
	} else {
		klog.Info("********** FINISHED **********")
	}
}

func printVersion() {
	fmt.Printf("Built:   %s\n", TimestampUtcRfc3339)
	fmt.Printf("Version: %s\n", GitDescribe)
	fmt.Printf("Commit:  %s\n", GitCommit)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
