package main

import (
	"strings"
	"testing"
)

func TestLifecycleAttrsIncludesEveryDeclaredField(t *testing.T) {
	t.Parallel()
	got := lifecycleAttrs(operationRow{
		Deprecated:     true,
		Lifecycle:      "sunset",
		Replacement:    "newPath",
		RemovalHorizon: "2026-Q4",
	})
	want := []string{
		"deprecated",
		"lifecycle=sunset",
		"replacement=newPath",
		"removal_horizon=2026-Q4",
	}
	for i, value := range want {
		if got[i] != value {
			t.Fatalf("attr[%d] = %q, want %q", i, got[i], value)
		}
	}
}

func TestRenderPRBodyLifecycleOmitsEmptyLifecycleSection(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	renderPRBodyLifecycle(&b, []operationRow{{GeneratedPath: "aiAgent.active"}})
	if strings.Contains(b.String(), "Lifecycle / Deprecation") {
		t.Fatalf("unexpected lifecycle section: %q", b.String())
	}
}

func TestOperationLifecycleFieldsIncludesReplacementOnlyFields(t *testing.T) {
	t.Parallel()
	got := operationLifecycleFields(operationRow{
		Replacement:    "replacement.path",
		RemovalHorizon: "2027-01",
	})
	for _, want := range []string{
		"replacement: 'replacement.path'",
		"removalHorizon: '2027-01'",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("fields = %q, missing %q", got, want)
		}
	}
}
