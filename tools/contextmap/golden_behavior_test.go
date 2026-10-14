package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const (
	contextMapGoldenDocSHA256      = "44bee06ed8bb1cad09d76a4c765c9dc952a81f5277998d4e6dd0ca3f2d8fd3ee"
	contextMapGoldenEvidenceSHA256 = "e93c03f13b60c99cd6b459cac0edcaf884816094a9e1f5a18b7f87c7d9b888e1"
)

func TestContextMapBehaviorGolden(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(resolve(root, "docs/20-domain/context-map.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(doc); got != contextMapGoldenDocSHA256 {
		t.Fatalf("context-map doc golden hash = %s", got)
	}
	out := filepath.Join(t.TempDir(), "context-map-evidence.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	body, err := canonicalJSONFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(body); got != contextMapGoldenEvidenceSHA256 {
		t.Fatalf("context-map evidence golden hash = %s", got)
	}
}

func canonicalJSONFile(path string) ([]byte, error) {
	var value any
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
