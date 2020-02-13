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

package auth

import (
	"strings"

	"github.com/pkg/errors"

	"k8s.io/release/pkg/command"
)

func GetCurrentGCPUser() (string, error) {
	authListStatus, authListErr := command.New(
		"gcloud",
		"auth",
		"list",
		"--filter=status:ACTIVE",
		"--format=value(account)",
		"--verbosity=debug",
	).RunSilentSuccessOutput()
	if authListErr != nil {
		return "", authListErr
	}

	gcpUser := authListStatus.Output()

	if gcpUser == "" {
		return "", errors.New("the GCP user name should not be empty")
	}

	gcpUser = strings.TrimSuffix(gcpUser, "\n")
	gcpUser = strings.ReplaceAll(gcpUser, "@", "-at-")
	gcpUser = strings.ReplaceAll(gcpUser, ".", "-")
	gcpUser = strings.ToLower(gcpUser)

	return gcpUser, nil
}
