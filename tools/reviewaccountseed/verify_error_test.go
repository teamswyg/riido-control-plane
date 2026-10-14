package main

import (
	"path/filepath"
	"testing"
)

func TestVerifySourceChecksRejectsMissingFileAndNeedle(t *testing.T) {
	root := t.TempDir()
	checks := []sourceCheck{{Name: "missing", File: "missing.go"}}
	if err := verifySourceChecks(root, checks); err == nil {
		t.Fatal("expected missing file error")
	}
	writeReviewSeedTestFile(t, filepath.Join(root, "source.go"), "needle")
	checks = []sourceCheck{{Name: "needle", File: "source.go", Contains: []string{"absent"}}}
	if err := verifySourceChecks(root, checks); err == nil {
		t.Fatal("expected missing needle error")
	}
}

func TestVerifySeedTermsRejectsForbiddenTermsAndUnknownSets(t *testing.T) {
	m := minimalReviewSeedManifest()
	root, _ := newReviewSeedRepo(t, m)
	m.ForbiddenSeedTerms = []string{"secret"}
	writeReviewSeedTestFile(t, filepath.Join(root, m.SeedSSOT), "has secret")
	if err := verifySeedTerms(root, m); err == nil {
		t.Fatal("expected forbidden term error")
	}
	m = minimalReviewSeedManifest()
	m.ForbiddenSeedTermSets = []string{"unknown"}
	if err := verifySeedTerms(root, m); err == nil {
		t.Fatal("expected unknown forbidden set error")
	}
}

func TestVerifySeedTermsRejectsMissingSeedSSOT(t *testing.T) {
	if err := verifySeedTerms(t.TempDir(), minimalReviewSeedManifest()); err == nil {
		t.Fatal("expected missing seed SSOT error")
	}
}

func TestVerifyDocRejectsMissingAndStaleGeneratedDoc(t *testing.T) {
	root := t.TempDir()
	m := minimalReviewSeedManifest()
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected missing doc error")
	}
	writeReviewSeedTestFile(t, filepath.Join(root, m.GeneratedDoc), "stale")
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected stale doc error")
	}
}

func TestVerifyAcceptsMatchingEvidenceWithoutDocCheck(t *testing.T) {
	root := t.TempDir()
	checkPath := filepath.Join(root, "source.go")
	writeReviewSeedTestFile(t, checkPath, "needle")
	m := minimalReviewSeedManifest()
	m.SourceChecks = []sourceCheck{{Name: "source", File: "source.go", Contains: []string{"needle"}}}
	result := []caseEvidence{{Name: "one"}}
	m.Cases = []caseSpec{{Name: "one"}}
	if err := verify(root, m, result, false); err != nil {
		t.Fatal(err)
	}
}
