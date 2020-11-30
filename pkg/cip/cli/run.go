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

package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	reg "k8s.io/release/pkg/cip/dockerregistry"
	"k8s.io/release/pkg/cip/gcloud"
	"k8s.io/release/pkg/cip/stream"
)

type RunOptions struct {
	Manifest                string
	ThinManifestDir         string
	KeyFiles                string
	Snapshot                string
	SnapshotTag             string
	OutputFormat            string
	SnapshotSvcAcct         string
	ManifestBasedSnapshotOf string
	Threads                 int
	MaxImageSize            int
	SeverityThreshold       int
	NoDryRun                bool
	JSONLogSummary          bool
	ParseOnly               bool
	MinimalSnapshot         bool
	UseServiceAcct          bool
}

const (
	PromoterDefaultThreads           = 10
	PromoterDefaultOutputFormat      = "yaml"
	PromoterDefaultMaxImageSize      = 2048
	PromoterDefaultSeverityThreshold = -1

	// flags.
	PromoterManifestFlag                = "manifest"
	PromoterThinManifestDirFlag         = "thin-manifest-dir"
	PromoterSnapshotFlag                = "snapshot"
	PromoterManifestBasedSnapshotOfFlag = "manifest-based-snapshot-of"
	PromoterOutputFlag                  = "output"
)

var PromoterAllowedOutputFormats = []string{
	"csv",
	"yaml",
}

// TODO: Function 'runPromoteCmd' has too many statements (97 > 40) (funlen)
// nolint: funlen,gocognit,gocyclo
func RunPromoteCmd(opts *RunOptions) error {
	if err := validateImageOptions(opts); err != nil {
		return errors.Wrap(err, "validating image options")
	}

	// Activate service accounts.
	if opts.UseServiceAcct && opts.KeyFiles != "" {
		if err := gcloud.ActivateServiceAccounts(opts.KeyFiles); err != nil {
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

	// TODO: Move this into the validation function
	if opts.Snapshot != "" || opts.ManifestBasedSnapshotOf != "" {
		if opts.Snapshot != "" {
			srcRegistry = &reg.RegistryContext{
				Name:           reg.RegistryName(opts.Snapshot),
				ServiceAccount: opts.SnapshotSvcAcct,
				Src:            true,
			}
		} else {
			srcRegistry = &reg.RegistryContext{
				Name:           reg.RegistryName(opts.ManifestBasedSnapshotOf),
				ServiceAccount: opts.SnapshotSvcAcct,
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
		// TODO: Move this into the validation function
	} else if opts.Manifest == "" && opts.ThinManifestDir == "" {
		logrus.Fatalf(
			"either %s or %s flag is required",
			PromoterManifestFlag,
			PromoterThinManifestDirFlag,
		)
	}

	doingPromotion := false

	// TODO: is deeply nested (complexity: 5) (nestif)
	// nolint: nestif
	if opts.Manifest != "" {
		mfest, err = reg.ParseManifestFromFile(opts.Manifest)
		if err != nil {
			logrus.Fatal(err)
		}

		mfests = append(mfests, mfest)
		for _, registry := range mfest.Registries {
			mi[registry.Name] = nil
		}

		sc, err = reg.MakeSyncContext(
			mfests,
			opts.Threads,
			opts.NoDryRun,
			opts.UseServiceAcct,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		doingPromotion = true
	} else if opts.ThinManifestDir != "" {
		mfests, err = reg.ParseThinManifestsFromDir(opts.ThinManifestDir)
		if err != nil {
			return errors.Wrap(err, "parsing thin manifest directory")
		}

		sc, err = reg.MakeSyncContext(
			mfests,
			opts.Threads,
			opts.NoDryRun,
			opts.UseServiceAcct,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		doingPromotion = true
	}

	if opts.ParseOnly {
		return nil
	}

	// If there are no images in the manifest, it may be a stub manifest file
	// (such as for brand new registries that would be watched by the promoter
	// for the very first time).
	// TODO: is deeply nested (complexity: 6) (nestif)
	// nolint: nestif
	if doingPromotion && opts.ManifestBasedSnapshotOf == "" {
		promotionEdges, err = reg.ToPromotionEdges(mfests)
		if err != nil {
			return errors.Wrap(
				err,
				"converting list of manifests to edges for promotion",
			)
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

		// nolint: gocritic
		if opts.SeverityThreshold >= 0 {
			logrus.Info("********** START (VULN CHECK) **********")
			logrus.Info(
				`DISCLAIMER: Vulnerabilities are found as issues with package
binaries within image layers, not necessarily with the image layers themselves.
So a 'fixable' vulnerability may not necessarily be immediately actionable. For
example, even though a fixed version of the binary is available, it doesn't
necessarily mean that a new version of the image layer is available.`,
			)
		} else if !opts.NoDryRun {
			logrus.Info("********** START (DRY RUN) **********")
		} else {
			logrus.Info("********** START **********")
		}
	}

	// TODO: is deeply nested (complexity: 12) (nestif)
	// nolint: nestif
	if len(opts.Snapshot) > 0 || len(opts.ManifestBasedSnapshotOf) > 0 {
		rii := make(reg.RegInvImage)
		if len(opts.ManifestBasedSnapshotOf) > 0 {
			promotionEdges, err = reg.ToPromotionEdges(mfests)
			if err != nil {
				return errors.Wrap(
					err,
					"converting list of manifests to edges for promotion",
				)
			}

			rii = reg.EdgesToRegInvImage(
				promotionEdges,
				opts.ManifestBasedSnapshotOf,
			)

			if opts.MinimalSnapshot {
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
				opts.Threads,
				opts.NoDryRun,
				opts.UseServiceAcct,
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
			if opts.SnapshotTag != "" {
				rii = reg.FilterByTag(rii, opts.SnapshotTag)
			}

			if opts.MinimalSnapshot {
				logrus.Info("removing tagless child digests of manifest lists")
				sc.ReadGCRManifestLists(reg.MkReadManifestListCmdReal)
				rii = sc.RemoveChildDigestEntries(rii)
			}
		}

		var snapshot string
		switch strings.ToLower(opts.OutputFormat) {
		case "csv":
			snapshot = rii.ToCSV()
		case "yaml":
			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		default:
			logrus.Errorf(
				"invalid value %s for '--%s'; defaulting to %s",
				opts.OutputFormat,
				PromoterOutputFlag,
				PromoterDefaultOutputFormat,
			)

			snapshot = rii.ToYAML(reg.YamlMarshalingOpts{})
		}

		fmt.Print(snapshot)
		return nil
	}

	if opts.JSONLogSummary {
		defer sc.LogJSONSummary()
	}

	// Check the pull request
	if !opts.NoDryRun {
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

	if opts.SeverityThreshold >= 0 {
		err = sc.RunChecks(
			[]reg.PreCheck{
				reg.MKImageVulnCheck(
					sc,
					promotionEdges,
					opts.SeverityThreshold,
					nil,
				),
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

	// nolint: gocritic
	if opts.SeverityThreshold >= 0 {
		logrus.Info("********** FINISHED (VULN CHECK) **********")
	} else if !opts.NoDryRun {
		logrus.Info("********** FINISHED (DRY RUN) **********")
	} else {
		logrus.Info("********** FINISHED **********")
	}

	return nil
}

// nolint: unused
func validateImageOptions(o *RunOptions) error {
	// TODO: Validate options
	return nil
}
