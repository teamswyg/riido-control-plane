package main

import (
	"strings"
	"testing"
)

func TestDecodeModulesRejectsBadJSON(t *testing.T) {
	t.Parallel()
	_, err := decodeModules([]byte(`{"Path":`))
	if err == nil || !strings.Contains(err.Error(), "decode go list module") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestListModulesReportsGoListFailure(t *testing.T) {
	t.Parallel()
	_, err := listModules(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "go list -m -json all") {
		t.Fatalf("expected go list error, got %v", err)
	}
}
