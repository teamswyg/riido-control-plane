package main

import "testing"

func TestVerifyRejectsWorkflowAndBoundaryDrift(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, true)
	if err := verifyFocusedWorkflows(root, nil); err == nil {
		t.Fatal("verifyFocusedWorkflows accepted empty workflows")
	}
	if err := verifyBoundaryWorkflow(root, m, boundary{ID: "x"}); err == nil {
		t.Fatal("verifyBoundaryWorkflow accepted missing workflow")
	}
	if err := verifyBoundaryWorkflow(root, m, boundary{ID: "x", Workflow: "missing.yml"}); err == nil {
		t.Fatal("verifyBoundaryWorkflow accepted unregistered workflow")
	}
	bad := m.Boundaries[0]
	bad.EvidenceArtifact = "missing-artifact"
	if err := verifyBoundaryArtifact(root, bad); err == nil {
		t.Fatal("verifyBoundaryArtifact accepted missing artifact")
	}
}

func TestVerifyRejectsBoundaryShapeDrift(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, true)
	m.Boundaries[0].ID = ""
	if err := verifyBoundaries(root, m); err == nil {
		t.Fatal("verifyBoundaries accepted blank boundary id")
	}
	m = validTestManifest()
	m.Boundaries[1].ID = m.Boundaries[0].ID
	root = writeTestRepo(t, m, true)
	if err := verifyBoundaries(root, m); err == nil {
		t.Fatal("verifyBoundaries accepted duplicate boundary id")
	}
	m = validTestManifest()
	m.Boundaries[0].SourceChecks = nil
	root = writeTestRepo(t, m, true)
	if err := verifyBoundaries(root, m); err == nil {
		t.Fatal("verifyBoundaries accepted missing source checks")
	}
}

func TestVerifyRejectsSourceDocAndPhraseDrift(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, false)
	if err := verifySourceCheck(root, "b", sourceCheck{}); err == nil {
		t.Fatal("verifySourceCheck accepted incomplete check")
	}
	check := sourceCheck{Name: "anchor", File: "missing.txt", Contains: []string{"x"}}
	if err := verifySourceCheck(root, "b", check); err == nil {
		t.Fatal("verifySourceCheck accepted missing file")
	}
	writeFixtureFile(t, root, "anchors/bad.txt", "wrong")
	check.File = "anchors/bad.txt"
	if err := verifySourceCheck(root, "b", check); err == nil {
		t.Fatal("verifySourceCheck accepted missing token")
	}
	if err := verifyDoc(root, m, renderDoc(m)); err == nil {
		t.Fatal("verifyDoc accepted missing generated doc")
	}
	phrase := []phrase{{File: "missing.md", Contains: "x"}}
	if err := verifyRequiredPhrases(root, m.GeneratedDoc, renderDoc(m), phrase); err == nil {
		t.Fatal("verifyRequiredPhrases accepted missing phrase file")
	}
}
