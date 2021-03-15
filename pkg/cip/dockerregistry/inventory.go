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

package inventory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-containerregistry/pkg/crane"
	ggcrV1 "github.com/google/go-containerregistry/pkg/v1"
	ggcrV1Google "github.com/google/go-containerregistry/pkg/v1/google"
	ggcrV1Types "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"k8s.io/release/pkg/cip/gcloud"
	cipJson "k8s.io/release/pkg/cip/json"
	"k8s.io/release/pkg/cip/stream"
)

// GetSrcRegistry gets the source registry.
func GetSrcRegistry(rcs []RegistryContext) (*RegistryContext, error) {
	for _, registry := range rcs {
		registry := registry
		if registry.Src {
			return &registry, nil
		}
	}

	return nil, fmt.Errorf("could not find source registry")
}

// MakeSyncContext creates a SyncContext.
func MakeSyncContext(
	mfests []Manifest,
	threads int,
	noDryRun, useSvcAcc bool,
) (SyncContext, error) {
	sc := SyncContext{
		Threads:           threads,
		NoDryRun:          noDryRun,
		UseServiceAccount: useSvcAcc,
		Inv:               make(MasterInventory),
		InvIgnore:         []ImageName{},
		Tokens:            make(map[RootRepo]gcloud.Token),
		RegistryContexts:  make([]RegistryContext, 0),
		DigestMediaType:   make(DigestMediaType),
		DigestImageSize:   make(DigestImageSize),
		ParentDigest:      make(ParentDigest),
	}

	registriesSeen := make(map[RegistryContext]interface{})
	for _, mfest := range mfests {
		for _, r := range mfest.Registries {
			registriesSeen[r] = nil
		}
	}

	// Populate SyncContext with registries found across all manifests.
	for r := range registriesSeen {
		sc.RegistryContexts = append(sc.RegistryContexts, r)
	}

	// Sort the list for determinism. We first sort it alphabetically, then sort
	// it by length (reverse order, so that the longest registry names come
	// first). This is so that we try to match the leading prefix against the
	// longest registry names first. We sort alphabetically first because we
	// want the final order to be deterministic.
	sort.Slice(
		sc.RegistryContexts,
		func(i, j int) bool {
			return sc.RegistryContexts[i].Name < sc.RegistryContexts[j].Name
		},
	)

	sort.Slice(
		sc.RegistryContexts,
		func(i, j int) bool {
			return len(sc.RegistryContexts[i].Name) > len(sc.RegistryContexts[j].Name)
		},
	)

	// Populate access tokens for all registries listed in the manifest.
	if useSvcAcc {
		err := sc.PopulateTokens()
		if err != nil {
			return SyncContext{}, err
		}
	}

	return sc, nil
}

// LogJSONSummary logs the SyncContext's Logs as a prettified JSON.
func (sc *SyncContext) LogJSONSummary() {
	marshalled, err := json.MarshalIndent(sc.Logs, "", "  ")
	if err != nil {
		logrus.Infof("There was a problem generating the JSON summary: %v",
			err)
	} else {
		logrus.Info(string(marshalled))
	}
}

// ParseManifestFromFile parses a Manifest from a filepath.
func ParseManifestFromFile(filePath string) (Manifest, error) {
	var mfest Manifest
	var empty Manifest

	b, err := os.ReadFile(filePath)
	if err != nil {
		return empty, err
	}

	mfest, err = ParseManifestYAML(b)
	if err != nil {
		return empty, err
	}

	mfest.Filepath = filePath

	err = mfest.Finalize()
	if err != nil {
		return empty, err
	}

	return mfest, nil
}

// ParseThinManifestFromFile parses a ThinManifest from a filepath and generates
// a Manifest.
func ParseThinManifestFromFile(filePath string) (Manifest, error) {
	var thinManifest ThinManifest
	var mfest Manifest
	var empty Manifest

	b, err := os.ReadFile(filePath)
	if err != nil {
		return empty, err
	}

	thinManifest, err = ParseThinManifestYAML(b)
	if err != nil {
		return empty, err
	}

	// Get directory name holding this thin manifest.
	subProject := filepath.Base(filepath.Dir(filePath))
	imagesPath := filepath.Join(filepath.Dir(filePath),
		"../../images",
		subProject,
		"images.yaml")
	images, err := ParseImagesFromFile(imagesPath)
	if err != nil {
		return empty, err
	}

	mfest.Filepath = filePath
	mfest.Images = images
	mfest.Registries = thinManifest.Registries

	err = mfest.Finalize()
	if err != nil {
		return empty, err
	}

	return mfest, nil
}

// ParseImagesFromFile parses an Images type from a file.
func ParseImagesFromFile(filePath string) (Images, error) {
	var images Images
	var empty Images

	b, err := os.ReadFile(filePath)
	if err != nil {
		return empty, err
	}

	images, err = ParseImagesYAML(b)
	if err != nil {
		return empty, err
	}

	return images, nil
}

// Finalize finalizes a Manifest by populating extra fields.
// TODO: ST1016: methods on the same type should have the same receiver name
// nolint: stylecheck
func (m *Manifest) Finalize() error {
	// Perform semantic checks (beyond just YAML validation).
	srcRegistry, err := GetSrcRegistry(m.Registries)
	if err != nil {
		return err
	}
	m.SrcRegistry = srcRegistry

	return nil
}

// ParseThinManifestsFromDir parses all thin Manifest files within a directory.
// We effectively have to create a map of manifests, keyed by the source
// registry (there can only be 1 source registry).
//
// nolint[funlen]
func ParseThinManifestsFromDir(
	dir string,
) ([]Manifest, error) {
	mfests := make([]Manifest, 0)

	// Check that the thin manifests dir follows a regular, predefined format.
	// This is to ensure that there isn't any funny business going on around
	// paths.
	if err := ValidateThinManifestDirectoryStructure(dir); err != nil {
		return mfests, err
	}

	var parseAsManifest filepath.WalkFunc = func(path string,
		info os.FileInfo,
		err error) error {

		if err != nil {
			// Prevent panic in case of incoming errors accessing this path.
			logrus.Errorf("failure accessing a path %q: %v\n", path, err)
		}

		// Skip directories (because they are not YAML files).
		if info.IsDir() {
			return nil
		}

		// First try to parse the path as a manifest file, which must be named
		// "promoter-manifest.yaml". This restriction is in place to limit the
		// scope of what is read in as a promoter manifest.
		if filepath.Base(path) != "promoter-manifest.yaml" {
			return nil
		}

		// If there are any files named "promoter-manifest.yaml", they must be
		// inside a subfolder within "manifests/<dir>" --- any other paths are
		// forbidden.
		shortened := strings.TrimPrefix(path, dir)
		shortenedList := strings.Split(shortened, "/")
		if len(shortenedList) != ThinManifestDepth {
			return fmt.Errorf("unexpected manifest path %q",
				path)
		}

		mfest, errParse := ParseThinManifestFromFile(path)
		if errParse != nil {
			logrus.Errorf("could not parse manifest file '%s'\n", path)
			return errParse
		}

		// Save successful parse result.
		mfests = append(mfests, mfest)

		return nil
	}

	// Only look at manifests starting with the "manifests" subfolder (no need
	// to walk any other toplevel subfolder).
	if err := filepath.Walk(filepath.Join(dir, "manifests"), parseAsManifest); err != nil {
		return mfests, err
	}

	if len(mfests) == 0 {
		return nil, fmt.Errorf("no manifests found in dir: %s", dir)
	}

	return mfests, nil
}

// ValidateThinManifestDirectoryStructure enforces a particular directory
// structure for thin manifests. Most importantly, it requires that if a file
// named "foo/manifests/bar/promoter-manifest.yaml" exists, that a corresponding
// file named "foo/images/bar/promoter-manifest.yaml" must also exist.
//
// nolint[gocyclo]
func ValidateThinManifestDirectoryStructure(
	dir string,
) error {
	// First, enforce that there are directories named "images" and "manifests".
	if err := validateIsDirectory(filepath.Join(dir, "images")); err != nil {
		return err
	}

	manifestDir := filepath.Join(dir, "manifests")
	if err := validateIsDirectory(manifestDir); err != nil {
		return err
	}

	// For every subfolder in <dir>/manifests, ensure that a
	// "promoter-manifest.yaml" file exists, and also that a corresponding file
	// exists in the "images" folder.
	files, err := os.ReadDir(manifestDir)
	if err != nil {
		return err
	}

	logrus.Infof("*looking at %q", dir)
	for _, file := range files {
		p, err := os.Stat(filepath.Join(manifestDir, file.Name()))
		if err != nil {
			return err
		}

		// Skip non-directory sub-paths.
		if !p.IsDir() {
			continue
		}

		// Search for a "promoter-manifest.yaml" file under this directory.
		manifestInfo, err := os.Stat(
			filepath.Join(manifestDir,
				file.Name(),
				"promoter-manifest.yaml"))
		if err != nil {
			logrus.Warningln(err)
			continue
		}
		if !manifestInfo.Mode().IsRegular() {
			logrus.Warnf("ignoring irregular file %q", manifestInfo)
			continue
		}

		// "promoter-manifest.yaml" exists, so check for corresponding images
		// file, which MUST exist. This is why we fail early if we detect an
		// error here.
		imagesPath := filepath.Join(dir,
			"images",
			file.Name(),
			"images.yaml")
		imagesInfo, err := os.Stat(imagesPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("corresponding file %q does not exist",
					imagesPath)
			}
			return err
		}

		if !imagesInfo.Mode().IsRegular() {
			return fmt.Errorf("corresponding file %q is not a regular file",
				imagesPath)
		}
	}

	return nil
}

// validateIsDirectory returns nil if it does exist, otherwise a non-nil error.
func validateIsDirectory(dir string) error {
	p, err := os.Stat(filepath.Join(dir))
	if err != nil {
		return err
	}
	if !p.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	return nil
}

