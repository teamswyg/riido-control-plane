package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsHarnessPromotionWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "harness-promotion.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("harness-promotion workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[11].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("harness-promotion digest substitution must fail closed")
	}
}

func TestVerifyRejectsHarnessPromotionAdapterEnvironmentOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[11].NativeMapping.HarnessPromotion.AdapterRequired = true
	document.BoundedChildren[11].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("harness-promotion parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[11].ParityClaim.HarnessPromotionLoopEnvironmentExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("harness-promotion loop environment must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[11].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[11].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[11].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("harness-promotion retirement requires separate owner review")
	}
}
