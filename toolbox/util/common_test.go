package util

import (
	"os"
	"testing"
)

func TestGetSha256(t *testing.T) {
	f, err := os.Create("/tmp/sha256_calc_testfile")
	if err != nil {
		t.Errorf("Unexpected error during creating test file: %v", err)
	}
	f.WriteString("Hello world.\n")
	f.Close()

	result, err := GetSha256("/tmp/sha256_calc_testfile")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	answer := "6472bf692aaf270d5f9dc40c5ecab8f826ecc92425c8bac4d1ea69bcbbddaea4"

	if result != answer {
		t.Errorf("Sha256sum was incorrect, want: %s, got: %s", answer, result)
	}

}
