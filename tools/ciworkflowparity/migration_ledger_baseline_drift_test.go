package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsMigrationLedgerWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "migration-ledger.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("migration-ledger workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[5].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("migration-ledger digest substitution must fail closed")
	}
}

func TestVerifyRejectsMigrationLedgerAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[5].NativeMapping.MigrationLedger.AdapterRequired = true
	document.BoundedChildren[5].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("migration-ledger parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[5].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[5].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[5].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("migration-ledger retirement requires separate owner review")
	}
}
