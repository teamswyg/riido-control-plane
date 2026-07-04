package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const controlPlaneAuditGoldenSHA256 = "9fac81bc2e343107f90cd2c9c14241733239387af321ecd0cc896169e0353524"

func TestControlPlaneAuditBehaviorGolden(t *testing.T) {
	repo := t.TempDir()
	writeAuditFixture(t, repo)
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	out := filepath.Join(repo, "evidence.json")
	err := mainRun([]string{"-repo", repo, "-manifest", "manifest.json", "-evidence-out", out})
	if err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != controlPlaneAuditGoldenSHA256 {
		t.Fatalf("evidence SHA mismatch: got %s want %s\n%s", got, controlPlaneAuditGoldenSHA256, body)
	}
}

func writeAuditFixture(t *testing.T, root string) {
	t.Helper()
	mustWriteAuditFixture(t, root, "go.mod", "module example.com/audit\n\ngo 1.21\n")
	mustWriteAuditFixture(t, root, "internal/server/hot.go", controlPlaneAuditFixtureSource)
	mustWriteAuditFixture(t, root, ".github/workflows/control-plane-performance.yml", controlPlaneAuditFixtureWorkflow)
	mustWriteAuditFixture(t, root, "manifest.json", controlPlaneAuditFixtureManifest)
}

func mustWriteAuditFixture(t *testing.T, root, slashPath, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(slashPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", slashPath, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", slashPath, err)
	}
}