// ToPromotionEdges converts a list of manifests to a set of edges we want to
// try promoting.
func ToPromotionEdges(mfests []Manifest) (map[PromotionEdge]interface{}, error) {
	edges := make(map[PromotionEdge]interface{})
	for _, mfest := range mfests {
		for _, image := range mfest.Images {
			for digest, tagArray := range image.Dmap {
				for _, destRC := range mfest.Registries {
					if destRC == *mfest.SrcRegistry {
						continue
					}

					if len(tagArray) > 0 {
						for _, tag := range tagArray {
							edge := mkPromotionEdge(
								*mfest.SrcRegistry,
								destRC,
								image.ImageName,
								digest,
								tag)
							edges[edge] = nil
						}
					} else {
						// If this digest does not have any associated tags, still create
						// a promotion edge for it (tagless promotion).
						edge := mkPromotionEdge(
							*mfest.SrcRegistry,
							destRC,
							image.ImageName,
							digest,
							"",
						)

						edges[edge] = nil
					}
				}
			}
		}
	}

	return CheckOverlappingEdges(edges)
}

func mkPromotionEdge(
	srcRC, dstRC RegistryContext,
	srcImageName ImageName,
	digest Digest,
	tag Tag,
) PromotionEdge {
	edge := PromotionEdge{
		SrcRegistry: srcRC,
		SrcImageTag: ImageTag{
			ImageName: srcImageName,
			Tag:       tag,
		},

		Digest:      digest,
		DstRegistry: dstRC,
	}

	// The name in the destination is the same as the name in the source.
	edge.DstImageTag = ImageTag{
		ImageName: srcImageName,
		Tag:       tag,
	}

	return edge
}

// This filters out those edges from ToPromotionEdges (found in []Manifest), to
// only those PromotionEdges that makes sense to keep around. For example, we
// want to remove all edges that have already been promoted.
//
// nolint[gocyclo]
func (sc *SyncContext) GetPromotionCandidates(edges map[PromotionEdge]interface{}) (
	map[PromotionEdge]interface{},
	bool,
) {
	clean := true

	// Create lookup-optimized structure for images to ignore.
	ignoreMap := make(map[ImageName]interface{})
	for _, ignoreMe := range sc.InvIgnore {
		ignoreMap[ignoreMe] = nil
	}

	toPromote := make(map[PromotionEdge]interface{})
	// nolint[lll]
	for edge := range edges {
		// If the edge should be ignored because of a bad read in sc.Inv, drop it
		if img, ok := ignoreMap[edge.SrcImageTag.ImageName]; ok {
			logrus.Warnf("edge %v: ignoring because src image could not be read: %s\n", edge, img)
			continue
		}

		sp, dp := edge.VertexProps(&sc.Inv)

		// If dst vertex exists, NOP.
		if dp.PqinDigestMatch {
			logrus.Infof("edge %v: skipping because it was already promoted (case 1)\n", edge)
			continue
		}

		// If this edge is for a tagless promotion, skip if the digest exists in
		// the destination.
		if edge.DstImageTag.Tag == "" && dp.DigestExists {
			// Still, log a warning if the source is missing the image.
			if !sp.DigestExists {
				logrus.Errorf("edge %v: skipping %s/%s@%s because it was already promoted, but it is still _LOST_ (can't find it in src registry! please backfill it!)\n", edge, edge.SrcRegistry.Name, edge.SrcImageTag.ImageName, edge.Digest)
			}
			continue
		}

		// If src vertex missing, LOST && NOP. We just need the digest to exist
		// in src (we don't care if it points to the wrong tag).
		if !sp.DigestExists {
			logrus.Errorf("edge %v: skipping %s/%s@%s because it is _LOST_ (can't find it in src registry!)\n", edge, edge.SrcRegistry.Name, edge.SrcImageTag.ImageName, edge.Digest)
			continue
		}

		if dp.PqinExists {
			if dp.DigestExists {
				// If the destination already has the digest, but is pointing to
				// a different tag, then it's an error.
				if dp.PqinDigestMatch {
					// NOP (already promoted).
					logrus.Infof("edge %v: skipping because it was already promoted (case 2)\n", edge)
					continue
				} else {
					logrus.Errorf("edge %v: tag %s: ERROR: tag move detected from %s to %s", edge, edge.DstImageTag.Tag, edge.Digest, *sc.getDigestForTag(edge.DstImageTag.Tag))
					clean = false
					// We continue instead of returning early, because we want
					// to see and log as many errors as possible as we go
					// through each promotion edge.
					continue
				}
			} else {
				// Pqin points to the wrong digest.
				logrus.Warnf("edge %v: tag %s points to the wrong digest; moving\n", edge, dp.BadDigest)
			}
		} else {
			if dp.DigestExists {
				// Digest exists in dst, but the pqin we desire does not
				// exist. Just add the pqin to this existing digest.
				logrus.Infof("edge %v: digest %q already exists, but does not have the pqin we want (%s)\n", edge, edge.Digest, dp.OtherTags)
			} else {
				// Neither the digest nor the pqin exists in dst.
				logrus.Infof("edge %v: regular promotion (neither digest nor pqin exists in dst)\n", edge)
			}
		}

		toPromote[edge] = nil
	}

	return toPromote, clean
}

func (sc *SyncContext) getDigestForTag(inputTag Tag) *Digest {
	for _, rii := range sc.Inv {
		for _, digestTags := range rii {
			for digest, tagSlice := range digestTags {
				for _, tag := range tagSlice {
					if tag == inputTag {
						return &digest
					}
				}
			}
		}
	}
	return nil
}

// CheckOverlappingEdges checks to ensure that all the edges taken together as a
// whole are consistent. It checks that there are no duplicate promotions
// desired to the same destination vertex (same destination PQIN). If the
// digests are the same for those edges, ignore because by definition the
// digests are cryptographically guaranteed to be the same thing (it doesn't
// matter if 2 different parties want the same exact image to become promoted to
// the same destination --- in many ways this is actually a good thing because
// it's a form of redundancy). However, return an error if the digests are
// different, because most likely this is at best just human error and at worst
// a malicious attack (someone trying to push an image to an endpoint they
// shouldn't own).
//
// nolint[gocyclo]
func CheckOverlappingEdges(
	edges map[PromotionEdge]interface{}) (map[PromotionEdge]interface{}, error) {

	// Build up a "promotionIntent". This will be checked below.
	promotionIntent := make(map[string]map[Digest][]PromotionEdge)
	checked := make(map[PromotionEdge]interface{})
	for edge := range edges {
		// Skip overlap checks for edges that are tagless, because by definition
		// they cannot overlap with another edge.
		if edge.DstImageTag.Tag == "" {
			checked[edge] = nil
			continue
		}

		dstPQIN := ToPQIN(edge.DstRegistry.Name,
			edge.DstImageTag.ImageName,
			edge.DstImageTag.Tag)

		digestToEdges, ok := promotionIntent[dstPQIN]
		if ok {
			// Store the edge.
			digestToEdges[edge.Digest] = append(digestToEdges[edge.Digest], edge)
			promotionIntent[dstPQIN] = digestToEdges
		} else {
			// Make this edge lay claim to this destination vertex.
			edgeList := make([]PromotionEdge, 0)
			edgeList = append(edgeList, edge)
			digestToEdges := make(map[Digest][]PromotionEdge)
			digestToEdges[edge.Digest] = edgeList
			promotionIntent[dstPQIN] = digestToEdges
		}
	}

	// Review the promotionIntent to ensure that there are no issues.
	overlapError := false
	emptyEdgeListError := false
	for pqin, digestToEdges := range promotionIntent {
		if len(digestToEdges) < 2 {
			for _, edgeList := range digestToEdges {
				switch len(edgeList) {
				case 0:
					logrus.Errorf("no edges for %v", pqin)
					emptyEdgeListError = true
				case 1:
					checked[edgeList[0]] = nil
				default:
					logrus.Infof("redundant promotion: multiple edges want to promote the same digest to the same destination endpoint %v:", pqin)
					for _, edge := range edgeList {
						logrus.Infof("%v", edge)
					}
					logrus.Infof("using the first one: %v", edgeList[0])
					checked[edgeList[0]] = nil
				}
			}
		} else {
			logrus.Errorf("multiple edges want to promote *different* images (digests) to the same destination endpoint %v:", pqin)
			for digest, edgeList := range digestToEdges {
				logrus.Errorf("  for digest %v:\n", digest)
				for _, edge := range edgeList {
					logrus.Errorf("%v\n", edge)
				}
			}
			overlapError = true
		}
	}

	if overlapError {
		return nil, fmt.Errorf("overlapping edges detected")
	}

	if emptyEdgeListError {
		return nil, fmt.Errorf("empty edgeList(s) detected")
	}

	return checked, nil
}

// VertexProps determines the properties of each vertex (src and dst) in the
// edge, depending on the state of the world in the MasterInventory.
func (edge *PromotionEdge) VertexProps(
	mi *MasterInventory,
) (d, s VertexProperty) {
	return edge.VertexPropsFor(&edge.SrcRegistry, &edge.SrcImageTag, mi),
		edge.VertexPropsFor(&edge.DstRegistry, &edge.DstImageTag, mi)
}

// VertexPropsFor examines one of the two vertices (src or dst) of a
// PromotionEdge.
func (edge *PromotionEdge) VertexPropsFor(
	rc *RegistryContext,
	imageTag *ImageTag,
	mi *MasterInventory,
) VertexProperty {
	p := VertexProperty{}

	rii, ok := (*mi)[rc.Name]
	if !ok {
		return p
	}
	digestTags, ok := rii[imageTag.ImageName]
	if !ok {
		return p
	}

	if tagSlice, ok := digestTags[edge.Digest]; ok {
		p.DigestExists = true
		// Record the tags that are associated with this digest; it may turn out
		// that within this tagslice, we indeed have the correct digest, in
		// which we set it back to an empty slice.
		p.OtherTags = tagSlice
	}

	for digest, tagSlice := range digestTags {
		for _, tag := range tagSlice {
			if tag == imageTag.Tag {
				p.PqinExists = true
				if digest == edge.Digest {
					p.PqinDigestMatch = true
					// Both the digest and tag match what we wanted in the
					// imageTag, so there are no extraneous tags to bother with.
					p.OtherTags = TagSlice{}
				} else {
					p.BadDigest = digest
				}
			}
		}
	}

	return p
}

