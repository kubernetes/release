package main

//"reflect"
//"strings"
//"testing"

/*
func TestGetKubeadmDependencies(t *testing.T) {
	testcases := []struct {
		name    string
		version string
		deps    []string
	}{
		{
			name:    "simple test",
			version: "1.10.0",
			deps: []string{
				"kubelet (>= 1.6.0)",
				"kubectl (>= 1.6.0)",
				"kubernetes-cni (>= 0.7.5)",
				"${misc:Depends}",
			},
		},
		{
			name:    "newer than 1.11",
			version: "1.11.0",
			deps: []string{
				"kubelet (>= 1.6.0)",
				"kubectl (>= 1.6.0)",
				"kubernetes-cni (>= 0.7.5)",
				"${misc:Depends}",
				"cri-tools (>= 1.11.1)",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := version{Version: tc.version}
			deps, err := getKubeadmDependencies(v)
			if err != nil {
				t.Fatalf("did not expect an error: %v", err)
			}
			actual := strings.Split(deps, ", ")
			if len(actual) != len(tc.deps) {
				t.Fatalf("Expected %d deps but found %d", len(tc.deps), len(actual))
			}
			if !reflect.DeepEqual(actual, tc.deps) {
				t.Fatalf("expected %q but got %q", tc.deps, actual)
			}
		})
	}

}
*/
