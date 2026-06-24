package main

import (
	"path/filepath"
	"testing"
)

func TestPreCommitBaselineManifest(t *testing.T) {
	err := run(options{
		Repo:        filepath.Join("..", ".."),
		Manifest:    defaultManifest,
		EvidenceOut: filepath.Join(t.TempDir(), "evidence.json"),
		CheckDoc:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPreCommitBaselineRejectsMissingPhrase(t *testing.T) {
	result := verifyResult{}
	err := requirePhrases("id: other", []string{"id: loop-registry-claim-binding"}, "hook", &result)
	if err == nil {
		t.Fatal("expected missing phrase failure")
	}
}
