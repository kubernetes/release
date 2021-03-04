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

package v1alpha1

// PlatformType is an OS / architecture tuple.
type PlatformType string

// Platform types.
const (
	PlatformTypeLinuxAMD64   PlatformType = "linux_amd64"
	PlatformTypeDarwinAMD64  PlatformType = "darwin_amd64"
	PlatformTypeWindoesAMD64 PlatformType = "windows_amd64"
	PlatformTypeLinux386     PlatformType = "linux_386"
	PlatformTypeWindows386   PlatformType = "windows_386"
	PlatformTypeLinuxARM     PlatformType = "linux_arm"
	PlatformTypeLinuxARM64   PlatformType = "linux_arm64"
	PlatformTypeLinuxPPC64LE PlatformType = "linux_ppc64le"
	PlatformTypeLinuxS390X   PlatformType = "linux_s390x"
)

// TypeMeta partially copies apimachinery/pkg/apis/meta/v1.TypeMeta
// No need for a direct dependence; the fields are stable.
type TypeMeta struct {
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// Config contains krel configuration.
type Config struct {
	TypeMeta `yaml:",inline"`
	Binaries []Binary `json:"binaries" yaml:"binaries"`
	Images   []Image  `json:"images" yaml:"images"`
}

// A Binary is a golang binary.
type Binary struct {
	Name      string         `json:"name" yaml:"name"`
	Platforms []PlatformType `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

// A File is a content file.
type File struct {
	Name string `json:"name" yaml:"name"`
}

// An Image is an OCI image.
type Image struct {
	Name         string   `json:"name" yaml:"name"`
	PlatformType []string `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}
