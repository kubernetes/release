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

package notes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettySIG(t *testing.T) {
	cases := map[string]string{
		"scheduling":        "Scheduling",
		"cluster-lifecycle": "Cluster Lifecycle",
		"cli":               "CLI",
		"aws":               "AWS",
		"api-machinery":     "API Machinery",
		"vsphere":           "vSphere",
		"openstack":         "OpenStack",
	}

	for input, expected := range cases {
		require.Equal(t, expected, (prettySIG(input)))
	}
}
