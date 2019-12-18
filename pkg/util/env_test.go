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

package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvDefaultSuccess(t *testing.T) {
	const (
		env      = "TEST_ENV_1"
		expected = "value"
	)
	require.Nil(t, os.Setenv(env, expected))
	require.Equal(t, expected, EnvDefault(env, ""))
}

func TestEnvDefaultFailure(t *testing.T) {
	const (
		env      = "TEST_ENV_2"
		expected = "value"
	)
	require.Nil(t, os.Unsetenv(env))
	require.Equal(t, expected, EnvDefault(env, expected))
}

func TestIsEnvSetSuccess(t *testing.T) {
	const env = "TEST_ENV_3"
	require.Nil(t, os.Setenv(env, "value"))
	require.True(t, IsEnvSet(env))
}

func TestIsEnvSetFailure(t *testing.T) {
	const env = "TEST_ENV_4"
	require.Nil(t, os.Unsetenv(env))
	require.False(t, IsEnvSet(env))
}
