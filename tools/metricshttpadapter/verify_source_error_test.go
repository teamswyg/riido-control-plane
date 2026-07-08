package main

import (
	"strings"
	"testing"
)

func TestVerifyRejectsSourceCheckFailure(t *testing.T) {
	t.Parallel()
	m := testManifest()
	result, err := exerciseAdapter(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := verify(t.TempDir(), m, result, false); err == nil || !strings.Contains(err.Error(), "source") {
		t.Fatalf("verify error = %v, want source check", err)
	}
}
