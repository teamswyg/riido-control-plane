package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const configReferenceGoldenSHA256 = "4fd4b7faae96500e91863e476f1163e38701a822ebdfac741bd4245d074f00d6"

func TestConfigReferenceEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.Status != "verified" || got.RuntimeEnvCount != got.ManifestCount {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.RuntimeEnvCount < 40 || got.SecretCount == 0 || got.OperatorCount == 0 {
		t.Fatalf("weak config evidence: %+v", got)
	}
}

func TestConfigReferenceBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != configReferenceGoldenSHA256 {
		t.Fatalf("config reference golden hash = %s", got)
	}
}

func TestConfigReferenceGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}
