/*
Copyright 2021 The Kubernetes Authors.

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

package docs

import (
	"bytes"
	"context"
	"text/template"

	"sigs.k8s.io/release-sdk/github"
)

type MessageVar struct {
	Name    string
	Release string
}

func CommentOnIssue(ctx context.Context, glClient github.Client, repo string, issueNumber int, msgTpl string, msgVar MessageVar) error {
	tpl, err := template.New("msg").Parse(msgTpl)
	if err != nil {
		return err
	}

	var retBytes bytes.Buffer
	if err := tpl.Execute(&retBytes, msgVar); err != nil {
		return err
	}

	msg := retBytes.String()
	_, _, err = glClient.CreateComment(ctx, "kubernetes", repo, issueNumber, msg)
	if err != nil {
		return err
	}

	return nil
}
