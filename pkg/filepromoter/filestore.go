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

package filepromoter

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	api "k8s.io/release/pkg/api/files"
	"k8s.io/release/pkg/object"
)

// FilestorePromoter manages the promotion of files.
type FilestorePromoter struct {
	Source *api.Filestore
	Dest   *api.Filestore

	Files []api.File

	// UseServiceAccount must be true, for service accounts to be used
	// This gives some protection against a hostile manifest.
	UseServiceAccount bool
}

type syncFilestore interface {
	// OpenReader opens an io.ReadCloser for the specified file
	OpenReader(ctx context.Context, name string) (io.ReadCloser, error)

	// UploadFile uploads a local file to the specified destination
	UploadFile(ctx context.Context, dest string, localFile string) error

	// ListFiles returns all the file artifacts in the filestore, recursively.
	ListFiles(ctx context.Context) (map[string]*syncFileInfo, error)
}

func openFilestore(
	ctx context.Context,
	filestore *api.Filestore,
	useServiceAccount bool) (syncFilestore, error) {
	u, err := url.Parse(filestore.Base)
	if err != nil {
		return nil, fmt.Errorf(
			"error parsing filestore base %q: %v",
			filestore.Base, err)
	}

	if u.Scheme != "gs" {
		return nil, fmt.Errorf(
			"unrecognized scheme %q (supported schemes: %s)",
			object.GcsPrefix, filestore.Base,
		)
	}

	var opts []option.ClientOption
	if useServiceAccount && filestore.ServiceAccount != "" {
		ts := &gcloudTokenSource{ServiceAccount: filestore.ServiceAccount}
		opts = append(opts, option.WithTokenSource(ts))
	} else {
		opts = append(opts, option.WithoutAuthentication())
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("error building GCS client: %v", err)
	}

	prefix := strings.TrimPrefix(u.Path, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	bucket := u.Host

	s := &gcsSyncFilestore{
		filestore: filestore,
		client:    client,
		bucket:    bucket,
		prefix:    prefix,
	}
	return s, nil
}

// computeNeededOperations determines the list of files that need to be copied
// nolint[funlen]
func (p *FilestorePromoter) computeNeededOperations(
	source, dest map[string]*syncFileInfo,
	destFilestore syncFilestore) ([]SyncFileOp, error) {
	// nolint[prealloc]
	var ops []SyncFileOp

	for i := range p.Files {
		f := &p.Files[i]
		relativePath := f.Name
		sourceFile := source[relativePath]
		if sourceFile == nil {
			// TODO: Should this be a warning?
			absolutePath := joinFilepath(p.Source, relativePath)
			return nil, fmt.Errorf(
				"file %q not found in source (%q)",
				relativePath, absolutePath)
		}

		destFile := dest[relativePath]
		if destFile == nil {
			destFile = &syncFileInfo{}
			destFile.RelativePath = sourceFile.RelativePath
			destFile.AbsolutePath = joinFilepath(
				p.Dest,
				sourceFile.RelativePath)
			destFile.filestore = destFilestore
			ops = append(ops, &copyFileOp{
				Source:       sourceFile,
				Dest:         destFile,
				ManifestFile: f,
			})
			continue
		}

		changed := false
		if destFile.MD5 != sourceFile.MD5 {
			logrus.Warnf("MD5 mismatch on source %q vs dest %q: %q vs %q",
				sourceFile.AbsolutePath,
				destFile.AbsolutePath,
				sourceFile.MD5,
				destFile.MD5)
			changed = true
		}

		if destFile.Size != sourceFile.Size {
			logrus.Warnf("Size mismatch on source %q vs dest %q: %d vs %d",
				sourceFile.AbsolutePath,
				destFile.AbsolutePath,
				sourceFile.Size,
				destFile.Size)
			changed = true
		}

		if !changed {
			logrus.Infof("metadata match for %q", destFile.AbsolutePath)
			continue
		}
		ops = append(ops, &copyFileOp{
			Source:       sourceFile,
			Dest:         destFile,
			ManifestFile: f,
		})
	}

	return ops, nil
}

func joinFilepath(filestore *api.Filestore, relativePath string) string {
	s := strings.TrimSuffix(filestore.Base, "/")
	s += "/"
	s += strings.TrimPrefix(relativePath, "/")
	return s
}

// BuildOperations builds the required operations to sync from the
// Source Filestore to the Dest Filestore.
func (p *FilestorePromoter) BuildOperations(
	ctx context.Context) ([]SyncFileOp, error) {
	sourceFilestore, err := openFilestore(ctx, p.Source, p.UseServiceAccount)
	if err != nil {
		return nil, err
	}
	destFilestore, err := openFilestore(ctx, p.Dest, p.UseServiceAccount)
	if err != nil {
		return nil, err
	}

	sourceFiles, err := sourceFilestore.ListFiles(ctx)
	if err != nil {
		return nil, err
	}

	destFiles, err := destFilestore.ListFiles(ctx)
	if err != nil {
		return nil, err
	}

	return p.computeNeededOperations(sourceFiles, destFiles, destFilestore)
}
