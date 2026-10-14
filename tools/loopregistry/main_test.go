package main

import "testing"

func TestLoopRegistryManifestVerifies(t *testing.T) {
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestLoopRegistryVerifyAlias(t *testing.T) {
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}

func TestClaimSemanticHashDriftFails(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatal(err)
	}
	m.Claims[0].SemanticHash = "stale"
	if _, err := verifyAll(root, m, hashes); err == nil {
		t.Fatal("expected semantic hash drift failure")
	}
}

func TestLoopRegistryRejectsStaleExpiryScopeAssertion(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatal(err)
	}
	m.Assertions = append(m.Assertions,
		"loops with evidence expiry at or below 24 hours must declare a scheduled refresh workflow")
	if _, err := verifyAll(root, m, hashes); err == nil {
		t.Fatal("expected stale expiry scope assertion failure")
	}
}
