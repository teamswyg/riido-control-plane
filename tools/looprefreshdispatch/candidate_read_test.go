package main

import (
	"encoding/json"
	"os"
	"testing"
)

func readCandidateEvidence(t *testing.T, path string) candidateEvidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got candidateEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
