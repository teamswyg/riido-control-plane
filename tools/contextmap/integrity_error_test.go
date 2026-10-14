package main

import "testing"

func TestVerifyRejectsSSOTAndSourceDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := verifyLinks(root, nil); err == nil {
		t.Fatal("verifyLinks accepted empty links")
	}
	if err := verifyLinks(root, []link{{Name: "", Path: "x.md"}}); err == nil {
		t.Fatal("verifyLinks accepted blank link metadata")
	}
	if err := verifyLinks(root, []link{{Name: "missing", Path: "missing.md"}}); err == nil {
		t.Fatal("verifyLinks accepted missing target")
	}
	if err := verifySourceChecks(root, nil); err == nil {
		t.Fatal("verifySourceChecks accepted empty checks")
	}
	writeFixtureFile(t, root, "anchors/context.txt", "wrong")
	checks := []sourceCheck{{Name: "anchor", File: "anchors/context.txt", Contains: []string{"needle"}}}
	if err := verifySourceChecks(root, checks); err == nil {
		t.Fatal("verifySourceChecks accepted missing token")
	}
}

func TestVerifyRejectsDocAndLoopDrift(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	root := writeTestRepo(t, m, false)
	writeFixtureFile(t, root, m.GeneratedDoc, "stale")
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("verifyDoc accepted stale generated doc")
	}
	m.Loop.Execute = ""
	if err := verifyLoop(m.Loop); err == nil {
		t.Fatal("verifyLoop accepted missing step")
	}
}

func TestVerifyRejectsForbiddenGoImport(t *testing.T) {
	t.Parallel()
	m := validTestManifest()
	m.DirectionRules.ForbiddenGoImports = []string{"bad/import"}
	root := writeTestRepo(t, m, true)
	writeFixtureFile(t, root, "bad.go", "package bad\nimport _ \"bad/import\"\n")
	if err := verifyForbiddenImports(root, m.DirectionRules); err == nil {
		t.Fatal("verifyForbiddenImports accepted forbidden import")
	}
}
