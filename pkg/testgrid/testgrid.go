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

package testgrid

import (
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/testgrid/config"
	pb "github.com/GoogleCloudPlatform/testgrid/pb/config"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/http"
)

const testgridConfigURL = "https://storage.googleapis.com/k8s-testgrid/config"

// TestGrid is the default test grid client
type TestGrid struct {
	client Client
}

// New creates a new TestGrid
func New() *TestGrid {
	return &TestGrid{
		&testGridClient{},
	}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Client
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt testgridfakes/fake_client.go > testgridfakes/_fake_client.go && mv testgridfakes/_fake_client.go testgridfakes/fake_client.go"
type Client interface {
	GetURLResponse(string, bool) (string, error)
}

type testGridClient struct{}

func (t *testGridClient) GetURLResponse(url string, trim bool) (string, error) {
	return http.GetURLResponse(url, trim)
}

// SetClient can be used to set the internal HTTP client
func (t *TestGrid) SetClient(client Client) {
	t.client = client
}

// BlockingTests returns the blocking tests for the provided branch name or an
// error if those are not available
func (t *TestGrid) BlockingTests(branch string) (tests []string, err error) {
	conf, err := t.configFromURL(testgridConfigURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get config: %w", err)
	}

	dashboardName := "sig-" + branch + "-blocking"
	dashboard := config.FindDashboard(dashboardName, conf)
	if dashboard == nil {
		return nil, fmt.Errorf("dashboard %s not found", dashboardName)
	}

	for _, tab := range dashboard.GetDashboardTab() {
		tests = append(tests, tab.GetTestGroupName())
	}

	return tests, nil
}

func (t *TestGrid) configFromURL(url string) (cfg *pb.Configuration, err error) {
	logrus.Info("Retrieving testgrid configuration")

	tmpFile, err := os.CreateTemp("", "testgrid-jobs-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = os.Remove(tmpFile.Name())
		}
	}()

	response, err := t.client.GetURLResponse(url, false)
	if err != nil {
		return nil, fmt.Errorf("retrieving remote content: %w", err)
	}

	if _, err := tmpFile.WriteString(response); err != nil {
		return nil, fmt.Errorf("writing response to file: %w", err)
	}

	return config.ReadPath(tmpFile.Name())
}
