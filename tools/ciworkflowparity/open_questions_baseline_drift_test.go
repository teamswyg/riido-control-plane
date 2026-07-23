package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsOpenQuestionsWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "open-questions.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("open-questions workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[10].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("open-questions digest substitution must fail closed")
	}
}

func TestVerifyRejectsOpenQuestionsAdapterEnvironmentOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[10].NativeMapping.OpenQuestions.AdapterRequired = true
	document.BoundedChildren[10].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("open-questions parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[10].ParityClaim.OpenQuestionsLoopEnvironmentExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("open-questions loop environment must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[10].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[10].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[10].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("open-questions retirement requires separate owner review")
	}
}
