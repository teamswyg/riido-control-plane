package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

const knowledgeCoverageBehaviorGoldenSHA256 = "d4b5debd7a3edfe8eab3377ea89c5ba87fb3ca89bb34abeffcf144df45d8434f"

func TestKnowledgeCoverageBehaviorGolden(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	m, err := loadManifest(resolvePath(root, "docs/executable-knowledge.riido.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if problems := validateManifest(root, m); len(problems) > 0 {
		t.Fatalf("manifest problems: %v", problems)
	}
	docs, problems := scanDocs(root, m)
	if len(problems) > 0 {
		t.Fatalf("scan problems: %v", problems)
	}
	doc := renderDoc(root, m, docs, problems)
	evidenceJSON, err := json.MarshalIndent(buildEvidence(root, m, docs, problems), "", "  ")
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}

	sum := sha256.New()
	sum.Write([]byte("doc\n"))
	sum.Write([]byte(doc))
	sum.Write([]byte("evidence\n"))
	sum.Write(append(evidenceJSON, '\n'))
	got := fmt.Sprintf("%x", sum.Sum(nil))
	if got != knowledgeCoverageBehaviorGoldenSHA256 {
		t.Fatalf("knowledge coverage behavior golden mismatch: got %s want %s", got, knowledgeCoverageBehaviorGoldenSHA256)
	}
}
