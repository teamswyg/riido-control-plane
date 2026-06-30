package main

import "testing"

func TestOperationalReadinessVerifyAlias(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-verify",
		"-evidence-out", out,
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}
