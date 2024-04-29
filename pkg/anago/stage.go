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

package anago

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	gogit "github.com/go-git/go-git/v5"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/bom/pkg/provenance"
	"sigs.k8s.io/bom/pkg/spdx"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/log"
)

// stageClient is a client for staging releases.
//
//counterfeiter:generate . stageClient
type stageClient interface {
	// Submit can be used to submit a Google Cloud Build (GCB) job.
	Submit(stream bool) error

	// InitState initializes the default internal state.
	InitState()

	// InitLogFile sets up the log file target.
	InitLogFile() error

	// Validate if the provided `StageOptions` are correctly set.
	ValidateOptions() error

	// CheckPrerequisites verifies that a valid GITHUB_TOKEN environment
	// variable is set. It also checks for the existence and version of
	// required packages and if the correct Google Cloud project is set. A
	// basic hardware check will ensure that enough disk space is available,
	// too.
	CheckPrerequisites() error

	// CheckReleaseBranchState discovers if the provided release branch has to
	// be created.
	CheckReleaseBranchState() error

	// GenerateReleaseVersion discovers the next versions to be released.
	GenerateReleaseVersion() error

	// PrepareWorkspace verifies that the working directory is in the desired
	// state. This means that the build directory is cleaned up and the checked
	// out repository is in a clean state.
	PrepareWorkspace() error

	// TagRepository creates all necessary git objects by tagging the
	// repository for the provided `versions` the main version `versionPrime`
	// and the `parentBranch`.
	TagRepository() error

	// Build runs 'make cross-in-a-container' by using the latest kubecross
	// container image. This step also build all necessary release tarballs.
	Build() error

	// GenerateChangelog builds the CHANGELOG-x.y.md file and commits it
	// into the local repository.
	GenerateChangelog() error

	// VerifyArtifacts performs verification of the generated artifacts
	VerifyArtifacts() error

	// GenerateBillOfMaterials generates the SBOM documents for the Kubernetes
	// source code and the release artifacts.
	GenerateBillOfMaterials() error

	// StageArtifacts copies the build artifacts to a Google Cloud Bucket.
	StageArtifacts() error
}

// DefaultStage is the default staging implementation used in production.
type DefaultStage struct {
	impl    stageImpl
	options *StageOptions
	state   *StageState
}

// NewDefaultStage creates a new defaultStage instance.
func NewDefaultStage(options *StageOptions) *DefaultStage {
	return &DefaultStage{&defaultStageImpl{}, options, nil}
}

