/*
Copyright 2024 The Kubernetes Authors.

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

package github

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/git"
)

// Options data for building the release page
type Options struct {
	// ReleaseType indicates if we are dealing with an alpha,
	// beta, rc or official
	ReleaseType string

	// AssetFiles is a list of paths of files to be uploaded
	// as assets of this release
	AssetFiles []string

	// Tag is the release the github page will be edited
	Tag string

	// The release can have a name
	Name string

	// Owner GitHub organization which owns the repository
	Owner string

	// Name of the repository where we will publish the
	// release page. The specified tag has to exist there already
	Repo string

	// Run the whole process in non-mocked mode. Which means that it uses
	// production remote locations for storing artifacts and modifying git
	// repositories.
	NoMock bool

	// Create a draft release
	Draft bool

	// If the release exists, we do not overwrite the release page
	// unless specified so.
	UpdateIfReleaseExists bool

	// We can use a custom page template by spcifiying the path. The
	// file is a go template file that renders markdown.
	PageTemplate string

	// File to read the release notes from
	ReleaseNotesFile string

	// We automatizally calculate most values, but more substitutions for
	// the template can be supplied
	Substitutions map[string]string
}

// Validate the GitHub page options to ensure they are correct
func (o *Options) Validate() error {
	// TODO: Check that the tag is well formed
	if o.Tag == "" {
		return errors.New("cannot update github page without a tag")
	}
	if o.Repo == "" {
		return errors.New("cannot update github page, repository not defined")
	}
	if o.Owner == "" {
		return errors.New("cannot update github page, github organization not defined")
	}

	return nil
}

// ParseSubstitutions gets a slice of strings with the substitutions
// for the template and parses it as Substitutions in the options
func (o *Options) ParseSubstitutions(subs []string) error {
	o.Substitutions = map[string]string{}
	for _, sString := range subs {
		p := strings.SplitN(sString, ":", 2)
		if len(p) != 2 || p[0] == "" {
			return errors.New("substitution value not well formed: " + sString)
		}
		o.Substitutions[p[0]] = p[1]
	}
	return nil
}

// SetRepository takes a repository slug in the form org/repo,
// paeses it and assigns the values to the options
func (o *Options) SetRepository(repoSlug string) error {
	org, repo, err := git.ParseRepoSlug(repoSlug)
	if err != nil {
		return fmt.Errorf("parsing repository slug: %w", err)
	}
	o.Owner = org
	o.Repo = repo
	return nil
}

// ReadTemplate reads a custom template from a file and sets
// the PageTemplate option with its content
func (o *Options) ReadTemplate(templatePath string) error {
	// If path is empty, no custom template will be used
	if templatePath == "" {
		o.PageTemplate = ""
		return nil
	}

	// Otherwise, read a custom template from a file
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading page template text: %w", err)
	}
	logrus.Infof("Using custom template from %s", templatePath)
	o.PageTemplate = string(templateData)
	return nil
}
