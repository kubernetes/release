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

package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"google.golang.org/api/iterator"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
)

func uploadFile(directory, filename, bucket string) error {
	const timeout = 60
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Open local file
	f, err := os.Open(directory + filename)
	if err != nil {
		return errors.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	// Upload the object with storage.Writer
	wc := client.Bucket(bucket).Object(filename).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return errors.Errorf("io.Copy: %v", err)
	}

	if err := wc.Close(); err != nil {
		return errors.Errorf("Writer.Close: %v", err)
	}

	return nil
}

// GetAllVulnerabilities gets all of the vulnerability occurrences associated
// with images in a specific project using the Container Analysis Service.
func GetAllVulnerabilities(
	projectID string,
) ([]*grafeaspb.Occurrence, error) {
	ctx := context.Background()
	client, err := containeranalysis.NewClient(ctx)
	if err != nil {
		return nil, errors.Errorf("NewClient: %v", err)
	}
	defer client.Close()

	req := &grafeaspb.ListOccurrencesRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Filter: fmt.Sprintf("kind = %q", "VULNERABILITY"),
	}

	var occurrenceList []*grafeaspb.Occurrence
	it := client.GetGrafeasClient().ListOccurrences(ctx, req)
	for {
		occ, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Errorf("occurrence iteration error: %v", err)
		}
		occurrenceList = append(occurrenceList, occ)
	}

	return occurrenceList, err
}

func parseImageResourceURL(resourceURL string) (registryImageName, digest string) {
	FQIN := path.Base(resourceURL)
	splitFQIN := strings.Split(FQIN, "@")
	registryImageName, digest = splitFQIN[0], splitFQIN[1]

	return registryImageName, digest
}

func parseVulnName(noteName string) string {
	return path.Base(noteName)
}

// GenerateVulnerabilityBreakdown parses the a slice of vulnerability
// occurrences into a breakdown that only contains the necessary information
// for each vulnerability.
func GenerateVulnerabilityBreakdown(
	productionVulnerabilities []*grafeaspb.Occurrence,
) map[string]ImageVulnBreakdown {
	vulnBreakdowns := make(map[string]ImageVulnBreakdown)

	for _, occ := range productionVulnerabilities {
		// resourceURI is a url pointing to a specific image
		// in the form gcr.io/project/foo@sha256:111
		if _, found := vulnBreakdowns[occ.ResourceUri]; !found {
			imageName, imageDigest := parseImageResourceURL(occ.ResourceUri)
			vulnBreakdowns[occ.ResourceUri] = ImageVulnBreakdown{
				occ.ResourceUri,
				imageName,
				imageDigest,
				0,
				[]string{},
				[]string{},
			}
		}

		imageVulnBreakdown := vulnBreakdowns[occ.ResourceUri]
		imageVulnBreakdown.NumVulnerabilities++

		vulnName := parseVulnName(occ.NoteName)
		vuln := occ.GetVulnerability()
		if vuln.GetSeverity() == grafeaspb.Severity_CRITICAL {
			imageVulnBreakdown.CriticalVulnerabilities = append(
				imageVulnBreakdown.CriticalVulnerabilities,
				vulnName,
			)
		}
		if vuln.GetFixAvailable() {
			imageVulnBreakdown.FixableVulnerabilities = append(
				imageVulnBreakdown.FixableVulnerabilities,
				vulnName,
			)
		}
		vulnBreakdowns[occ.ResourceUri] = imageVulnBreakdown
	}

	return vulnBreakdowns
}

// UpdateVulnerabilityDashboard updates the vulnerability dashboard by uploading
// the lastest versions of all the vulnerability dashboard's files.
func UpdateVulnerabilityDashboard(
	dashboardPath string,
	vulnProject string,
	dashboardBucket string,
) error {
	htmlReader, openErr := os.Open(dashboardPath + "dashboard.html")
	if openErr != nil {
		return errors.Wrap(openErr, "opening dashboard file")
	}

	_, err := html.Parse(htmlReader)
	if err != nil {
		return errors.Errorf("dashboard.html is not valid HTML: %v", err)
	}
	err = uploadFile(dashboardPath, "dashboard.html", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard HTML: %v", err)
	}

	err = uploadFile(dashboardPath, "dashboard.js", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard JS: %v", err)
	}

	productionVulnerabilities, getVulnErr := GetAllVulnerabilities(vulnProject)
	if getVulnErr != nil {
		return errors.Wrap(getVulnErr, "getting all vulnerabilities")
	}

	vulnBreakdowns := GenerateVulnerabilityBreakdown(productionVulnerabilities)
	jsonFile, err := json.MarshalIndent(vulnBreakdowns, "", " ")
	if err != nil {
		return errors.Errorf("Unable to generate dashboard json: %v", err)
	}

	err = ioutil.WriteFile(dashboardPath+"dashboard.json",
		jsonFile, 0644)
	if err != nil {
		return errors.Errorf("Unable to create temporary local"+
			"JSON file for the dashboard: %v", err)
	}

	err = uploadFile(dashboardPath, "dashboard.json", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard JSON: %v", err)
	}

	return nil
}
