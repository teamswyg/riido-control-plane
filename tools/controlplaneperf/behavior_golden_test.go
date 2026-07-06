package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const controlPlanePerformanceGoldenSHA256 = "893b8d165c2f83e82a890b0b8b4bc898b2f2a8c7f8826341bc962a142685f3bf"

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
