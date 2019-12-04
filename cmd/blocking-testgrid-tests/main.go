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

package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/test-infra/testgrid/config"
)

const (
	TestgridConfigURL = "https://storage.googleapis.com/k8s-testgrid/config"
	DownloadTimeout   = 10 * time.Second

	usageFmt = `usage: %[1]s [release]
	e.g. %[1]s release-1.14
`
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, usageFmt, filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	branch := os.Args[1]
	if branch == "master" {
		branch = "release-master"
	}

	ctx := context.Background()

	conf, err := readConfFromURL(ctx, TestgridConfigURL)
	bailOnErr(err, "cannot get config")

	dashboardName := "sig-" + branch + "-blocking"
	dashboard := conf.FindDashboard(dashboardName)
	if dashboard == nil {
		bailOnErr(fmt.Errorf("%s not found", dashboardName), "finding dashboard")
	}

	for _, tab := range dashboard.DashboardTab {
		fmt.Println(tab.TestGroupName)
	}
}

func bailOnErr(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
		os.Exit(1)
	}
}

func readConfFromURL(ctx context.Context, url string) (*config.Configuration, error) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	if err := downloadFile(ctx, tmpFile.Name(), url); err != nil {
		return nil, err
	}

	return config.ReadPath(tmpFile.Name())
}

func downloadFile(ctx context.Context, filePath, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}
