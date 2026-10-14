package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

const knowledgeCoverageBehaviorGoldenSHA256 = "dd48e8846b1140991839d295cc01cae469ad40c47a0e71c373a57ba5ab1d9184"

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
