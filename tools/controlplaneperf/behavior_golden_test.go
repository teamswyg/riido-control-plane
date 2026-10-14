package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const controlPlanePerformanceGoldenSHA256 = "9e45c92c8bea6fec85c75111a78bbf90443f3179a4261ec55027c24fa4fb8492"

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
