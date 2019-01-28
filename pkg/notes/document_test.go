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
