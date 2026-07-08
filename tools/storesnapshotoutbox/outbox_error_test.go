package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadOutboxEventsRejectsMissingAndMalformedRecords(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, err := readOutboxEvents(filepath.Join(root, "missing.jsonl")); err == nil {
		t.Fatal("readOutboxEvents accepted missing file")
	}
	path := filepath.Join(root, "outbox.jsonl")
	if err := os.WriteFile(path, []byte("{bad-json}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readOutboxEvents(path); err == nil {
		t.Fatal("readOutboxEvents accepted malformed record")
	}
}

func TestOutboxCaseRejectsExpectationDrift(t *testing.T) {
	t.Parallel()
	tc := testCaseByKind(t, "outbox")
	tc.Name = "outbox-expectation-drift"
	tc.WantRecords = 99
	if _, err := verifyOutboxCase(tc); err == nil {
		t.Fatal("verifyOutboxCase accepted record drift")
	}
}

func TestOutboxFailureCaseRejectsMetricDrift(t *testing.T) {
	t.Parallel()
	tc := testCaseByKind(t, "outbox-failure")
	tc.WantOutboxErrors = 99
	if _, err := verifyOutboxFailureCase(tc); err == nil {
		t.Fatal("verifyOutboxFailureCase accepted metric drift")
	}
}
