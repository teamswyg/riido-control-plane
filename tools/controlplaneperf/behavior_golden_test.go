package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const controlPlanePerformanceGoldenSHA256 = "aa6fe66d67892597ed86b251cc30170f10af540853b963adc5102f041feea676"

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