// ParseManifestYAML parses a Manifest from a byteslice. This function is
// separate from ParseManifestFromFile() so that it can be tested independently.
func ParseManifestYAML(b []byte) (Manifest, error) {
	var m Manifest
	if err := yaml.UnmarshalStrict(b, &m); err != nil {
		return m, err
	}

	return m, m.Validate()
}

// ParseThinManifestYAML parses a ThinManifest from a byteslice.
func ParseThinManifestYAML(b []byte) (ThinManifest, error) {
	var m ThinManifest
	if err := yaml.UnmarshalStrict(b, &m); err != nil {
		return m, err
	}

	return m, nil
}

// ParseImagesYAML parses Images from a byteslice.
func ParseImagesYAML(b []byte) (Images, error) {
	var images Images
	if err := yaml.UnmarshalStrict(b, &images); err != nil {
		return images, err
	}

	return images, nil
}

// Validate checks for semantic errors in the yaml fields (the structure of the
// yaml is checked during unmarshaling).
func (m Manifest) Validate() error {
	if err := validateRequiredComponents(m); err != nil {
		return err
	}
	return validateImages(m.Images)
}

func validateImages(images []Image) error {
	for _, image := range images {
		for digest, tagSlice := range image.Dmap {
			if err := ValidateDigest(digest); err != nil {
				return err
			}

			for _, tag := range tagSlice {
				if err := ValidateTag(tag); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ValidateDigest validates the digest.
func ValidateDigest(digest Digest) error {
	validDigest := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	if !validDigest.Match([]byte(digest)) {
		return fmt.Errorf("invalid digest: %v", digest)
	}

	return nil
}

// ValidateTag validates the tag.
func ValidateTag(tag Tag) error {
	validTag := regexp.MustCompile(`^[\w][\w.-]{0,127}$`)
	if !validTag.Match([]byte(tag)) {
		return fmt.Errorf("invalid tag: %v", tag)
	}

	return nil
}

// ValidateRegistryImagePath validates the RegistryImagePath.
func ValidateRegistryImagePath(rip RegistryImagePath) error {
	// \w is [0-9a-zA-Z_]
	validRegistryImagePath := regexp.MustCompile(`^[\w-]+(\.[\w-]+)+(/[\w-]+)+$`)

	if !validRegistryImagePath.Match([]byte(rip)) {
		return fmt.Errorf("invalid registry image path: %v", rip)
	}

	return nil
}

func (m Manifest) srcRegistryCount() int {
	var count int
	for _, registry := range m.Registries {
		if registry.Src {
			count++
		}
	}

	return count
}

func (m Manifest) srcRegistryName() RegistryName {
	for _, registry := range m.Registries {
		if registry.Src {
			return registry.Name
		}
	}

	return RegistryName("")
}

// nolint[gocyclo]
func validateRequiredComponents(m Manifest) error {
	errs := make([]string, 0)
	srcRegistryName := RegistryName("")

	if len(m.Registries) > 0 {
		if m.srcRegistryCount() > 1 {
			errs = append(errs, fmt.Sprintf("cannot have more than 1 source registry"))
		}

		srcRegistryName = m.srcRegistryName()
		if len(srcRegistryName) == 0 {
			errs = append(errs, fmt.Sprintf("source registry must be set"))
		}
	}

	knownRegistries := make([]RegistryName, 0)
	if len(m.Registries) == 0 {
		errs = append(errs, fmt.Sprintf("'registries' field cannot be empty"))
	}

	for _, registry := range m.Registries {
		if len(registry.Name) == 0 {
			errs = append(
				errs,
				fmt.Sprintf("registries: 'name' field cannot be empty"))
		}
		knownRegistries = append(knownRegistries, registry.Name)
	}

	for _, image := range m.Images {
		if len(image.ImageName) == 0 {
			errs = append(
				errs,
				fmt.Sprintf("images: 'name' field cannot be empty"))
		}

		if len(image.Dmap) == 0 {
			errs = append(
				errs,
				fmt.Sprintf("images: 'dmap' field cannot be empty"))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}

// PrettyValue creates a prettified string representation of MasterInventory.
//
// nolint[gocyclo]
func (mi *MasterInventory) PrettyValue() string {
	var b strings.Builder
	regNames := []RegistryName{}
	for regName := range *mi {
		regNames = append(regNames, regName)
	}

	sort.Slice(
		regNames,
		func(i, j int) bool {
			return regNames[i] < regNames[j]
		},
	)

	for _, regName := range regNames {
		v, ok := (*mi)[regName]
		if !ok {
			logrus.Error("corrupt MasterInventory")
			return ""
		}

		fmt.Fprintln(&b, regName)
		imageNamesSorted := make([]string, 0)
		for imageName := range v {
			imageNamesSorted = append(imageNamesSorted, string(imageName))
		}

		sort.Strings(imageNamesSorted)
		for _, imageName := range imageNamesSorted {
			fmt.Fprintf(&b, "  %v\n", imageName)
			digestTags, ok := v[ImageName(imageName)]
			if !ok {
				continue
			}

			digestSorted := make([]string, 0)
			for digest := range digestTags {
				digestSorted = append(digestSorted, string(digest))
			}

			sort.Strings(digestSorted)
			for _, digest := range digestSorted {
				fmt.Fprintf(&b, "    %v\n", digest)
				tags, ok := digestTags[Digest(digest)]
				if !ok {
					continue
				}

				for _, tag := range tags {
					fmt.Fprintf(&b, "      %v\n", tag)
				}
			}
		}
	}

	return b.String()
}

// PrettyValue converts a Registry to a prettified string representation.
func (r *Registry) PrettyValue() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v (%v)\n", r.RegistryNameLong, r.RegistryName)
	fmt.Fprintln(&b, r.RegInvImageDigest.PrettyValue())
	return b.String()
}

// PrettyValue converts a RegInvImageDigest to a prettified string
// representation.
func (riid *RegInvImageDigest) PrettyValue() string {
	var b strings.Builder
	ids := make([]ImageDigest, 0)
	for id := range *riid {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		iID := string(ids[i].ImageName) + string(ids[i].Digest)
		jID := string(ids[j].ImageName) + string(ids[i].Digest)
		return iID < jID
	})
	for _, id := range ids {
		fmt.Fprintf(&b, "  %v\n", id.ImageName)
		fmt.Fprintf(&b, "    %v\n", id.Digest)
		tagArray, ok := (*riid)[id]
		if !ok {
			continue
		}
		tagArraySorted := make([]string, 0)
		for _, tag := range tagArray {
			tagArraySorted = append(tagArraySorted, string(tag))
		}
		sort.Strings(tagArraySorted)
		for _, tag := range tagArraySorted {
			fmt.Fprintf(&b, "      %v\n", tag)
		}
	}
	return b.String()
}

func getRegistryTagsWrapper(
	req stream.ExternalRequest,
) (*ggcrV1Google.Tags, error) {
	var googleTags *ggcrV1Google.Tags

	retryFn := func() error {
		var retryErr error

		googleTags, retryErr = getRegistryTagsFrom(req)

		return retryErr
	}

	b := stream.BackoffDefault()
	notify := func(err error, t time.Duration) {
		logrus.Errorf("error: %v happened at time: %v", err, t)
	}

	err := backoff.RetryNotify(
		retryFn,
		b,
		notify,
	)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return googleTags, nil
}

func getRegistryTagsFrom(req stream.ExternalRequest,
) (*ggcrV1Google.Tags, error) {
	reader, _, err := req.StreamProducer.Produce()
	if err != nil {
		logrus.Warning("error reading from stream:", err)
		// Skip ggcrV1Google.Tags JSON parsing if there were errors reading from
		// the HTTP stream.
		return nil, err
	}

	// nolint[errcheck]
	defer req.StreamProducer.Close()

	tags, err := extractRegistryTags(reader)
	if err != nil {
		logrus.Warn("error parsing *ggcrV1Google.Tags from io.Reader handle:",
			err)
		return nil, err
	}

	return tags, nil
}

func getGCRManifestListWrapper(
	req stream.ExternalRequest,
) (*ggcrV1.IndexManifest, error) {
	var gcrManifestList *ggcrV1.IndexManifest

	retryFn := func() error {
		var retryErr error

		gcrManifestList, retryErr = getGCRManifestListFrom(req)

		return retryErr
	}

	b := stream.BackoffDefault()
	notify := func(err error, t time.Duration) {
		logrus.Errorf("error: %v happened at time: %v", err, t)
	}

	err := backoff.RetryNotify(
		retryFn,
		b,
		notify,
	)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return gcrManifestList, nil
}

func getGCRManifestListFrom(req stream.ExternalRequest) (*ggcrV1.IndexManifest, error) {
	reader, _, err := req.StreamProducer.Produce()
	if err != nil {
		logrus.Warn("error reading from stream:", err)
		return nil, err
	}

	// nolint[errcheck]
	defer req.StreamProducer.Close()

	gcrManifestList, err := extractGCRManifestList(reader)
	if err != nil {
		logrus.Warnf("for request %s: error parsing GCRManifestList from io.Reader handle: %s", req.RequestParams, err)
		return nil, err
	}

	return gcrManifestList, nil
}

func getJSONSFromProcess(req stream.ExternalRequest) (cipJson.Objects, Errors) {
	var jsons cipJson.Objects
	streamErrs := make(Errors, 0)

	stdoutReader, stderrReader, err := req.StreamProducer.Produce()
	if err != nil {
		streamErrs = append(
			streamErrs,
			Error{
				Context: "running process",
				Error:   err,
			},
		)
	}

	jsons, err = cipJson.Consume(stdoutReader)
	if err != nil {
		streamErrs = append(
			streamErrs,
			Error{
				Context: "parsing JSON",
				Error:   err,
			},
		)
	}

	be, err := io.ReadAll(stderrReader)
	if err != nil {
		streamErrs = append(
			streamErrs,
			Error{
				Context: "reading process stderr",
				Error:   err,
			},
		)
	}

	if len(be) > 0 {
		streamErrs = append(
			streamErrs,
			Error{
				Context: "process had stderr",
				Error:   fmt.Errorf("%v", string(be)),
			},
		)
	}

	err = req.StreamProducer.Close()
	if err != nil {
		streamErrs = append(
			streamErrs,
			Error{
				Context: "closing process",
				Error:   err,
			},
		)
	}

	return jsons, streamErrs
}

// IgnoreFromPromotion works by building up a new Inv type of those images that
// should NOT be bothered to be Promoted; these get ignored in the Promote()
// step later down the pipeline.
func (sc *SyncContext) IgnoreFromPromotion(regName RegistryName) {
	// regName will look like gcr.io/foo/bar/baz. We then look for the key
	// "foo/bar/baz".
	_, imgName, err := ParseContainerParts(string(regName))
	if err != nil {
		logrus.Errorf("unable to ignore from promotion: %s\n", err)
		return
	}

	logrus.Infof("ignoring from promotion: %s\n", imgName)
	sc.InvIgnore = append(sc.InvIgnore, ImageName(imgName))
}

// ParseContainerParts splits up a registry name into its component pieces.
// Unfortunately it has some specialized logic around particular inputs; this
// could be removed in a future promoter manifest version which could force the
// user to provide these delineations for us.
//
// nolint[gocyclo]
func ParseContainerParts(s string) (string, string, error) {
	parts := strings.Split(s, "/")
	if len(parts) <= 1 {
		goto error
	}
	// String may not have a double slash, or a trailing slash (which would
	// result in an empty substring).
	for _, part := range parts {
		if len(part) == 0 {
			goto error
		}
	}
	switch parts[0] {
	case "gcr.io":
		fallthrough
	case "asia.gcr.io":
		fallthrough
	case "eu.gcr.io":
		fallthrough
	case "us.gcr.io":
		if len(parts) == 2 {
			goto error
		}
		return strings.Join(parts[0:2], "/"), strings.Join(parts[2:], "/"), nil
	case "k8s.gcr.io":
		fallthrough
	case "staging-k8s.gcr.io":
		fallthrough
	default:
		return parts[0], strings.Join(parts[1:], "/"), nil
	}
error:
	return "", "", fmt.Errorf("invalid string '%s'", s)
}

// PopulateTokens populates the SyncContext's Tokens map with actual usable
// access tokens.
func (sc *SyncContext) PopulateTokens() error {
	for _, rc := range sc.RegistryContexts {
		token, err := gcloud.GetServiceAccountToken(rc.ServiceAccount, sc.UseServiceAccount)
		if err != nil {
			logrus.Errorf(
				"could not get service account token for %v",
				rc.ServiceAccount,
			)

			return err
		}

		tokenKey, _, _ := GetTokenKeyDomainRepoPath(rc.Name)
		sc.Tokens[RootRepo(tokenKey)] = token
	}

	return nil
}

// GetTokenKeyDomainRepoPath splits a string by '/'. It's OK to do this because
// the RegistryName is already parsed against a Regex. (Maybe we should store
// the repo path separately when we do the initial parse...).
func GetTokenKeyDomainRepoPath(registryName RegistryName) (key, domain, repoPath string) {
	s := string(registryName)
	i := strings.IndexByte(s, '/')
	if strings.Count(s, "/") < 2 {
		key = s
	} else {
		key = strings.Join(strings.Split(s, "/")[0:2], "/")
	}

	// key, domain, repository path
	return key, s[:i], s[i+1:]
}

// ReadRegistries reads all images in all registries in the SyncContext Each
// registry is composed of a image repositories, which can be recursive.
//
// To summarize: a docker *registry* is a set of *repositories*. It just so
// happens that to end-users, repositores resemble a tree structure because they
// are delineated by familiar filesystem-like "directory" paths.
//
// We use the term "registry" to mean the "root repository" in this program, but
// to be technically correct, for gcr.io/google-containers/foo/bar/baz:
//
//  - gcr.io is the registry
//  - gcr.io/google-containers is the toplevel repository (or "root" repo)
//  - gcr.io/google-containers/foo is a child repository
//  - gcr.io/google-containers/foo/bar is a child repository
//  - gcr.io/google-containers/foo/bar/baz is a child repository
//
// It may or may not be the case that the child repository is empty. E.g., if
// only one image gcr.io/google-containers/foo/bar/baz:1.0 exists in the entire
// registry, the foo/ and bar/ subdirs are empty repositories.
//
// The root repo, or "registry" in the loose sense, is what we care about. This
// is because in GCR, each root repo is given its own service account and
// credentials that extend to all child repos. And also in GCR, the name of the
// root repo is the same as the name of the GCP project that hosts it.
//
// NOTE: Repository names may overlap with image names. e.g., it may be in the
// example above that there are images named gcr.io/google-containers/foo:2.0
// and gcr.io/google-containers/foo/baz:2.0.
//
// nolint[gocyclo]
func (sc *SyncContext) ReadRegistries(
	toRead []RegistryContext,
	recurse bool,
	mkProducer func(*SyncContext, RegistryContext) stream.Producer,
) {
	// Collect all images in sc.Inv (the src and dest registry names found in
	// the manifest).
	var populateRequests PopulateRequests = func(
		sc *SyncContext,
		reqs chan<- stream.ExternalRequest,
		wg *sync.WaitGroup) {

		// For each registry, start the very first root "repo" read call.
		for _, rc := range toRead {
			// Create the request.
			var req stream.ExternalRequest
			req.RequestParams = rc
			req.StreamProducer = mkProducer(sc, rc)
			// Load request into the channel.
			wg.Add(1)
			reqs <- req
		}
	}
	var processRequest ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {

		for req := range reqs {
			reqRes := RequestResult{Context: req}

			// Now run the request (make network HTTP call with
			// ExponentialBackoff()).
			tagsStruct, err := getRegistryTagsWrapper(req)
			if err != nil {
				// Skip this request if it has unrecoverable errors (even after
				// ExponentialBackoff).
				reqRes.Errors = Errors{
					Error{
						Context: "getRegistryTagsWrapper",
						Error:   err,
					},
				}
				requestResults <- reqRes

				// Invalidate promotion conservatively for the subset of images
				// that touch this network request. If we have trouble reading
				// "foo" from a destination registry, do not bother trying to
				// promote it for all registries
				mutex.Lock()
				sc.IgnoreFromPromotion(req.RequestParams.(RegistryContext).Name)
				mutex.Unlock()

				continue
			}
			// Process the current repo.
			rName := req.RequestParams.(RegistryContext).Name
			digestTags := make(DigestTags)

			for digest, mfestInfo := range tagsStruct.Manifests {
				tagSlice := TagSlice{}
				for _, tag := range mfestInfo.Tags {
					tagSlice = append(tagSlice, Tag(tag))
				}
				digestTags[Digest(digest)] = tagSlice

				// Store MediaType.
				mutex.Lock()
				mediaType, err := supportedMediaType(mfestInfo.MediaType)
				if err != nil {
					fmt.Printf("digest %s: %s\n", digest, err)
				}
				sc.DigestMediaType[Digest(digest)] = mediaType

				// Store ImageSize
				sc.DigestImageSize[Digest(digest)] = int(mfestInfo.Size)
				mutex.Unlock()
			}

			// Only write an entry into our inventory if the entry has some
			// non-nil value for digestTags. This is because we only want to
			// populate the inventory with image names that have digests in
			// them, and exclude any image paths that are actually just folder
			// names without any images in them.
			if len(digestTags) > 0 {
				rootReg, imageName, err := SplitByKnownRegistries(rName, sc.RegistryContexts)
				if err != nil {
					logrus.Fatal(err)
				}

				currentRepo := make(RegInvImage)
				currentRepo[imageName] = digestTags

				mutex.Lock()
				existingRegEntry := sc.Inv[rootReg]
				if len(existingRegEntry) == 0 {
					sc.Inv[rootReg] = currentRepo
				} else {
					sc.Inv[rootReg][imageName] = digestTags
				}
				mutex.Unlock()
			}

			// Process child repos.
			if recurse {
				for _, childRepoName := range tagsStruct.Children {
					parentRC, _ := req.RequestParams.(RegistryContext)

					childRc := RegistryContext{
						Name: RegistryName(
							string(parentRC.Name) + "/" + childRepoName),
						// Inherit the service account used at the parent
						// (cascades down from toplevel to all subrepos). In the
						// future if the token exchange fails, we can refresh
						// the token here instead of using the one we inherit
						// below.
						ServiceAccount: parentRC.ServiceAccount,
						// Inherit the token as well.
						Token: parentRC.Token,
						// Don't need src, because we are just reading data
						// (don't care if it's the source reg or not).
					}

					var childReq stream.ExternalRequest
					childReq.RequestParams = childRc
					childReq.StreamProducer = mkProducer(sc, childRc)

					// Every time we "descend" into child nodes, increment the
					// semaphore.
					wg.Add(1)
					reqs <- childReq
				}
			}
			// When we're done processing this node (req), decrement the
			// semaphore.
			reqRes.Errors = Errors{}
			requestResults <- reqRes
		}
	}
	sc.ExecRequests(populateRequests, processRequest)
}

// ReadGCRManifestLists reads all manifest lists and populates the ParentDigest
// field of the SyncContext. ParentDigest is a map of values of the form
// map[ChildDigest]ParentDigest; and so, if a digest has an entry in this map,
// it is referenced by a parent DockerManifestList.
//
// TODO: Combine this function with ReadRegistries().
//
// nolint[gocyclo]
func (sc *SyncContext) ReadGCRManifestLists(
	mkProducer func(*SyncContext, *GCRManifestListContext) stream.Producer) {

	// Collect all images in sc.Inv (the src and dest registry names found in
	// the manifest).
	var populateRequests PopulateRequests = func(
		sc *SyncContext,
		reqs chan<- stream.ExternalRequest,
		wg *sync.WaitGroup,
	) {
		// Find all images that are of ggcrV1Types.MediaType == DockerManifestList; these
		// images will be queried.
		for registryName, rii := range sc.Inv {
			var rc RegistryContext
			for _, registryContext := range sc.RegistryContexts {
				if registryContext.Name == registryName {
					rc = registryContext
				}
			}
			for imageName, digestTags := range rii {
				for digest, tagSlice := range digestTags {
					if sc.DigestMediaType[digest] == ggcrV1Types.DockerManifestList {
						// Create the request.
						var req stream.ExternalRequest
						var tag Tag
						if len(tagSlice) > 0 {
							// It could be that this ManifestList has been
							// tagged multiple times. Just grab the first tag.
							tag = tagSlice[0]
						}
						gmlc := GCRManifestListContext{
							RegistryContext: rc,
							ImageName:       imageName,
							Tag:             tag,
							Digest:          digest}
						req.RequestParams = gmlc
						req.StreamProducer = mkProducer(sc, &gmlc)
						wg.Add(1)
						reqs <- req
					}
				}
			}
		}
	}

	var processRequest ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex,
	) {
		for req := range reqs {
			reqRes := RequestResult{Context: req}

			// Now run the request (make network HTTP call with
			// ExponentialBackoff()).
			gcrManifestList, err := getGCRManifestListWrapper(req)
			if err != nil {
				// Skip this request if it has unrecoverable errors (even after
				// ExponentialBackoff).
				reqRes.Errors = Errors{
					Error{
						Context: "getGCRManifestListWrapper",
						Error:   err,
					},
				}
				requestResults <- reqRes
				continue
			}

			gmlc := req.RequestParams.(GCRManifestListContext)

			for _, gManifest := range gcrManifestList.Manifests {
				mutex.Lock()
				sc.ParentDigest[(Digest)((gManifest.Digest.Algorithm)+":"+(gManifest.Digest.Hex))] = gmlc.Digest
				mutex.Unlock()
			}

			reqRes.Errors = Errors{}
			requestResults <- reqRes
		}
	}

	sc.ExecRequests(populateRequests, processRequest)
}

// FilterByTag removes all images in RegInvImage that do not match the
// filterTag.
func FilterByTag(rii RegInvImage, filterTag string) RegInvImage {
	filtered := make(RegInvImage)

	for imageName, digestTags := range rii {
		for digest, tags := range digestTags {
			for _, tag := range tags {
				if string(tag) == filterTag {
					if filtered[imageName] == nil {
						filtered[imageName] = make(DigestTags)
					}

					filtered[imageName][digest] = append(
						filtered[imageName][digest],
						tag,
					)
				}
			}
		}
	}

	return filtered
}

// RemoveChildDigestEntries removes all tagless images in RegInvImage that are
// referenced by ManifestLists in the Registries.
func (sc *SyncContext) RemoveChildDigestEntries(rii RegInvImage) RegInvImage {
	filtered := make(RegInvImage)
	for imageName, digestTags := range rii {
		for digest, tagSlice := range digestTags {
			_, hasParent := sc.ParentDigest[digest]

			// If this image digest is only referenced as part of a parent
			// ManfestList (i.e. not directly tagged), we filter it out.
			if hasParent && len(tagSlice) == 0 {
				continue
			}

			if filtered[imageName] == nil {
				filtered[imageName] = make(DigestTags)
			}

			filtered[imageName][digest] = tagSlice
		}
	}

	return filtered
}

// SplitByKnownRegistries splits a registry name into a RegistryName and
// ImageName. The purpose of this function is to split a long image path into 2
// pieces --- the repository and the image name. We can't just split by the last
// "/" all the time, because some manifests have an image with a "/" in it.
func SplitByKnownRegistries(
	r RegistryName,
	rcs []RegistryContext,
) (RegistryName, ImageName, error) {
	for _, rc := range rcs {
		if strings.HasPrefix(string(r), string(rc.Name)) {
			trimmed := strings.TrimPrefix(string(r), string(rc.Name))

			// nolint[gocritic]
			if trimmed == "" {
				// The unparsed full image path `r` and rc.Name is the same ---
				// this happens for images pushed to the root directory. Just
				// get everything past the last '/' seen in `r` to get the image
				// name.
				i := strings.LastIndex(string(r), "/")
				return rc.Name[:i], ImageName(string(r)[i+1:]), nil
			} else if trimmed[0] == '/' {
				// Remove leading "/" character. This denotes a clean split
				// along directory boundaries.
				return rc.Name, ImageName(trimmed[1:]), nil
			} else {
				// This is an unclean split where we cut the string in the
				// middle of a path name. E.g., if we have
				//
				//  rc.Name == "us.gcr.io/k8s-artifacts-prod/metrics-server"
				//  r       == "us.gcr.io/k8s-artifacts-prod/metrics-server-amd64"
				//
				// then we'l get trimmed == "-amd64". In such a case, we don't
				// make any assumptions about how this should look and delay it
				// to the next rc.
				continue
			}
		}
	}

	return "", "", fmt.Errorf("unknown registry %q", r)
}

// MkReadRepositoryCmdReal creates a stream.Producer which makes a real call
// over the network.
func MkReadRepositoryCmdReal(
	sc *SyncContext,
	rc RegistryContext,
) stream.Producer {
	var sh stream.HTTP

	tokenKey, domain, repoPath := GetTokenKeyDomainRepoPath(rc.Name)

	httpReq, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://%s/v2/%s/tags/list", domain, repoPath),
		nil,
	)
	if err != nil {
		logrus.Fatalf(
			"could not create HTTP request for '%s/%s'",
			domain,
			repoPath)
	}

	if sc.UseServiceAccount {
		token, ok := sc.Tokens[RootRepo(tokenKey)]
		if !ok {
			logrus.Fatalf("access token for key '%s' not found\n", tokenKey)
		}

		rc.Token = token
		bearer := "Bearer " + string(rc.Token)

		httpReq.Header.Add("Authorization", bearer)
	}

	sh.Req = httpReq
	return &sh
}

// MkReadManifestListCmdReal creates a stream.Producer which makes a real call
// over the network to read ManifestList information.
//
// TODO: Consider replacing stream.Producer return type with a simple ([]byte,
// error) tuple instead.
func MkReadManifestListCmdReal(sc *SyncContext, gmlc *GCRManifestListContext) stream.Producer {
	var sh stream.HTTP

	tokenKey, domain, repoPath := GetTokenKeyDomainRepoPath(gmlc.RegistryContext.Name)

	endpoint := fmt.Sprintf(
		"https://%s/v2/%s/%s/manifests/%s",
		domain,
		repoPath,
		gmlc.ImageName,
		// Always refer by a digest, because it may be the case that this
		// manifest list is not actually tagged!
		gmlc.Digest,
	)

	httpReq, err := http.NewRequest("GET", endpoint, nil)

	// Without this, GCR responds as we had used the "Accept:
	// application/vnd.docker.distribution.manifest.v1+prettyjws" header.
	httpReq.Header.Add("Accept", "*/*")

	if err != nil {
		logrus.Fatalf(
			"could not create HTTP request for manifest list '%s/%s/%s:%s'",
			domain,
			repoPath,
			gmlc.ImageName,
			gmlc.Digest,
		)
	}

	if sc.UseServiceAccount {
		token, ok := sc.Tokens[RootRepo(tokenKey)]
		if !ok {
			logrus.Fatalf("access token for key '%s' not found\n", tokenKey)
		}

		bearer := "Bearer " + string(token)
		httpReq.Header.Add("Authorization", bearer)
	}

	sh.Req = httpReq
	return &sh
}

// ExecRequests uses the Worker Pool pattern, where MaxConcurrentRequests
// determines the number of workers to spawn.
//
// nolint[funlen]
func (sc *SyncContext) ExecRequests(
	populateRequests PopulateRequests,
	processRequest ProcessRequest,
) error {
	// Run requests.
	MaxConcurrentRequests := 10

	if sc.Threads > 0 {
		MaxConcurrentRequests = sc.Threads
	}

	mutex := &sync.Mutex{}
	reqs := make(chan stream.ExternalRequest, MaxConcurrentRequests)
	requestResults := make(chan RequestResult)

	// We have to use a WaitGroup, because even though we know beforehand the
	// number of workers, we don't know the number of jobs.
	wg := new(sync.WaitGroup)

	var err error

	// Log any errors encountered.
	go func() {
		for reqRes := range requestResults {
			if len(reqRes.Errors) > 0 {
				(*mutex).Lock()
				err = fmt.Errorf("Encountered an error while executing requests")
				sc.Logs.Errors = append(sc.Logs.Errors, reqRes.Errors...)
				(*mutex).Unlock()

				logrus.Errorf(
					"Request %v: error(s) encountered: %v\n",
					reqRes.Context,
					reqRes.Errors,
				)
			} else {
				logrus.Infof("Request %v: OK\n", reqRes.Context.RequestParams)
			}

			wg.Add(-1)
		}
	}()
	for w := 0; w < MaxConcurrentRequests; w++ {
		go processRequest(sc, reqs, requestResults, wg, mutex)
	}
	// This can't be a goroutine, because the semaphore could be 0 by the time
	// wg.Wait() is called. So we need to block against the initial "seeding" of
	// workloads into the reqs channel.
	populateRequests(sc, reqs, wg)

	// Wait for all workers to finish draining the jobs.
	wg.Wait()
	close(reqs)

	// Close requestResults channel because no more new jobs are being created
	// (it's OK to close a channel even if it has nonzero length). On the other
	// hand, we cannot close the channel before we Wait() for the workers to
	// finish, because we would end up closing it too soon, and a worker would
	// end up trying to send a result to the already-closed channel.
	//
	// NOTE: This requestResults channel is only useful if we want a central
	// place to process how each request happened (good for maybe debugging slow
	// reqs? benchmarking?). If we just want to print something, for example,
	// whenever there's an error, we could do away with this channel and just
	// spit out to STDOUT wheneven we encounter an error, from whichever
	// goroutine (no need to put the error into a channel for consumption from a
	// single point).
	close(requestResults)

	return err
}

func extractRegistryTags(reader io.Reader) (*ggcrV1Google.Tags, error) {
	tags := ggcrV1Google.Tags{}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	for {
		err := decoder.Decode(&tags)
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.Error("DECODING ERROR: ", err)
			return nil, err
		}
	}
	return &tags, nil
}

func extractGCRManifestList(reader io.Reader) (*ggcrV1.IndexManifest, error) {
	gcrManifestList := ggcrV1.IndexManifest{}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	for {
		err := decoder.Decode(&gcrManifestList)
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.Error("DECODING ERROR: ", err)
			return nil, err
		}
	}
	return &gcrManifestList, nil
}

