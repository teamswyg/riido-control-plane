package main

import (
	"strings"
	"testing"
)

func TestVerifyRejectsSourceCheckFailure(t *testing.T) {
	t.Parallel()
	shape, err := buildEMFShape()
	if err != nil {
		t.Fatal(err)
	}
	if err := verify(t.TempDir(), testManifest(), shape, false); err == nil ||
		!strings.Contains(err.Error(), "source") {
		t.Fatalf("verify error = %v, want source check", err)
	}
}
