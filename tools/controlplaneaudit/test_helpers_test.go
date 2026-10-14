package main

import (
	"encoding/json"
	"os"
	"testing"
)

func loadManifestForTest(t *testing.T) manifest {
	t.Helper()
	var m manifest
	if err := readJSON("../../"+defaultManifest, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func readEvidence(t *testing.T, path string) evidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