// Overwrite insert's b's values into a.
func (a DigestTags) Overwrite(b DigestTags) {
	for k, v := range b {
		a[k] = v
	}
}

// ToFQIN combines a RegistryName, ImageName, and Digest to form a
// fully-qualified image name (FQIN).
func ToFQIN(registryName RegistryName, imageName ImageName, digest Digest) string {
	return string(registryName) + "/" + string(imageName) + "@" + string(digest)
}

// ToPQIN converts a RegistryName, ImageName, and Tag to form a
// partially-qualified image name (PQIN). It's less exact than a FQIN because
// the digest information is not used.
func ToPQIN(registryName RegistryName, imageName ImageName, tag Tag) string {
	return string(registryName) + "/" + string(imageName) + ":" + string(tag)
}

// ToSorted converts a RegInvImage type to a sorted structure.
func (rii *RegInvImage) ToSorted() []ImageWithDigestSlice {
	images := make([]ImageWithDigestSlice, 0)

	for name, dmap := range *rii {
		var digests []digest
		for k, v := range dmap {
			var tags []string
			for _, tag := range v {
				tags = append(tags, string(tag))
			}

			sort.Strings(tags)

			digests = append(digests, digest{
				hash: string(k),
				tags: tags,
			})
		}
		sort.Slice(digests, func(i, j int) bool {
			return digests[i].hash < digests[j].hash
		})

		images = append(images, ImageWithDigestSlice{
			name:    string(name),
			digests: digests,
		})
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].name < images[j].name
	})

	return images
}