// SetImpl can be used to set the internal stage implementation.
func (d *DefaultStage) SetImpl(impl stageImpl) {
	d.impl = impl
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (d *DefaultStage) SetState(state *StageState) {
	d.state = state
}

// State returns the internal state.
func (d *DefaultStage) State() *StageState {
	return d.state
}

// defaultStageImpl is the default internal stage client implementation.
type defaultStageImpl struct{}

// stageImpl is the implementation of the stage client.
//
//counterfeiter:generate . stageImpl
type stageImpl interface {
	Submit(options *gcb.Options) error
	ToFile(fileName string) error
	CheckPrerequisites() error
	BranchNeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (bool, error)
	PrepareWorkspaceStage(noMock bool) error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	OpenRepo(repoPath string) (*git.Repo, error)
	RevParse(repo *git.Repo, rev string) (string, error)
	RevParseTag(repo *git.Repo, rev string) (string, error)
	Checkout(repo *git.Repo, rev string, args ...string) error
	CurrentBranch(repo *git.Repo) (string, error)
	CommitEmpty(repo *git.Repo, msg string) error
	Tag(repo *git.Repo, name, message string) error
	Merge(repo *git.Repo, rev string) error
	CheckReleaseBucket(options *build.Options) error
	DockerHubLogin() error
	MakeCross(version string) error
	GenerateChangelog(options *changelog.Options) error
	StageLocalSourceTree(
		options *build.Options, workDir, buildVersion string,
	) error
	DeleteLocalSourceTarball(*build.Options, string) error
	StageLocalArtifacts(options *build.Options) error
	PushReleaseArtifacts(
		options *build.Options, srcPath, gcsPath string,
	) error
	PushContainerImages(options *build.Options) error
	GenerateVersionArtifactsBOM(string) error
	GenerateSourceTreeBOM(options *spdx.DocGenerateOptions) (*spdx.Document, error)
	WriteSourceBOM(spdxDoc *spdx.Document, version string) error
	ListBinaries(version string) ([]struct{ Path, Platform, Arch string }, error)
	ListImageArchives(string) ([]string, error)
	ListTarballs(version string) ([]string, error)
	BuildBaseArtifactsSBOM(*spdx.DocGenerateOptions) (*spdx.Document, error)
	AddBinariesToSBOM(*spdx.Document, string) error
	AddTarfilesToSBOM(*spdx.Document, string) error
	VerifyArtifacts([]string) error
	GenerateAttestation(*StageState, *StageOptions) (*provenance.Statement, error)
	PushAttestation(*provenance.Statement, *StageOptions) error
	GetProvenanceSubjects(*StageOptions, string) ([]intoto.Subject, error)
	GetOutputDirSubjects(*StageOptions, string, string) ([]intoto.Subject, error)
}

func (d *defaultStageImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (d *defaultStageImpl) ToFile(fileName string) error {
	return log.ToFile(fileName)
}

func (d *defaultStageImpl) CheckPrerequisites() error {
	return release.NewPrerequisitesChecker().Run(workspaceDir)
}

func (d *defaultStageImpl) BranchNeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (bool, error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
}

func (d *defaultStageImpl) PrepareWorkspaceStage(noMock bool) error {
	if err := release.PrepareWorkspaceStage(gitRoot, noMock); err != nil {
		return err
	}
	return os.Chdir(gitRoot)
}

func (d *defaultStageImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

func (d *defaultStageImpl) OpenRepo(repoPath string) (*git.Repo, error) {
	return git.OpenRepo(repoPath)
}

func (d *defaultStageImpl) RevParse(repo *git.Repo, rev string) (string, error) {
	return repo.RevParse(rev)
}

func (d *defaultStageImpl) RevParseTag(repo *git.Repo, rev string) (string, error) {
	return repo.RevParseTag(rev)
}

func (d *defaultStageImpl) Checkout(repo *git.Repo, rev string, args ...string) error {
	return repo.Checkout(rev, args...)
}

func (d *defaultStageImpl) CurrentBranch(repo *git.Repo) (string, error) {
	return repo.CurrentBranch()
}

func (d *defaultStageImpl) CommitEmpty(repo *git.Repo, msg string) error {
	return repo.CommitEmpty(msg)
}

func (d *defaultStageImpl) Tag(repo *git.Repo, name, message string) error {
	return repo.Tag(name, message)
}

func (d *defaultStageImpl) Merge(repo *git.Repo, rev string) error {
	return repo.Merge(rev)
}

func (d *defaultStageImpl) MakeCross(version string) error {
	return build.NewMake().MakeCross(version)
}

func (d *defaultStageImpl) DockerHubLogin() error {
	return release.DockerHubLogin()
}

func (d *defaultStageImpl) GenerateChangelog(options *changelog.Options) error {
	return changelog.New(options).Run()
}

func (d *defaultStageImpl) CheckReleaseBucket(
	options *build.Options,
) error {
	return build.NewInstance(options).CheckReleaseBucket()
}

func (d *defaultStageImpl) StageLocalSourceTree(
	options *build.Options, workDir, buildVersion string,
) error {
	return build.NewInstance(options).StageLocalSourceTree(workDir, buildVersion)
}

func (d *defaultStageImpl) DeleteLocalSourceTarball(options *build.Options, workDir string) error {
	return build.NewInstance(options).DeleteLocalSourceTarball(workDir)
}

func (d *defaultStageImpl) StageLocalArtifacts(
	options *build.Options,
) error {
	return build.NewInstance(options).StageLocalArtifacts()
}

func (d *defaultStageImpl) PushReleaseArtifacts(
	options *build.Options, srcPath, gcsPath string,
) error {
	return build.NewInstance(options).PushReleaseArtifacts(srcPath, gcsPath)
}

func (d *defaultStageImpl) PushContainerImages(
	options *build.Options,
) error {
	return build.NewInstance(options).PushContainerImages()
}

func (d *DefaultStage) Submit(stream bool) error {
	options := gcb.NewDefaultOptions()
	options.Stream = stream
	options.Stage = true
	options.NoMock = d.options.NoMock
	options.Branch = d.options.ReleaseBranch
	options.ReleaseType = d.options.ReleaseType
	return d.impl.Submit(options)
}

// ListBinaries returns a list of all the binaries obtained
// from the build with platform and arch details
func (d *defaultStageImpl) ListBinaries(version string) (list []struct{ Path, Platform, Arch string }, err error) {
	return release.ListBuildBinaries(gitRoot, version)
}

// ListImageArchives returns a list of the image archives produced
// fior the specified version
func (d *defaultStageImpl) ListImageArchives(version string) ([]string, error) {
	return release.ListBuildImages(gitRoot, version)
}

// ListTarballs returns the produced tarballs produced for this version
func (d *defaultStageImpl) ListTarballs(version string) ([]string, error) {
	return release.ListBuildTarballs(gitRoot, version)
}

// VerifyArtifacts check the artifacts produced are correct
func (d *defaultStageImpl) VerifyArtifacts(versions []string) error {
	// Create a new artifact checker to verify the consistency of
	// the produced artifacts.
	checker := release.NewArtifactCheckerWithOptions(
		&release.ArtifactCheckerOptions{
			GitRoot:  gitRoot,
			Versions: versions,
		},
	)

	// Ensure binaries are of the correct architecture
	if err := checker.CheckBinaryArchitectures(); err != nil {
		return fmt.Errorf("checking binary architectures: %w", err)
	}

	return nil
}

func (d *DefaultStage) InitLogFile() error {
	logrus.SetFormatter(
		&logrus.TextFormatter{FullTimestamp: true, ForceColors: true},
	)
	logFile := filepath.Join(os.TempDir(), "stage.log")
	if err := d.impl.ToFile(logFile); err != nil {
		return fmt.Errorf("setup log file: %w", err)
	}
	d.state.logFile = logFile
	logrus.Infof("Additionally logging to file %s", d.state.logFile)
	return nil
}

func (d *DefaultStage) InitState() {
	d.state = &StageState{DefaultState()}
}

func (d *DefaultStage) ValidateOptions() error {
	if err := d.options.Validate(d.state.State); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}
	return nil
}

func (d *DefaultStage) CheckPrerequisites() error {
	return d.impl.CheckPrerequisites()
}

func (d *DefaultStage) CheckReleaseBranchState() error {
	createReleaseBranch, err := d.impl.BranchNeedsCreation(
		d.options.ReleaseBranch,
		d.options.ReleaseType,
		d.state.semverBuildVersion,
	)
	if err != nil {
		return fmt.Errorf("check if release branch needs creation: %w", err)
	}
	d.state.createReleaseBranch = createReleaseBranch
	return nil
}

func (d *DefaultStage) GenerateReleaseVersion() error {
	versions, err := d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		d.state.createReleaseBranch,
	)
	if err != nil {
		return fmt.Errorf("generating release versions for stage: %w", err)
	}
	// Set the versions on the state
	d.state.versions = versions
	return nil
}

func (d *DefaultStage) PrepareWorkspace() error {
	if err := d.impl.PrepareWorkspaceStage(d.options.NoMock); err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}
	return nil
}

