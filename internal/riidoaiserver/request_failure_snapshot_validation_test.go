package riidoaiserver

import (
	"errors"
	"fmt"
	"testing"
)

func TestRequestFailureSeparatesSnapshotValidationFromUnknownDependency(t *testing.T) {
	if got := boundedRequestFailureType(fmt.Errorf("private wrapper: %w", invalidRuntimeSnapshot(errors.New("private value")))); got != "invalid_runtime_snapshot" {
		t.Fatalf("snapshot validation diagnostic=%q", got)
	}
	if got := boundedRequestFailureType(errors.New("private upstream body")); got != "unknown" {
		t.Fatalf("unclassified dependency was guessed=%q", got)
	}
}