// ToYAML displays a RegInvImage as YAML, but with the map items sorted
// alphabetically.
//
// nolint[gocognit]
func (rii *RegInvImage) ToYAML(o YamlMarshalingOpts) string {
	images := rii.ToSorted()

	var b strings.Builder
	for _, image := range images {
		fmt.Fprintf(&b, "- name: %s\n", image.name)
		fmt.Fprintf(&b, "  dmap:\n")
		for _, digestEntry := range image.digests {
			if o.BareDigest {
				fmt.Fprintf(&b, "    %s:", digestEntry.hash)
			} else {
				fmt.Fprintf(&b, "    %q:", digestEntry.hash)
			}

			switch len(digestEntry.tags) {
			case 0:
				fmt.Fprintf(&b, " []\n")
			default:
				if o.SplitTagsOverMultipleLines {
					fmt.Fprintf(&b, "\n")
					for _, tag := range digestEntry.tags {
						fmt.Fprintf(&b, "    - %s\n", tag)
					}
				} else {
					fmt.Fprintf(&b, " [")
					for i, tag := range digestEntry.tags {
						if i == len(digestEntry.tags)-1 {
							fmt.Fprintf(&b, "%q", tag)
						} else {
							fmt.Fprintf(&b, "%q, ", tag)
						}
					}
					fmt.Fprintf(&b, "]\n")
				}
			}
		}
	}

	return b.String()
}