func (d *DefaultStage) TagRepository() error {
	repo, err := d.impl.OpenRepo(gitRoot)
	if err != nil {
		return fmt.Errorf("open Kubernetes repository: %w", err)
	}

	for _, version := range d.state.versions.Ordered() {
		logrus.Infof("Preparing version %s", version)

		// Ensure that the tag not already exists
		if _, err := d.impl.RevParseTag(repo, version); err == nil {
			return fmt.Errorf("tag %s already exists: %w", version, err)
		}

		// Usually the build version contains a commit we can reference. If
		// not, because the build version is exactly a tag, then we fallback to
		// that tag.
		commit := d.options.BuildVersion
		if len(d.state.semverBuildVersion.Build) > 0 {
			commit = d.state.semverBuildVersion.Build[0]
		}

		if d.state.createReleaseBranch {
			logrus.Infof("Creating release branch %s", d.options.ReleaseBranch)

			if version == d.state.versions.Prime() {
				logrus.Infof("Version %s is the prime version", version)
				logrus.Infof(
					"Creating release branch %s from commit %s",
					d.options.ReleaseBranch, commit,
				)
				if err := d.impl.Checkout(
					repo, "-b", d.options.ReleaseBranch, commit,
				); err != nil {
					return fmt.Errorf("create new release branch: %w", err)
				}
			} else {
				logrus.Infof(
					"Version %s is not the prime, checking out %s branch",
					version, git.DefaultBranch,
				)
				if err := d.impl.Checkout(repo, git.DefaultBranch); err != nil {
					return fmt.Errorf("checkout %s branch: %w", git.DefaultBranch, err)
				}
			}
		} else {
			logrus.Infof("Checking out branch %s", d.options.ReleaseBranch)
			if err := d.impl.Checkout(repo, d.options.ReleaseBranch); err != nil {
				return fmt.Errorf("checking out branch %s: %w", d.options.ReleaseBranch, err)
			}
		}

		// `branch == ""` in case we checked out a commit directly, which is
		// then in detached head state.
		branch, err := d.impl.CurrentBranch(repo)
		if err != nil {
			return fmt.Errorf("get current branch: %w", err)
		}
		if branch != "" {
			logrus.Infof("Current branch is %s", branch)
		}

		// For release branches, we create an empty release commit to avoid
		// potential ambiguous 'git describe' logic between the official
		// release, 'x.y.z' and the next beta of that release branch,
		// 'x.y.(z+1)-beta.0'.
		//
		// We avoid doing this empty release commit on 'master', as:
		//   - there is a potential for branch conflicts as upstream/master
		//     moves ahead
		//   - we're checking out a git ref, as opposed to a branch, which
		//     means the tag will detached from 'upstream/master'
		//
		// A side-effect of the tag being detached from 'master' is the primary
		// build job (ci-kubernetes-build) will build as the previous alpha,
		// instead of the assumed tag. This causes the next anago run against
		// 'master' to fail due to an old build version.
		//
		// Example: 'v1.18.0-alpha.2.663+df908c3aad70be'
		//          (should instead be:
		//			 'v1.18.0-alpha.3.<commits-since-tag>+<commit-ish>')
		//
		// ref:
		//   - https://github.com/kubernetes/release/issues/1020
		//   - https://github.com/kubernetes/release/pull/1030
		//   - https://github.com/kubernetes/release/issues/1080
		//   - https://github.com/kubernetes/kubernetes/pull/88074

		// When tagging a release branch, always create an empty commit:
		if strings.HasPrefix(branch, "release-") {
			logrus.Infof("Creating empty release commit for tag %s", version)
			if err := d.impl.CommitEmpty(
				repo,
				"Release commit for Kubernetes "+version,
			); err != nil {
				return fmt.Errorf("create empty release commit: %w", err)
			}
		}

		// If we are on master/main we do not create an empty commit,
		// but we detach the head at the specified commit to avoid having
		// commits merged between the BuildVersion commit and the tag:
		detachHead := release.IsDefaultK8sUpstream() &&
			branch != "" &&
			!strings.HasPrefix(branch, "release-")
		if detachHead {
			logrus.Infof("Detaching HEAD at commit %s to create tag %s", commit, version)
			if err := d.impl.Checkout(repo, commit); err != nil {
				return fmt.Errorf("checkout release commit: %w", err)
			}
		}

		// If a custom ref is provided, try to merge it into the release
		// branch.
		ref := release.GetK8sRef()
		if ref != release.DefaultK8sRef {
			logrus.Infof("Merging custom ref: %s", ref)
			if err := d.impl.Merge(repo, git.Remotify(ref)); err != nil {
				return fmt.Errorf("merge k8s ref: %w", err)
			}
		}

		// Tag the repository:
		logrus.Infof("Tagging version %s", version)
		if err := d.impl.Tag(
			repo,
			version,
			fmt.Sprintf(
				"Kubernetes %s release %s", d.options.ReleaseType, version,
			),
		); err != nil {
			return fmt.Errorf("tag version: %w", err)
		}

		// if we are working on master/main at this point, we are in
		// detached HEAD state. So we checkout the branch again.
		// The next stage (build) will checkout the branch it needs, but
		// let's not end this step with a detached HEAD
		if detachHead {
			logrus.Infof("Checking out %s to reattach HEAD", d.options.ReleaseBranch)
			if err := d.impl.Checkout(repo, d.options.ReleaseBranch); err != nil {
				return fmt.Errorf("checking out branch %s: %w", d.options.ReleaseBranch, err)
			}
		}
	}
	return nil
}

