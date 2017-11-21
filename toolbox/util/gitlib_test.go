package util

import "testing"

func TestIsVer(t *testing.T) {
	tables := []struct {
		v     string
		t     string
		isVer bool
	}{
		{"v1.8", "dotzero", false},
		{"v1.8.0", "dotzero", true},
		{"v1.8.00", "dotzero", false},
		{"v1.8.0.0", "dotzero", false},
		{"v1.8.1", "dotzero", false},
		{"v1.8.1.0", "dotzero", false},
	}

	for _, table := range tables {
		result := IsVer(table.v, table.t)
		if result != table.isVer {
			t.Errorf("%v %v: IsVer check failed, want: %v, got: %v", table.v, table.t, table.isVer, result)
		}
	}
}
