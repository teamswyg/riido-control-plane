package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const controlPlanePerformanceGoldenSHA256 = "7e7ec0d3f6f14990b0505b7d785a5ed5fe1e82e3578cd6ad8be0fbe786abe888"

func TestControlPlanePerformanceBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-evidence-out", out,
	})
	if err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != controlPlanePerformanceGoldenSHA256 {
		t.Fatalf("evidence SHA mismatch: got %s want %s\n%s", got, controlPlanePerformanceGoldenSHA256, body)
	}
}