func (d *DefaultStage) Build() error {
	// Log in to Docker Hub to avoid getting rate limited
	if err := d.impl.DockerHubLogin(); err != nil {
		return fmt.Errorf("logging into Docker Hub: %w", err)
	}

	// Call MakeCross for each of the versions we are building
	for _, version := range d.state.versions.Ordered() {
		if err := d.impl.MakeCross(version); err != nil {
			return fmt.Errorf("build artifacts: %w", err)
		}
	}
	return nil
}

// VerifyArtifacts checks the artifacts to ensure they are correct
func (d *DefaultStage) VerifyArtifacts() error {
	return d.impl.VerifyArtifacts(d.state.versions.Ordered())
}

func (d *DefaultStage) GenerateChangelog() error {
	repoPath := gitRoot

	if !release.IsDefaultK8sUpstream() {
		// We cannot guarantee that the k/k fork contains all release branches
		// or tags, which is why we always default to the base k/k repository.
		logrus.Info(
			"Non default Kubernetes upstream used, " +
				"cloning fresh default repository for changelog",
		)

		opts := &gogit.CloneOptions{}
		repo, err := git.CleanCloneGitHubRepo(
			release.DefaultK8sOrg, release.DefaultK8sRepo, false, true, opts,
		)
		if err != nil {
			return fmt.Errorf("clone k/k repo: %w", err)
		}
		defer func() {
			if err := repo.Cleanup(); err != nil {
				logrus.Errorf("Unable to cleanup changelog repo: %v", err)
			}
		}()

		repoPath = repo.Dir()
	}

	branch := d.options.ReleaseBranch
	if d.state.createReleaseBranch {
		branch = git.DefaultBranch
	}
	buildDir := filepath.Join(
		gitRoot,
		fmt.Sprintf("%s-%s", release.BuildDir, d.state.versions.Prime()),
	)
	return d.impl.GenerateChangelog(&changelog.Options{
		RepoPath:     repoPath,
		Tag:          d.state.versions.Prime(),
		Branch:       branch,
		Bucket:       d.options.Bucket(),
		HTMLFile:     releaseNotesHTMLFile,
		JSONFile:     releaseNotesJSONFile,
		Dependencies: true,
		CloneCVEMaps: true,
		Tars:         filepath.Join(buildDir, release.ReleaseTarsPath),
		Images:       buildDir,
	})
}