// ToCSV is like ToYAML, but instead of printing things in an indented
// format, it prints one image on each line as a CSV. If there is a tag pointing
// to the image, then it is printed next to the image on the same line.
//
// E.g.
//
// nolint[lll]
// a@sha256:0000000000000000000000000000000000000000000000000000000000000000,a:1.0
// a@sha256:0000000000000000000000000000000000000000000000000000000000000000,a:latest
// b@sha256:1111111111111111111111111111111111111111111111111111111111111111,-
func (rii *RegInvImage) ToCSV() string {
	images := rii.ToSorted()

	var b strings.Builder
	for _, image := range images {
		for _, digestEntry := range image.digests {
			if len(digestEntry.tags) > 0 {
				for _, tag := range digestEntry.tags {
					fmt.Fprintf(&b, "%s@%s,%s:%s\n",
						image.name,
						digestEntry.hash,
						image.name,
						tag)
				}
			} else {
				fmt.Fprintf(&b, "%s@%s,-\n", image.name, digestEntry.hash)
			}
		}
	}

	return b.String()
}

// ToLQIN converts a RegistryName and ImangeName to form a loosely-qualified
// image name (LQIN). Notice that it is missing tag information --- hence
// "loosely-qualified".
func ToLQIN(registryName RegistryName, imageName ImageName) string {
	return string(registryName) + "/" + string(imageName)
}

// SplitRegistryImagePath takes an arbitrary image path, and splits it into its
// component parts, according to the knownRegistries field. E.g., consider
// "gcr.io/foo/a/b/c" as the registryImagePath. If "gcr.io/foo" is in
// knownRegistries, then we split it into "gcr.io/foo" and "a/b/c". But if we
// were given "gcr.io/foo/a", we would split it into "gcr.io/foo/a" and "b/c".
func SplitRegistryImagePath(
	registryImagePath RegistryImagePath,
	knownRegistries []RegistryName,
) (RegistryName, ImageName, error) {
	for _, rName := range knownRegistries {
		if strings.HasPrefix(string(registryImagePath), string(rName)) {
			return rName, ImageName(registryImagePath[len(rName)+1:]), nil
		}
	}

	return RegistryName(""),
		ImageName(""),
		fmt.Errorf("could not determine registry name for '%v'", registryImagePath)
}

// MKPopulateRequestsForPromotionEdges takes in a map of PromotionEdges to promote
// and a PromotionContext and returns a PopulateRequests which can generate
// requests to be processed
func MKPopulateRequestsForPromotionEdges(
	toPromote map[PromotionEdge]interface{},
	mkProducer PromotionContext,
) PopulateRequests {
	return func(sc *SyncContext, reqs chan<- stream.ExternalRequest, wg *sync.WaitGroup) {
		if len(toPromote) == 0 {
			logrus.Info("Nothing to promote.")
			return
		}

		if !sc.NoDryRun {
			logrus.Info("---------- BEGIN PROMOTION (DRY RUN) ----------")
		} else {
			logrus.Info("---------- BEGIN PROMOTION ----------")
		}

		for promoteMe := range toPromote {
			var req stream.ExternalRequest
			oldDigest := Digest("")

			_, dp := promoteMe.VertexProps(&sc.Inv)

			if dp.PqinExists {
				if !dp.DigestExists {
					// Pqin points to the wrong digest.
					logrus.Errorf(
						"edge %v: tag '%s' in dest points to %s, not %s (as per the manifest), but tag moves are not supported; skipping\n",
						promoteMe,
						promoteMe.DstImageTag.Tag,
						dp.BadDigest,
						promoteMe.Digest,
					)

					continue
				}
			}

			// Save some information about this request. It's a bit like
			// HTTP "headers".
			req.RequestParams = PromotionRequest{
				// Only support adding new tags during a promotion run. Tag
				// moves and deletions are not supported.
				//
				// Although disallowing tag moves sounds a bit draconian, it
				// does make protect production from a malformed set of promoter
				// manifests with incorrect tag information.
				Add,
				// TODO: Clean up types to avoid having to split up promoteMe
				// prematurely like this.
				promoteMe.SrcRegistry.Name,
				promoteMe.DstRegistry.Name,
				promoteMe.DstRegistry.ServiceAccount,
				promoteMe.SrcImageTag.ImageName,
				promoteMe.DstImageTag.ImageName,
				promoteMe.Digest,
				oldDigest,
				promoteMe.DstImageTag.Tag,
			}
			wg.Add(1)
			reqs <- req
		}
	}
}

// RunChecks runs defined PreChecks in order to check the promotion.
func (sc *SyncContext) RunChecks(preChecks []PreCheck) error {
	var preCheckErrs []error
	for _, preCheck := range preChecks {
		err := preCheck.Run()
		if err != nil {
			logrus.Error(err)
			preCheckErrs = append(preCheckErrs, err)
		}
	}

	if preCheckErrs != nil {
		return fmt.Errorf("%v error(s) encountered during the prechecks",
			len(preCheckErrs))
	}
	return nil
}

// FilterPromotionEdges generates all "edges" that we want to promote.
func (sc *SyncContext) FilterPromotionEdges(
	edges map[PromotionEdge]interface{},
	readRepos bool,
) (map[PromotionEdge]interface{}, bool) {
	if readRepos {
		regs := getRegistriesToRead(edges)
		for _, reg := range regs {
			logrus.Info("reading this reg:", reg)
		}

		sc.ReadRegistries(
			regs,
			// Do not read these registries recursively, because we already know
			// exactly which repositories to read (getRegistriesToRead()).
			false,
			MkReadRepositoryCmdReal)
	}

	return sc.GetPromotionCandidates(edges)
}

