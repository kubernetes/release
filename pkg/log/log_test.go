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

package log_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"k8s.io/release/v1/pkg/log"
)

func TestToFile(t *testing.T) {
	file, err := ioutil.TempFile("", "log-test-")
	require.Nil(t, err)
	defer os.Remove(file.Name())

	require.Nil(t, log.SetupGlobalLogger("info"))
	require.Nil(t, log.ToFile(file.Name()))
	logrus.Info("test")

	content, err := ioutil.ReadFile(file.Name())
	require.Nil(t, err)

	require.Contains(t, string(content), "INFO")
	require.Contains(t, string(content), "test")
}
