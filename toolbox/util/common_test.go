package util

import (
	"os"
	"testing"
)

func TestGetSha512(t *testing.T) {
	f, err := os.Create("/tmp/sha512_calc_testfile")
	if err != nil {
		t.Errorf("Unexpected error during creating test file: %v", err)
	}
	f.WriteString("Hello world.\n")
	f.Close()

	result, err := GetSha512("/tmp/sha512_calc_testfile")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	answer := "1bb17e1ccbc87ac54a1a62335b1cc46f2016baea319ab2cf314381ace5ec7eb24b9d39e74cf34204c3e8e53a59bd9bbf66032ced69e71312689f4a52642382b4"

	if result != answer {
		t.Errorf("Sha512sum was incorrect, want: %s, got: %s", answer, result)
	}

}