// EdgesToRegInvImage takes the destination endpoints of all edges and converts
// their information to a RegInvImage type. It uses only those edges that are
// trying to promote to the given destination registry.
func EdgesToRegInvImage(
	edges map[PromotionEdge]interface{},
	destRegistry string,
) RegInvImage {
	rii := make(RegInvImage)

	destRegistry = strings.TrimRight(destRegistry, "/")

	for edge := range edges {
		imgName := ""
		prefix := ""

		if strings.HasPrefix(string(edge.DstRegistry.Name), destRegistry) {
			prefix = strings.TrimPrefix(
				string(edge.DstRegistry.Name),
				destRegistry)

			if len(prefix) > 0 {
				imgName = prefix + "/" + string(edge.DstImageTag.ImageName)
			} else {
				imgName = string(edge.DstImageTag.ImageName)
			}

			imgName = strings.TrimLeft(imgName, "/")
		} else {
			continue
		}

		if rii[ImageName(imgName)] == nil {
			rii[ImageName(imgName)] = make(DigestTags)
		}

		digestTags := rii[ImageName(imgName)]
		if len(edge.DstImageTag.Tag) > 0 {
			digestTags[edge.Digest] = append(
				digestTags[edge.Digest],
				edge.DstImageTag.Tag)
		} else {
			digestTags[edge.Digest] = TagSlice{}
		}
	}

	return rii
}

// getRegistriesToRead collects all unique Docker repositories we want to read
// from. This way, we don't have to read the entire Docker registry, but only
// those paths that we are thinking of modifying.
func getRegistriesToRead(edges map[PromotionEdge]interface{}) []RegistryContext {
	rcs := make(map[RegistryContext]interface{})

	// Save the src and dst endpoints as registries. We only care about the
	// registry and image name, not the tag or digest; this is to collect all
	// unique Docker repositories that we care about.
	for edge := range edges {
		srcReg := edge.SrcRegistry
		srcReg.Name = srcReg.Name +
			"/" +
			RegistryName(edge.SrcImageTag.ImageName)

		rcs[srcReg] = nil

		dstReg := edge.DstRegistry
		dstReg.Name = dstReg.Name +
			"/" +
			RegistryName(edge.DstImageTag.ImageName)

		rcs[dstReg] = nil
	}

	rcsFinal := []RegistryContext{}
	for rc := range rcs {
		rcsFinal = append(rcsFinal, rc)
	}

	return rcsFinal
}

// Promote performs container image promotion by realizing the intent in the
// Manifest.
//
// nolint[gocyclo]
func (sc *SyncContext) Promote(
	edges map[PromotionEdge]interface{},
	mkProducer func(
		RegistryName,
		ImageName,
		RegistryContext,
		ImageName,
		Digest,
		Tag,
		TagOp) stream.Producer,
	customProcessRequest *ProcessRequest) error {

	if len(edges) == 0 {
		logrus.Info("Nothing to promote.")
		return nil
	}

	logrus.Info("Pending promotions:")
	for edge := range edges {
		logrus.Infof("  %v\n", edge)
	}

	var populateRequests = MKPopulateRequestsForPromotionEdges(
		edges,
		mkProducer)

	var processRequest ProcessRequest
	var processRequestReal ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {

		for req := range reqs {
			reqRes := RequestResult{Context: req}
			errors := make(Errors, 0)
			// If we're adding or moving (i.e., creating a new image or
			// overwriting), do not bother shelling out to gcloud. Instead just
			// use the gcrane.doCopy() method directly.

			rpr := req.RequestParams.(PromotionRequest)
			switch rpr.TagOp {
			case Add:
				srcVertex := ToFQIN(rpr.RegistrySrc, rpr.ImageNameSrc, rpr.Digest)

				var dstVertex string

				if len(rpr.Tag) > 0 {
					dstVertex = ToPQIN(
						rpr.RegistryDest,
						rpr.ImageNameDest,
						rpr.Tag)
				} else {
					// If there is no tag, then it is a tagless promotion. So
					// the destination vertex must be referenced with a digest
					// (FQIN), not a tag (PQIN).
					dstVertex = ToFQIN(
						rpr.RegistryDest,
						rpr.ImageNameDest,
						rpr.Digest)
				}

				if err := crane.Copy(srcVertex, dstVertex); err != nil {
					logrus.Error(err)
					errors = append(errors, Error{
						Context: "running writeImage()",
						Error:   err})
				}
			case Move:
				logrus.Infof("tag moves are no longer supported")
			case Delete:
				logrus.Infof("deletions are no longer supported")
			}

			reqRes.Errors = errors
			requestResults <- reqRes
		}
	}

	captured := make(CapturedRequests)

	if !sc.NoDryRun {
		processRequestDryRun := MkRequestCapturer(&captured)
		processRequest = processRequestDryRun
	} else {
		processRequest = processRequestReal
	}

	if customProcessRequest != nil {
		processRequest = *customProcessRequest
	}

	err := sc.ExecRequests(populateRequests, processRequest)

	if !sc.NoDryRun {
		sc.PrintCapturedRequests(&captured)
	}

	return err
}

// PrintCapturedRequests pretty-prints all given PromotionRequests.
func (sc *SyncContext) PrintCapturedRequests(capReqs *CapturedRequests) {
	prs := make([]PromotionRequest, 0)

	for req, count := range *capReqs {
		for i := 0; i < count; i++ {
			prs = append(prs, req)
		}
	}

	sort.Slice(prs, func(i, j int) bool {
		return prs[i].PrettyValue() < prs[j].PrettyValue()
	})
	if len(prs) > 0 {
		fmt.Println("")
		fmt.Println("captured reqs summary:")
		fmt.Println("")
		// TODO: Consider pointers or indexing (rangeValCopy)
		// nolint: gocritic
		for _, pr := range prs {
			fmt.Printf("captured req: %v", pr.PrettyValue())
		}
		fmt.Println("")
	} else {
		fmt.Println("No requests captured.")
	}
}

// PrettyValue is a prettified string representation of a TagOp.
func (op *TagOp) PrettyValue() string {
	var tagOpPretty string
	switch *op {
	case Add:
		tagOpPretty = "ADD"
	case Move:
		tagOpPretty = "MOVE"
	case Delete:
		tagOpPretty = "DELETE"
	}

	return tagOpPretty
}

// PrettyValue is a prettified string representation of a PromotionRequest.
func (pr *PromotionRequest) PrettyValue() string {
	var b strings.Builder
	fmt.Fprintf(
		&b, "%v -> %v: Tag: '%v' <%v> %v",
		ToLQIN(pr.RegistrySrc, pr.ImageNameSrc),
		ToLQIN(pr.RegistryDest, pr.ImageNameDest),
		string(pr.Tag),
		pr.TagOp.PrettyValue(),
		string(pr.Digest),
	)

	if len(pr.DigestOld) > 0 {
		fmt.Fprintf(&b, " (move from '%v')", string(pr.DigestOld))
	}

	fmt.Fprintf(&b, "\n")

	return b.String()
}

// MkRequestCapturer returns a function that simply records requests as they are
// captured (slurped out from the reqs channel).
func MkRequestCapturer(captured *CapturedRequests) ProcessRequest {
	return func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex,
	) {
		for req := range reqs {
			// TODO: Why are we not checking errors here?
			// nolint: errcheck
			pr := req.RequestParams.(PromotionRequest)

			mutex.Lock()
			(*captured)[pr]++
			mutex.Unlock()

			// Add a request result to signal the processing of this "request".
			// This is necessary because ExecRequests() is the sole function in
			// the codebase that decrements the WaitGroup semaphore.
			requestResults <- RequestResult{}
		}
	}
}

// GarbageCollect deletes all images that are not referenced by Docker tags.
// nolint[gocyclo]
func (sc *SyncContext) GarbageCollect(
	mfest Manifest,
	mkProducer func(RegistryContext, ImageName, Digest) stream.Producer,
	customProcessRequest *ProcessRequest) {

	var populateRequests PopulateRequests = func(
		sc *SyncContext,
		reqs chan<- stream.ExternalRequest,
		wg *sync.WaitGroup) {

		for _, registry := range mfest.Registries {
			if registry.Name == sc.SrcRegistry.Name {
				continue
			}
			for imageName, digestTags := range sc.Inv[registry.Name] {
				for digest, tagArray := range digestTags {
					if len(tagArray) == 0 {
						var req stream.ExternalRequest
						req.StreamProducer = mkProducer(
							registry,
							imageName,
							digest)
						req.RequestParams = PromotionRequest{
							Delete,
							sc.SrcRegistry.Name,
							registry.Name,
							registry.ServiceAccount,

							// No source image name, because tag deletions
							// should only delete the what's in the
							// destination registry
							ImageName(""),

							imageName,
							digest,
							"",
							"",
						}
						wg.Add(1)
						reqs <- req
					}
				}
			}
		}
	}

	var processRequest ProcessRequest
	var processRequestReal ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {

		for req := range reqs {
			reqRes := RequestResult{Context: req}
			jsons, errors := getJSONSFromProcess(req)
			if len(errors) > 0 {
				reqRes.Errors = errors
				requestResults <- reqRes
				continue
			}
			for _, json := range jsons {
				logrus.Info("DELETED image:", json)
			}
			reqRes.Errors = errors
			requestResults <- reqRes
		}
	}

	captured := make(CapturedRequests)

	if !sc.NoDryRun {
		processRequestDryRun := MkRequestCapturer(&captured)
		processRequest = processRequestDryRun
	} else {
		processRequest = processRequestReal
	}

	if customProcessRequest != nil {
		processRequest = *customProcessRequest
	}

	err := sc.ExecRequests(populateRequests, processRequest)
	if err != nil {
		logrus.Info(err)
	}

	if !sc.NoDryRun {
		sc.PrintCapturedRequests(&captured)
	}
}

func supportedMediaType(v string) (ggcrV1Types.MediaType, error) {
	switch ggcrV1Types.MediaType(v) {
	case ggcrV1Types.DockerManifestList:
		return ggcrV1Types.DockerManifestList, nil
	case ggcrV1Types.DockerManifestSchema1:
		return ggcrV1Types.DockerManifestSchema1, nil
	case ggcrV1Types.DockerManifestSchema1Signed:
		return ggcrV1Types.DockerManifestSchema1Signed, nil
	case ggcrV1Types.DockerManifestSchema2:
		return ggcrV1Types.DockerManifestSchema2, nil
	default:
		return ggcrV1Types.MediaType(""),
			fmt.Errorf("unsupported MediaType %s", v)
	}
}

