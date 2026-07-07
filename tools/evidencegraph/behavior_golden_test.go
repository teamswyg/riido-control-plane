package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

const evidenceGraphBehaviorGoldenSHA256 = "2a7c587e2c99fb1779e9c23b6e79b2647f5fb7e7425b4175bc0ecbf64161d107"

func TestEvidenceGraphBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")

	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatalf("verify manifest: %v", err)
	}
	doc := renderDoc(m, result)
	evidenceJSON, err := json.MarshalIndent(newEvidence(m, result, nil), "", "  ")
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}

	sum := sha256.New()
	sum.Write([]byte("doc\n"))
	sum.Write([]byte(doc))
	sum.Write([]byte("evidence\n"))
	sum.Write(append(evidenceJSON, '\n'))
	got := fmt.Sprintf("%x", sum.Sum(nil))
	if got != evidenceGraphBehaviorGoldenSHA256 {
		t.Fatalf("evidence graph behavior golden mismatch: got %s want %s", got, evidenceGraphBehaviorGoldenSHA256)
	}
}
