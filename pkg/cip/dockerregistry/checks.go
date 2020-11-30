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

package inventory

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"k8s.io/release/pkg/cip/stream"
)

// MBToBytes converts a value from MiB to Bytes.
func MBToBytes(value int) int {
	const mbToBytesShift = 20
	return value << mbToBytesShift
}

// BytesToMB converts a value from Bytes to MiB.
func BytesToMB(value int) int {
	const bytesToMBShift = 20
	return value >> bytesToMBShift
}

func getGitShaFromEnv(envVar string) (plumbing.Hash, error) {
	potenitalSHA := os.Getenv(envVar)
	const gitShaLength = 40
	if len(potenitalSHA) != gitShaLength {
		return plumbing.Hash{},
			fmt.Errorf("length of SHA is %d characters, should be %d",
				len(potenitalSHA), gitShaLength)
	}
	_, err := hex.DecodeString(potenitalSHA)
	if err != nil {
		return plumbing.Hash{}, fmt.Errorf("not a valid SHA: %v", err)
	}
	return plumbing.NewHash(potenitalSHA), nil
}

// MKRealImageRemovalCheck returns an instance of ImageRemovalCheck.
func MKRealImageRemovalCheck(
	gitRepoPath string,
	edges map[PromotionEdge]interface{},
) (*ImageRemovalCheck, error) {
	// The "PULL_BASE_SHA" and "PULL_PULL_SHA" environment variables are given
	// by the PROW job running the promoter container and represent the Git SHAs
	// for the master branch and the pull request branch respectively.
	masterSHA, err := getGitShaFromEnv("PULL_BASE_SHA")
	if err != nil {
		return nil, fmt.Errorf("the PULL_BASE_SHA environment variable "+
			"is invalid: %v", err)
	}
	pullRequestSHA, err := getGitShaFromEnv("PULL_PULL_SHA")
	if err != nil {
		return nil, fmt.Errorf("the PULL_PULL_SHA environment variable "+
			"is invalid: %v", err)
	}
	return &ImageRemovalCheck{
			gitRepoPath,
			masterSHA,
			pullRequestSHA,
			edges,
		},
		nil
}