// AddBinariesToSBOM reads the produced "naked" binaries and adds them to the sbom
func (d *defaultStageImpl) AddBinariesToSBOM(sbom *spdx.Document, version string) error {
	binaries, err := d.ListBinaries(version)
	if err != nil {
		return fmt.Errorf("getting binaries list for %s: %w", version, err)
	}

	// Add the binaries, taking care of their docs
	for _, bin := range binaries {
		file := spdx.NewFile()
		if err := file.ReadSourceFile(bin.Path); err != nil {
			return fmt.Errorf("reading binary sourcefile from %s: %w", bin.Path, err)
		}
		file.Name = filepath.Join("bin", bin.Platform, bin.Arch, filepath.Base(bin.Path))
		file.FileName = file.Name
		file.LicenseConcluded = LicenseIdentifier
		if err := sbom.AddFile(file); err != nil {
			return fmt.Errorf("adding file to artifacts sbom: %w", err)
		}
		file.AddRelationship(&spdx.Relationship{
			FullRender:       false,
			PeerReference:    "SPDXRef-Package-kubernetes",
			PeerExtReference: "kubernetes-" + version,
			Comment:          "Source code",
			Type:             spdx.GENERATED_FROM,
		})
	}
	return nil
}

// AddImagesToSBOM reads the image archives from disk and adds them to the sbom
func (d *defaultStageImpl) AddTarfilesToSBOM(sbom *spdx.Document, version string) error {
	tarballs, err := d.ListTarballs(version)
	if err != nil {
		return fmt.Errorf("listing release tarballs for %s: %w", version, err)
	}

	// Once the initial doc is generated, add the tarfiles
	for _, tar := range tarballs {
		file := spdx.NewFile()
		if err := file.ReadSourceFile(tar); err != nil {
			return fmt.Errorf("reading tarball sourcefile from %s: %w", tar, err)
		}
		file.Name = filepath.Base(tar)
		file.LicenseConcluded = LicenseIdentifier
		file.FileName = filepath.Base(tar)
		if err := sbom.AddFile(file); err != nil {
			return fmt.Errorf("adding file to artifacts sbom: %w", err)
		}
		file.AddRelationship(&spdx.Relationship{
			FullRender:       false,
			PeerReference:    "SPDXRef-Package-kubernetes",
			PeerExtReference: "kubernetes-" + version,
			Comment:          "Source code",
			Type:             spdx.GENERATED_FROM,
		})
	}
	return nil
}

func (d *defaultStageImpl) BuildBaseArtifactsSBOM(options *spdx.DocGenerateOptions) (*spdx.Document, error) {
	logrus.Info("Generating release artifacts SBOM")
	return spdx.NewDocBuilder().Generate(options)
}

