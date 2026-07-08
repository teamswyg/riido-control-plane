package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

const loopRegistryBehaviorGoldenSHA256 = "25ce05e5d3e7d57aa6bd1c7935ec710854410b0348ebf30a23a3a42bc0dcf7f0"

var (
	loopRegistryDocSemanticHashPattern = regexp.MustCompile(
		"(?m)(\\| `[^`]+` \\| `[^`]+` \\| `[0-9]+` \\| `[0-9]+` \\| `)[0-9a-f]{12}(` \\|)$",
	)
	loopRegistryEvidenceHashPattern = regexp.MustCompile(`"[0-9a-f]{64}"`)
)

func TestLoopRegistryBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T12:00:00Z")
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatalf("claim hashes: %v", err)
	}
	result, err := verifyAll(root, m, hashes)
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
	sum.Write([]byte(normalizeLoopRegistryGoldenDoc(doc)))
	sum.Write([]byte("evidence\n"))
	sum.Write([]byte(normalizeLoopRegistryGoldenEvidence(string(evidenceJSON))))
	sum.Write([]byte("\n"))
	got := fmt.Sprintf("%x", sum.Sum(nil))
	if got != loopRegistryBehaviorGoldenSHA256 {
		t.Fatalf("loop registry behavior golden mismatch: got %s want %s", got, loopRegistryBehaviorGoldenSHA256)
	}
}

func normalizeLoopRegistryGoldenDoc(doc string) string {
	return loopRegistryDocSemanticHashPattern.ReplaceAllString(doc, "$1<semantic-hash>$2")
}

func normalizeLoopRegistryGoldenEvidence(body string) string {
	body = normalizeSemanticHashFields(body)
	return loopRegistryEvidenceHashPattern.ReplaceAllString(body, `"<semantic-hash>"`)
}
