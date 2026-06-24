package main

import "testing"

func TestEvidenceGraphManifestVerifies(t *testing.T) {
	err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestEvidenceGraphVerifyAlias(t *testing.T) {
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}

func TestArtifactEvidenceMustBeRedacted(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	m.Chains[0].Evidence[0].Redacted = false
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected unredacted artifact evidence to fail")
	}
}