func (d *defaultStageImpl) GenerateVersionArtifactsBOM(version string) error {
	images, err := d.ListImageArchives(version)
	if err != nil {
		return fmt.Errorf("getting artifacts list: %w", err)
	}

	// Build the base artifacts sbom. We only pass it the images for
	// now as the binaries and tarballs need more processing
	doc, err := d.BuildBaseArtifactsSBOM(&spdx.DocGenerateOptions{
		Name:           "Kubernetes Release " + version,
		AnalyseLayers:  false,
		OnlyDirectDeps: false,
		License:        LicenseIdentifier,
		Namespace:      fmt.Sprintf("https://sbom.k8s.io/%s/release", version),
		ScanLicenses:   false,
		Tarballs:       images,
		OutputFile:     filepath.Join(),
	})
	if err != nil {
		return fmt.Errorf("generating base artifacts sbom for %s: %w", version, err)
	}

	// Add the binaries and tarballs
	if err := d.AddBinariesToSBOM(doc, version); err != nil {
		return fmt.Errorf("adding binaries to %s SBOM: %w", version, err)
	}
	if err := d.AddTarfilesToSBOM(doc, version); err != nil {
		return fmt.Errorf("adding tarballs to %s SBOM: %w", version, err)
	}

	// Reference the source code SBOM as external document
	extRef := spdx.ExternalDocumentRef{
		ID:  "kubernetes-" + version,
		URI: fmt.Sprintf("https://sbom.k8s.io/%s/source", version),
	}
	if err := extRef.ReadSourceFile(
		filepath.Join(os.TempDir(), fmt.Sprintf("source-bom-%s.spdx", version)),
	); err != nil {
		return fmt.Errorf("reading the source file as external reference: %w", err)
	}
	doc.ExternalDocRefs = append(doc.ExternalDocRefs, extRef)

	// Stamp all packages. We do this here because it includes both images and
	for _, pkg := range doc.Packages {
		pkg.AddRelationship(&spdx.Relationship{
			FullRender:       false,
			PeerReference:    "SPDXRef-Package-kubernetes",
			PeerExtReference: "kubernetes-" + version,
			Comment:          "Source code",
			Type:             spdx.GENERATED_FROM,
		})
	}

	// Write the Release Artifacts SBOM to disk
	if err := doc.Write(filepath.Join(os.TempDir(), fmt.Sprintf("release-bom-%s.spdx", version))); err != nil {
		return fmt.Errorf("writing artifacts SBOM for %s: %w", version, err)
	}
	return nil
}

func (d *defaultStageImpl) GenerateSourceTreeBOM(
	options *spdx.DocGenerateOptions,
) (*spdx.Document, error) {
	logrus.Info("Generating Kubernetes source SBOM file")
	doc, err := spdx.NewDocBuilder().Generate(options)
	if err != nil {
		return nil, fmt.Errorf("generating Kubernetes source code SBOM: %w", err)
	}
	return doc, nil
}

// WriteSourceBOM takes a source code SBOM and writes it into a file, updating
// its Namespace to match the final destination
func (d *defaultStageImpl) WriteSourceBOM(
	spdxDoc *spdx.Document, version string,
) error {
	spdxDoc.Namespace = fmt.Sprintf("https://sbom.k8s.io/%s/source", version)
	spdxDoc.Name = "kubernetes-" + version
	if err := spdxDoc.Write(filepath.Join(os.TempDir(), fmt.Sprintf("source-bom-%s.spdx", version))); err != nil {
		return fmt.Errorf("writing the source code SBOM: %w", err)
	}
	return nil
}

func (d *DefaultStage) GenerateBillOfMaterials() error {
	// For the Kubernetes source, we only generate the SBOM once as both
	// versions are cut from the same point in the git history. The
	// resulting SPDX document will be customized for each version
	// in WriteSourceBOM() before writing the actual files.
	spdxDOC, err := d.impl.GenerateSourceTreeBOM(&spdx.DocGenerateOptions{
		ProcessGoModules: true,
		License:          LicenseIdentifier,
		OutputFile:       filepath.Join(os.TempDir(), "kubernetes-source.spdx"),
		Namespace:        "https://sbom.k8s.io/REPLACE/source", // This one gets replaced when writing to disk
		ScanLicenses:     true,
		Directories:      []string{gitRoot},
	})
	if err != nil {
		return fmt.Errorf("generating the kubernetes source SBOM: %w", err)
	}

	// We generate an artifacts sbom for each of the versions
	// we are building
	for _, version := range d.state.versions.Ordered() {
		// Render the common source SBOM for this version
		if err := d.impl.WriteSourceBOM(spdxDOC, version); err != nil {
			return fmt.Errorf("writing SBOM for version %s: %w", version, err)
		}

		// Render the artifacts SBOM for version
		if err := d.impl.GenerateVersionArtifactsBOM(version); err != nil {
			return fmt.Errorf("generating SBOM for version %s: %w", version, err)
		}
	}

	return nil
}

