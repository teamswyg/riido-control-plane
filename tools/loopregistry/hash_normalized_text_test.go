package main

import "testing"

func TestNormalizedTextIgnoresSemanticHashValues(t *testing.T) {
	left := normalizedText([]byte(`{"semantic_hash": "old", "statement": "same"}`))
	right := normalizedText([]byte(`{"semantic_hash": "new", "statement": "same"}`))
	if left != right {
		t.Fatalf("semantic_hash metadata changed normalized text:\nleft=%s\nright=%s", left, right)
	}
}

func TestNormalizedTextKeepsSemanticContent(t *testing.T) {
	left := normalizedText([]byte(`{"semantic_hash": "same", "statement": "old"}`))
	right := normalizedText([]byte(`{"semantic_hash": "same", "statement": "new"}`))
	if left == right {
		t.Fatal("semantic content change should alter normalized text")
	}
}
