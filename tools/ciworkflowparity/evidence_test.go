package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvidenceIsMode0600AndIdentifierFree(t *testing.T) {
	result, err := verify("../..", repositoryContract)
	if err != nil {
		t.Fatalf("%v: %+v", err, result)
	}
	path := filepath.Join(t.TempDir(), "private", "parity.json")
	if err := writeEvidence(path, result); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected evidence mode: %v %v", info, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{"arn:aws", "account_id", "role_arn", "cookie", "bearer", "gho_", "akia"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("evidence contains forbidden value %q", forbidden)
		}
	}
}