func (d *DefaultStage) StageArtifacts() error {
	// Generate the intoto attestation, reloaded with the current run data
	statement, err := d.impl.GenerateAttestation(d.state, d.options)
	if err != nil {
		return fmt.Errorf("generating the provenance attestation: %w", err)
	}
	// Init push options for provenance document
	pushBuildOptions := &build.Options{
		Bucket:                     d.options.Bucket(),
		Registry:                   d.options.ContainerRegistry(),
		AllowDup:                   true,
		ValidateRemoteImageDigests: true,
	}
	if err := d.impl.CheckReleaseBucket(pushBuildOptions); err != nil {
		return fmt.Errorf("check release bucket access: %w", err)
	}

	// Stage the local source tree
	if err := d.impl.StageLocalSourceTree(
		pushBuildOptions,
		workspaceDir,
		d.options.BuildVersion,
	); err != nil {
		return fmt.Errorf("staging local source tree: %w", err)
	}

	// Add the sources tarball to the attestation
	subjects, err := d.impl.GetProvenanceSubjects(
		d.options, filepath.Join(workspaceDir, release.SourcesTar),
	)
	if err != nil {
		return fmt.Errorf("adding sources tarball to provenance attestation: %w", err)
	}
	statement.Subject = append(statement.Subject, subjects...)

	for _, version := range d.state.versions.Ordered() {
		logrus.Infof("Staging artifacts for version %s", version)
		buildDir := filepath.Join(
			gitRoot, fmt.Sprintf("%s-%s", release.BuildDir, version),
		)
		// Set the version-specific option for the push
		pushBuildOptions.Version = version
		pushBuildOptions.BuildDir = buildDir

		// Stage local artifacts and write checksums
		if err := d.impl.StageLocalArtifacts(pushBuildOptions); err != nil {
			return fmt.Errorf("staging local artifacts: %w", err)
		}
		gcsPath := filepath.Join(
			d.options.Bucket(), release.StagePath, d.options.BuildVersion, version,
		)

		// Push gcs-stage to GCS
		if err := d.impl.PushReleaseArtifacts(
			pushBuildOptions,
			filepath.Join(buildDir, release.GCSStagePath, version),
			filepath.Join(gcsPath, release.GCSStagePath, version),
		); err != nil {
			return fmt.Errorf("pushing release artifacts: %w", err)
		}

		// Push container release-images to GCS
		if err := d.impl.PushReleaseArtifacts(
			pushBuildOptions,
			filepath.Join(buildDir, release.ImagesPath),
			filepath.Join(gcsPath, release.ImagesPath),
		); err != nil {
			return fmt.Errorf("pushing release artifacts: %w", err)
		}

		// Push container images into registry
		if err := d.impl.PushContainerImages(pushBuildOptions); err != nil {
			return fmt.Errorf("pushing container images: %w", err)
		}

		// Add artifacts to the attestation, this should get both release-images
		// and gcs-stage directories in one call.
		subjects, err = d.impl.GetOutputDirSubjects(
			d.options, buildDir, version,
		)
		if err != nil {
			return fmt.Errorf("adding provenance of release-images for version %s: %w", version, err)
		}
		statement.Subject = append(statement.Subject, subjects...)
	}

	// Push the attestation metadata file to the bucket
	if err := d.impl.PushAttestation(statement, d.options); err != nil {
		return fmt.Errorf("writing provenance metadata to disk: %w", err)
	}

	// Delete the local source tarball
	if err := d.impl.DeleteLocalSourceTarball(pushBuildOptions, workspaceDir); err != nil {
		return fmt.Errorf("delete source tarball: %w", err)
	}

	args := ""
	if d.options.NoMock {
		args += " --nomock"
	}
	if d.options.ReleaseType != DefaultOptions().ReleaseType {
		args += " --type=" + d.options.ReleaseType
	}
	if d.options.ReleaseBranch != DefaultOptions().ReleaseBranch {
		args += " --branch=" + d.options.ReleaseBranch
	}
	args += " --build-version=" + d.options.BuildVersion

	logrus.Infof(
		"To release this staged build, run:\n\n$ krel release%s", args,
	)
	return nil
}

