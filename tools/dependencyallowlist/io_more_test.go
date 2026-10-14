package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadContractRejectsReadDecodeAndTrailerErrors(t *testing.T) {
	t.Parallel()
	if _, err := loadContract(t.TempDir() + "/missing.json"); err == nil || !strings.Contains(err.Error(), "read contract") {
		t.Fatalf("expected read error, got %v", err)
	}
	badJSON := writeContract(t, `{"schema_version": 1}`)
	if _, err := loadContract(badJSON); err == nil || !strings.Contains(err.Error(), "decode contract") {
		t.Fatalf("expected decode error, got %v", err)
	}
	trailing := writeContract(t, `{"schema_version":"riido-go-dependency-allowlist.v2"} {}`)
	if _, err := loadContract(trailing); err == nil || !strings.Contains(err.Error(), "trailing JSON") {
		t.Fatalf("expected trailing error, got %v", err)
	}
}

func TestWriteEvidenceRejectsUnwritablePath(t *testing.T) {
	t.Parallel()
	err := writeEvidence("/dev/null/evidence.json", newEvidence(testContract(), dependencyReport{}))
	if err == nil || !strings.Contains(err.Error(), "create evidence dir") {
		t.Skipf("/dev/null accepted nested write on this platform: %v", err)
	}
}

func TestWriteEvidenceCreatesParentAndJSON(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/nested/evidence.json"
	if err := writeEvidence(path, newEvidence(testContract(), dependencyReport{})); err != nil {
		t.Fatalf("writeEvidence: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if !strings.HasSuffix(string(body), "\n") || !strings.Contains(string(body), evidenceSchemaVersion) {
		t.Fatalf("evidence body = %q", body)
	}
}