// ClearRepository wipes out all Docker images from a registry! Use with caution.
// nolint[gocyclo]
//
// TODO: Maybe split this into 2 parts, so that each part can be unit-tested
// separately (deletion of manifest lists vs deletion of other media types).
func (sc *SyncContext) ClearRepository(
	regName RegistryName,
	mkProducer func(RegistryContext, ImageName, Digest) stream.Producer,
	customProcessRequest *ProcessRequest) {

	// deleteRequestsPopulator returns a PopulateRequests that
	// varies by a predicate. Closure city!
	var deleteRequestsPopulator func(func(ggcrV1Types.MediaType) bool) PopulateRequests = func(predicate func(ggcrV1Types.MediaType) bool) PopulateRequests {

		var populateRequests PopulateRequests = func(
			sc *SyncContext,
			reqs chan<- stream.ExternalRequest,
			wg *sync.WaitGroup) {

			for _, registry := range sc.RegistryContexts {
				// Skip over any registry that does not match the regName we want to
				// wipe.
				if registry.Name != regName {
					continue
				}
				for imageName, digestTags := range sc.Inv[registry.Name] {
					for digest := range digestTags {
						mediaType, ok := sc.DigestMediaType[digest]
						if !ok {
							fmt.Println("could not detect MediaType of digest", digest)
							continue
						}
						if !predicate(mediaType) {
							fmt.Printf("skipping digest %s mediaType %s\n", digest, mediaType)
							continue
						}
						var req stream.ExternalRequest
						req.StreamProducer = mkProducer(
							registry,
							imageName,
							digest)
						req.RequestParams = PromotionRequest{
							Delete,
							"",
							registry.Name,
							registry.ServiceAccount,

							// No source image name, because tag deletions
							// should only delete the what's in the
							// destination registry
							ImageName(""),

							imageName,
							digest,
							"",
							"",
						}
						wg.Add(1)
						reqs <- req
					}
				}
			}
		}
		return populateRequests
	}

	var processRequest ProcessRequest
	var processRequestReal ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {

		for req := range reqs {
			reqRes := RequestResult{Context: req}
			jsons, errors := getJSONSFromProcess(req)
			for _, json := range jsons {
				logrus.Info("DELETED image:", json)
			}
			reqRes.Errors = errors
			requestResults <- reqRes
		}
	}

	captured := make(CapturedRequests)

	if !sc.NoDryRun {
		processRequestDryRun := MkRequestCapturer(&captured)
		processRequest = processRequestDryRun
	} else {
		processRequest = processRequestReal
	}

	if customProcessRequest != nil {
		processRequest = *customProcessRequest
	}

	var isEqualTo (func(ggcrV1Types.MediaType) func(ggcrV1Types.MediaType) bool) = func(want ggcrV1Types.MediaType) func(ggcrV1Types.MediaType) bool {
		return func(got ggcrV1Types.MediaType) bool {
			return want == got
		}
	}

	var isNotEqualTo (func(ggcrV1Types.MediaType) func(ggcrV1Types.MediaType) bool) = func(want ggcrV1Types.MediaType) func(ggcrV1Types.MediaType) bool {
		return func(got ggcrV1Types.MediaType) bool {
			return want != got
		}
	}

	// Avoid the GCR error that complains if you try to delete an image which is
	// referenced by a DockerManifestList, by first deleting all such manifest
	// lists.
	deleteManifestLists := deleteRequestsPopulator(isEqualTo(ggcrV1Types.DockerManifestList))
	err := sc.ExecRequests(deleteManifestLists, processRequest)
	if err != nil {
		logrus.Info(err)
	}
	deleteOthers := deleteRequestsPopulator(isNotEqualTo(ggcrV1Types.DockerManifestList))
	err = sc.ExecRequests(deleteOthers, processRequest)
	if err != nil {
		logrus.Info(err)
	}

	if !sc.NoDryRun {
		sc.PrintCapturedRequests(&captured)
	}
}

// GetWriteCmd generates a gcloud command that is used to make modifications to
// a Docker Registry.
func GetWriteCmd(
	dest RegistryContext,
	useServiceAccount bool,
	srcRegistry RegistryName,
	srcImageName ImageName,
	destImageName ImageName,
	digest Digest,
	tag Tag,
	tp TagOp,
) []string {
	var cmd []string

	switch tp {
	case Delete:
		cmd = []string{
			"gcloud",
			"--quiet",
			"container",
			"images",
			"untag",
			ToPQIN(dest.Name, destImageName, tag),
		}
	default:
		logrus.Fatalf("unsupported tag operation: %v", tp)
	}

	// Use the service account if it is desired.
	return gcloud.MaybeUseServiceAccount(
		dest.ServiceAccount,
		useServiceAccount,
		cmd,
	)
}

// GetDeleteCmd generates the cloud command used to delete images (used for
// garbage collection).
func GetDeleteCmd(
	rc RegistryContext,
	useServiceAccount bool,
	img ImageName,
	digest Digest,
	force bool,
) []string {
	fqin := ToFQIN(rc.Name, img, digest)

	cmd := []string{
		"gcloud",
		"container",
		"images",
		"delete",
		fqin,
		"--format=json",
	}

	if force {
		cmd = append(
			cmd,
			"--force-delete-tags",
			"--quiet",
		)
	}

	return gcloud.MaybeUseServiceAccount(
		rc.ServiceAccount,
		useServiceAccount,
		cmd,
	)
}

// GcrPayloadMatch holds booleans for matching a GCRPubSubPayload against a
// promoter manifest.
type GcrPayloadMatch struct {
	// Path is true if the registry + image path (everything leading
	// up to either the digest or a tag) matches.
	PathMatch bool
	// Digest is set if the digest in the payload matches a digest in
	// the promoter manifest. This is ONLY matched if the path also matches.
	DigestMatch bool
	// Tag is ONLY matched if the digest also matches.
	TagMatch bool
	// Tag is only true if the digest matches, but the tag found in
	// the payload does NOT match what is found in the promoter manifest for the
	// digest. This can happen if somone manually tweaks a tag in GCR (assume
	// bad actor) to something other than what is specified in the promoter
	// manifest.
	TagMismatch bool
}

// Match checks whether a GCRPubSubPayload is mentioned in a Manifest. The
// degree of the match is reflected in the GcrPayloadMatch result.
func (payload *GCRPubSubPayload) Match(manifest *Manifest) GcrPayloadMatch {
	var m GcrPayloadMatch
	for _, rc := range manifest.Registries {
		m = payload.matchImages(&rc, manifest.Images)
		if m.PathMatch {
			return m
		}
	}
	return m
}

func (payload *GCRPubSubPayload) matchImages(
	rc *RegistryContext,
	images []Image,
) GcrPayloadMatch {
	var m GcrPayloadMatch
	// We do not look at source registries, because the payload will only
	// contain the image name as it appears on the destination (production).
	// So the prefix and substring checks only make sense for the
	// destination registries.
	if rc.Src {
		return m
	}
	// Speed up the search by skipping over registry names whose leading
	// characters do not match.
	if !strings.HasPrefix(payload.Path, string(rc.Name)) {
		return m
	}

	for _, image := range images {
		m = payload.matchImage(rc, image)
		// If we have just a path match, return early because we should
		// limit the scope of the search to just 1 image.
		if m.PathMatch {
			return m
		}
	}
	return m
}

func (payload *GCRPubSubPayload) matchImage(
	rc *RegistryContext,
	image Image,
) GcrPayloadMatch {
	var m GcrPayloadMatch

	constructedPath := strings.Join(
		[]string{string(rc.Name), string(image.ImageName)}, "/")
	if payload.Path != constructedPath {
		return m
	}
	m.PathMatch = true

	tags, ok := image.Dmap[payload.Digest]
	if !ok {
		return m
	}
	m.DigestMatch = true

	// Now perform an additional check on the tag, if that field is
	// available. The 'tag' field is derived from the PQIN field.
	//
	// Note: if payload.PQIN is non-empty, a particular PQIN change occurred
	// in GCR. Such a change MUST match both the digest and tag combination
	// as set out in the manifest.
	//
	// Matching solely on the tag (but not digest) or vice versa goes
	// AGAINST the intent of the promoter manifest and is cause for alarm!
	if payload.Tag == "" {
		return m
	}

	for _, tag := range tags {
		if payload.Tag == tag {
			m.TagMatch = true
			break
		}
	}

	// If the digest matched, but the tag did NOT match, it's very
	// bad!
	if !m.TagMatch {
		m.TagMismatch = true
	}

	return m
}

// PopulateExtraFields takes the existing fields in GCRPubSubPayload and uses
// them to populate the extra convenience fields (these fields are derived from
// the FQIN and PQIN fields, and do not add any new information of their own).
// This is because the payload bundles up digests, tags, etc into a single
// string. Instead of dealing with them later on, we just break them up into the
// pieces we would like to use.
func (payload *GCRPubSubPayload) PopulateExtraFields() error {
	// Populate digest, if found.
	if len(payload.FQIN) > 0 {
		parsed := strings.Split(payload.FQIN, "@")
		// nolint[gomnd]
		if len(parsed) != 2 {
			return fmt.Errorf("invalid FQIN: %v", payload.FQIN)
		}
		payload.Digest = Digest(parsed[1])
		payload.Path = parsed[0]
	}

	// Populate tag, if found.
	if len(payload.PQIN) > 0 {
		parsed := strings.Split(payload.PQIN, ":")
		// nolint[gomnd]
		if len(parsed) != 2 {
			return fmt.Errorf("invalid PQIN: %v", payload.PQIN)
		}
		payload.Tag = Tag(parsed[1])
		payload.Path = parsed[0]
	}

	return nil
}

// Prettified prints the payload in a way that is stable and which hides extra
// fields which are redundant.
func (payload *GCRPubSubPayload) String() string {
	return fmt.Sprintf(
		"{Action: %q, FQIN: %q, PQIN: %q, Path: %q, Digest: %q, Tag: %q}",
		payload.Action,
		payload.FQIN,
		payload.PQIN,
		payload.Path,
		payload.Digest,
		payload.Tag,
	)
}
