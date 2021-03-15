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
	"os"
	"path"
	"strings"
	"time"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"google.golang.org/api/iterator"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
)

// DefaultPageSize to be used in the vulnerabilities list to set how many items
// will be retrieved in each request
const DefaultPageSize = 200

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
	projectID,
	registryHostname string,
	pageSize int32,
) ([]*grafeaspb.Occurrence, error) {
	ctx := context.Background()
	client, err := containeranalysis.NewClient(ctx)
	if err != nil {
		return nil, errors.Errorf("NewClient: %v", err)
	}
	defer client.Close()

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	req := &grafeaspb.ListOccurrencesRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		Filter:   fmt.Sprintf("kind = %q", "VULNERABILITY"),
		PageSize: pageSize,
	}

	if registryHostname != "" {
		resourceURLPrefix := fmt.Sprintf("https://%s/%s/", registryHostname, projectID)
		logrus.Infof("checking all vulnerabilities in %s", resourceURLPrefix)
		req.Filter = fmt.Sprintf("kind = %q AND has_prefix(resourceUrl, %q)", "VULNERABILITY", resourceURLPrefix)
	}

	logrus.Info("listing the vulnerabilities, will take a while...")
	var occurrenceList []*grafeaspb.Occurrence
	it := client.GetGrafeasClient().ListOccurrences(ctx, req)
	logrus.Debug("got a list of occurrences")
	for {
		var occ *grafeaspb.Occurrence
		var err error
		occ, err = it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Errorf("occurrence iteration error: %v", err)
		}
		logrus.Debug("updating the list of occurrences")
		occurrenceList = append(occurrenceList, occ)
	}
	logrus.Infof("done listing the vulnerabilities")

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
	dashboardPath,
	vulnProject,
	registryHostname,
	dashboardBucket string,
	pageSize int32,
) error {
	dashboardHTML := dashboardPath + "dashboard.html"
	logrus.Infof("opening %s", dashboardHTML)
	htmlReader, openErr := os.Open(dashboardHTML)
	if openErr != nil {
		return errors.Wrap(openErr, "opening dashboard file")
	}

	logrus.Infof("parsing %s", dashboardHTML)
	_, err := html.Parse(htmlReader)
	if err != nil {
		return errors.Errorf("dashboard.html is not valid HTML: %v", err)
	}

	logrus.Infof("uploading %s to gcs", dashboardHTML)
	err = uploadFile(dashboardPath, "dashboard.html", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard HTML: %v", err)
	}

	logrus.Info("uploading updated dashboard.js to gcs")
	err = uploadFile(dashboardPath, "dashboard.js", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard JS: %v", err)
	}

	logrus.Infof("checking all vulnerabilities for %s", vulnProject)
	productionVulnerabilities, getVulnErr := GetAllVulnerabilities(vulnProject, registryHostname, pageSize)
	if getVulnErr != nil {
		return errors.Wrap(getVulnErr, "getting all vulnerabilities")
	}

	logrus.Infof("parsing the vulnerabilities for %s", vulnProject)
	vulnBreakdowns := GenerateVulnerabilityBreakdown(productionVulnerabilities)
	jsonFile, err := json.MarshalIndent(vulnBreakdowns, "", " ")
	if err != nil {
		return errors.Errorf("Unable to generate dashboard json: %v", err)
	}

	dashboardJSON := dashboardPath + "dashboard.json"
	logrus.Infof("writing the vulnerabilities for %s in the file %s", vulnProject, dashboardJSON)
	err = os.WriteFile(dashboardJSON, jsonFile, 0644)
	if err != nil {
		return errors.Errorf("Unable to create temporary local"+
			"JSON file for the dashboard: %v", err)
	}

	logrus.Infof("uploading updated %s to gcs", dashboardJSON)
	err = uploadFile(dashboardPath, "dashboard.json", dashboardBucket)
	if err != nil {
		return errors.Errorf("Unable to upload latest version of "+
			"dashboard JSON: %v", err)
	}

	return nil
}