// GenerateAttestation creates a provenance attestation with its predicate
// preloaded with the current krel run information
func (d *defaultStageImpl) GenerateAttestation(state *StageState, options *StageOptions) (attestation *provenance.Statement, err error) {
	// Build the arguments RawMessage:
	arguments := map[string]string{
		"type":          options.ReleaseType,
		"branch":        options.ReleaseBranch,
		"build-version": options.BuildVersion,
	}
	if options.NoMock {
		arguments["nomock"] = "true"
	}

	// Fetch the last commit:
	repo, err := git.OpenRepo(gitRoot)
	if err != nil {
		return nil, fmt.Errorf("opening repository to check commit hash: %w", err)
	}

	// Get the k/k commit we are building
	commitSHA, err := repo.LastCommitSha()
	if err != nil {
		return nil, fmt.Errorf("getting k/k build point: %w", err)
	}

	// Create the predicate to populate it with the current
	// run metadata:
	p := provenance.NewSLSAPredicate()

	// SLSA v02, builder ID is a TypeURI
	p.Builder.ID = "https://git.k8s.io/release/docs/krel"

	// Some of these fields have yet to be checked to assign the
	// correct values to them
	// This is commented as the in-toto go port does not have it
	p.Metadata.BuildInvocationID = os.Getenv("BUILD_ID")
	p.Metadata.Completeness.Parameters = true // The parameters are complete as we know the from GCB
	p.Metadata.Completeness.Materials = true  // The materials are complete as we only use the github repo
	startTime := state.startTime.UTC()
	endTime := time.Now().UTC()
	p.Metadata.BuildStartedOn = &startTime
	p.Metadata.BuildFinishedOn = &endTime
	p.Invocation.ConfigSource.EntryPoint = "https://git.k8s.io/release/gcb/stage/cloudbuild.yaml"
	p.BuildType = "https://cloudbuild.googleapis.com/CloudBuildYaml@v1"
	p.Invocation.Parameters = arguments

	p.AddMaterial("git+https://github.com/kubernetes/kubernetes", slsa.DigestSet{"sha1": commitSHA})

	// Create the new attestation and attach the predicate
	attestation = provenance.NewSLSAStatement()
	attestation.Predicate = p

	return attestation, nil
}

// PushAttestation writes the provenance metadata to the staging location in
// the Google Cloud Bucket.
func (d *defaultStageImpl) PushAttestation(attestation *provenance.Statement, options *StageOptions) (err error) {
	gcsPath := filepath.Join(options.Bucket(), release.StagePath, options.BuildVersion)

	// Create a temporary file:
	f, err := os.CreateTemp("", "provenance-")
	if err != nil {
		return fmt.Errorf("creating temp file for provenance metadata: %w", err)
	}
	// Write the provenance statement to disk:
	if err := attestation.Write(f.Name()); err != nil {
		return fmt.Errorf("writing provenance attestation to disk: %w", err)
	}

	// TODO for SLSA2: Sign the attestation

	// Upload the metadata file to the staging bucket
	pushBuildOptions := &build.Options{
		Bucket:   options.Bucket(),
		AllowDup: true,
	}

	if err := d.CheckReleaseBucket(pushBuildOptions); err != nil {
		return fmt.Errorf("check release bucket access: %w", err)
	}

	// Push the provenance file to GCS
	if err := d.PushReleaseArtifacts(pushBuildOptions, f.Name(), filepath.Join(gcsPath, release.ProvenanceFilename)); err != nil {
		return fmt.Errorf("pushing provenance manifest: %w", err)
	}

	return nil
}

// GetOutputDirSubjects reads the built artifacts and returns them
// as intoto subjects. All paths are translated to their final path in the bucket
func (d *defaultStageImpl) GetOutputDirSubjects(
	options *StageOptions, path, version string,
) ([]intoto.Subject, error) {
	return release.NewProvenanceReader(&release.ProvenanceReaderOptions{
		Bucket:       options.Bucket(),
		BuildVersion: options.BuildVersion,
		WorkspaceDir: workspaceDir,
	}).GetBuildSubjects(path, version)
}

// GetProvenanceSubjects returns artifacts as intoto subjects, normalized to
// the staging bucket location
func (d *defaultStageImpl) GetProvenanceSubjects(
	options *StageOptions, path string,
) ([]intoto.Subject, error) {
	return release.NewProvenanceReader(&release.ProvenanceReaderOptions{
		Bucket:       options.Bucket(),
		BuildVersion: options.BuildVersion,
		WorkspaceDir: workspaceDir,
	}).GetStagingSubjects(path)
}