// Run executes ImageRemovalCheck on a set of promotion edges.
// Returns an error if the pull request removes images from the
// promoter manifests.
func (check *ImageRemovalCheck) Run() error {
	r, err := gogit.PlainOpen(check.GitRepoPath)
	if err != nil {
		return fmt.Errorf("could not open the Git repo: %v", err)
	}
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("could not create Git worktree: %v", err)
	}

	// The Prow job that this check is running in has already cloned the
	// git repo for us so we can just checkout the master branch to get the
	// master branch's version of the promoter manifests.
	err = w.Checkout(&gogit.CheckoutOptions{
		Hash:  check.MasterSHA,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("could not checkout the master branch of the Git"+
			" repo: %v", err)
	}

	mfests, err := ParseThinManifestsFromDir(check.GitRepoPath)
	if err != nil {
		return fmt.Errorf("could not parse manifests from the directory: %v",
			err)
	}
	masterEdges, err := ToPromotionEdges(mfests)
	if err != nil {
		return fmt.Errorf("could not generate promotion edges from promoter"+
			" manifests: %v", err)
	}

	// Reset the current directory back to the pull request branch so that this
	// check doesn't leave lasting effects that could affect subsequent checks.
	err = w.Checkout(&gogit.CheckoutOptions{
		Hash:  check.PullRequestSHA,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("could not checkout the pull request branch of the"+
			" Git repo %v: %v",
			check.GitRepoPath, err)
	}

	return check.Compare(masterEdges, check.PullEdges)
}

// Compare is a function of the ImageRemovalCheck that handles
// the comparison of the pull requests's set of promotion edges and
// the master branch's set of promotion edges.
func (check *ImageRemovalCheck) Compare(
	edgesMaster map[PromotionEdge]interface{},
	edgesPullRequest map[PromotionEdge]interface{},
) error {
	// Generate a set of all destination images that appear in
	// the pull request's set of promotion edges.
	destinationImages := make(map[PromotionEdge]interface{})
	for edge := range edgesPullRequest {
		destinationImages[PromotionEdge{
			DstImageTag: edge.DstImageTag,
			Digest:      edge.Digest,
		}] = nil
	}

	// Check that every destination image in the master branch's set of
	// promotion edges exists in the pull request's set of promotion edges.
	removedImages := make([]string, 0)
	for edge := range edgesMaster {
		_, found := destinationImages[PromotionEdge{
			DstImageTag: edge.DstImageTag,
			Digest:      edge.Digest,
		}]
		if !found {
			removedImages = append(removedImages,
				string(edge.DstImageTag.ImageName))
		}
	}

	if len(removedImages) > 0 {
		return fmt.Errorf("the following images were removed in this pull "+
			"request: %v", strings.Join(removedImages, ", "))
	}
	return nil
}

// Error is a function of ImageSizeError and implements the error interface.
func (err ImageSizeError) Error() string {
	errStr := ""
	if len(err.OversizedImages) > 0 {
		errStr += fmt.Sprintf("The following images were over the max file "+
			"size of %dMiB:\n%v\n", err.MaxImageSize,
			err.joinImageSizesToString(err.OversizedImages))
	}
	if len(err.InvalidImages) > 0 {
		errStr += fmt.Sprintf("The following images had an invalid file size "+
			"of 0 bytes or less:\n%v\n",
			err.joinImageSizesToString(err.InvalidImages))
	}
	return errStr
}

func (err ImageSizeError) joinImageSizesToString(
	imageSizes map[string]int,
) string {
	imageSizesStr := ""
	imageNames := make([]string, 0)
	for k := range imageSizes {
		imageNames = append(imageNames, k)
	}
	sort.Strings(imageNames)
	for i, imageName := range imageNames {
		imageSizesStr += imageName + " (" +
			fmt.Sprint(BytesToMB(imageSizes[imageName])) + " MiB)"
		if i < len(imageNames)-1 {
			imageSizesStr += "\n"
		}
	}
	return imageSizesStr
}

// MKRealImageSizeCheck returns an instance of ImageSizeCheck which
// checks that all images to be promoted are under a max size.
func MKRealImageSizeCheck(
	maxImageSize int,
	edges map[PromotionEdge]interface{},
	digestImageSize DigestImageSize,
) *ImageSizeCheck {
	return &ImageSizeCheck{
		maxImageSize,
		digestImageSize,
		edges,
	}
}

// Run is a function of ImageSizeCheck and checks that all
// images to be promoted are under the max file size.
func (check *ImageSizeCheck) Run() error {
	maxImageSizeByte := MBToBytes(check.MaxImageSize)
	oversizedImages := make(map[string]int)
	invalidImages := make(map[string]int)
	for edge := range check.PullEdges {
		imageSize := check.DigestImageSize[edge.Digest]
		imageName := string(edge.DstImageTag.ImageName)
		if imageSize > maxImageSizeByte {
			oversizedImages[imageName] = imageSize
		}
		if imageSize <= 0 {
			invalidImages[imageName] = imageSize
		}
	}

	if len(oversizedImages) > 0 || len(invalidImages) > 0 {
		return ImageSizeError{
			check.MaxImageSize,
			oversizedImages,
			invalidImages,
		}
	}

	return nil
}

// MKImageVulnCheck returns an instance of ImageVulnCheck which
// checks against images that have known vulnerabilities.
// nolint[funlen]
func MKImageVulnCheck(
	syncContext SyncContext,
	newPullEdges map[PromotionEdge]interface{},
	severityThreshold int,
	fakeVulnProducer ImageVulnProducer,
) *ImageVulnCheck {
	return &ImageVulnCheck{
		syncContext,
		newPullEdges,
		severityThreshold,
		fakeVulnProducer,
	}
}

// Run is a function of ImageVulnCheck and checks that none of the
// images to be promoted have any severe vulnerabilities.
// nolint[errcheck]
func (check *ImageVulnCheck) Run() error {
	var populateRequests PopulateRequests = func(
		sc *SyncContext,
		reqs chan<- stream.ExternalRequest,
		wg *sync.WaitGroup) {
		srcImages := make(map[PromotionEdge]interface{})
		for edge := range check.PullEdges {
			srcImage := PromotionEdge{
				Digest: edge.Digest,
			}
			// Only check the vulnerability for the source image if it
			// hasn't been checked already since multiple promotion
			// edges can contain the same source image
			if _, found := srcImages[srcImage]; found {
				continue
			}
			srcImages[srcImage] = nil
			var req stream.ExternalRequest
			req.RequestParams = edge
			wg.Add(1)
			reqs <- req
		}
	}

	// If no custom ImageVulnProducer is provided, we use the default producer
	// which is simply a call to the Container Analysis API which lists out
	// the vulnerability occurrences for a given image
	var vulnProducer ImageVulnProducer
	if check.FakeVulnProducer != nil {
		vulnProducer = check.FakeVulnProducer
	} else {
		ctx := context.Background()
		client, err := containeranalysis.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("NewClient: %v", err)
		}
		defer client.Close()
		vulnProducer = mkRealVulnProducer(client)
	}

	vulnerableImages := make([]string, 0)
	var processRequest ProcessRequest = func(
		sc *SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {
		for req := range reqs {
			reqRes := RequestResult{Context: req}
			errors := make(Errors, 0)
			edge := req.RequestParams.(PromotionEdge)
			occurrences, err := vulnProducer(edge)
			if err != nil {
				errors = append(errors, Error{
					Context: "error getting vulnerabilities",
					Error:   err})
			}

			fixableSevereOccurrences := 0
			for _, occ := range occurrences {
				vuln := occ.GetVulnerability()
				vulnErr := ImageVulnError{
					edge.SrcImageTag.ImageName,
					edge.Digest,
					occ.GetName(),
					vuln,
				}
				// The vulnerability check should only reject a PR if it finds
				// vulnerabilities that are both fixable and severe
				if vuln.GetFixAvailable() &&
					IsSevereOccurrence(vuln, check.SeverityThreshold) {
					errors = append(errors, Error{
						Context: "Vulnerability Occurrence w/ Fix Available",
						Error:   vulnErr,
					})
					fixableSevereOccurrences++
				} else {
					logrus.Error(vulnErr)
				}
			}

			if fixableSevereOccurrences > 0 {
				vulnerableImages = append(vulnerableImages,
					fmt.Sprintf("%v@%v [%v fixable severe vulnerabilities, "+
						"%v total]",
						edge.SrcImageTag.ImageName,
						edge.Digest,
						fixableSevereOccurrences,
						len(occurrences)))
			}

			reqRes.Errors = errors
			requestResults <- reqRes
		}
	}

	err := check.SyncContext.ExecRequests(
		populateRequests,
		processRequest,
	)
	if err != nil {
		sort.Strings(vulnerableImages)
		return fmt.Errorf("VulnerabilityCheck: "+
			"The following vulnerable images were found:\n    %v",
			strings.Join(vulnerableImages, "\n    "))
	}
	return nil
}

// Error is a function of ImageSizeError and implements the error interface.
func (err ImageVulnError) Error() string {
	// TODO: Why are we not checking errors here?
	// nolint: errcheck
	vulnJSON, _ := json.MarshalIndent(err, "", "  ")
	return string(vulnJSON)
}

// IsSevereOccurrence checks if a vulnerability is a high enough severity to
// fail the ImageVulnCheck.
func IsSevereOccurrence(
	vuln *grafeaspb.VulnerabilityOccurrence,
	severityThreshold int,
) bool {
	severityLevel := vuln.GetSeverity()
	return int(severityLevel) >= severityThreshold
}

func parseImageProjectID(edge *PromotionEdge) (string, error) {
	const projectIDIndex = 1
	splitName := strings.Split(string(edge.SrcRegistry.Name), "/")
	if len(splitName) <= projectIDIndex {
		return "", fmt.Errorf("could not parse project ID from image name: %q",
			string(edge.SrcRegistry.Name))
	}

	return splitName[projectIDIndex], nil
}

// mkRealVulnProducer returns an ImageVulnProducer that gets all vulnerability
// Occurrences associated with the image represented in the PromotionEdge
// using the Container Analysis Service client library.
// nolint[errcheck]
func mkRealVulnProducer(client *containeranalysis.Client) ImageVulnProducer {
	return func(
		edge PromotionEdge,
	) ([]*grafeaspb.Occurrence, error) {
		// resourceURL is of the form https://gcr.io/[projectID]/my-image
		resourceURL := "https://" + path.Join(string(edge.SrcRegistry.Name),
			string(edge.SrcImageTag.ImageName)) + "@" + string(edge.Digest)

		projectID, err := parseImageProjectID(&edge)
		if err != nil {
			return nil, fmt.Errorf("ParsingProjectID: %v", err)
		}

		ctx := context.Background()

		req := &grafeaspb.ListOccurrencesRequest{
			Parent: fmt.Sprintf("projects/%s", projectID),
			Filter: fmt.Sprintf("resourceUrl = %q kind = %q",
				resourceURL, "VULNERABILITY"),
		}

		var occurrenceList []*grafeaspb.Occurrence
		it := client.GetGrafeasClient().ListOccurrences(ctx, req)
		for {
			occ, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("occurrence iteration error: %v", err)
			}
			occurrenceList = append(occurrenceList, occ)
		}

		return occurrenceList, nil
	}
}
