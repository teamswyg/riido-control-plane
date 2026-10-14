package main

import (
	"strings"
	"testing"
)

func TestChangeSummaryLinesReportsNoDiff(t *testing.T) {
	t.Parallel()
	op := operationRow{GeneratedPath: "aiAgent.same", Method: "GET", Path: "/same", OperationID: "same"}
	got := changeSummaryLines(previousManifest{Available: true, Operations: []operationRow{op}}, []operationRow{op})
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "추가 `0`, 제거 `0`, 변경 `0`") {
		t.Fatalf("summary = %q", joined)
	}
	if !strings.Contains(joined, "API surface diff는 없습니다") {
		t.Fatalf("summary missing no-diff line: %q", joined)
	}
}

func TestChangedOperationDescribesEveryChangedField(t *testing.T) {
	t.Parallel()
	before := operationRow{
		GeneratedPath:  "aiAgent.change",
		Method:         "GET",
		Path:           "/old",
		OperationID:    "oldOp",
		Lifecycle:      "active",
		Replacement:    "oldReplacement",
		RemovalHorizon: "2026-Q1",
	}
	after := operationRow{
		GeneratedPath:  "aiAgent.change",
		Method:         "POST",
		Path:           "/new",
		OperationID:    "newOp",
		Deprecated:     true,
		Lifecycle:      "deprecated",
		Replacement:    "newReplacement",
		RemovalHorizon: "2027-Q1",
	}
	got := describeChangedOperation(before, after)
	for _, want := range []string{
		"HTTP `GET /old` -> `POST /new`",
		"operationId `oldOp` -> `newOp`",
		"deprecated `false` -> `true`",
		"lifecycle `active` -> `deprecated`",
		"replacement `oldReplacement` -> `newReplacement`",
		"removal horizon `2026-Q1` -> `2027-Q1`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("changed operation = %q, missing %q", got, want)
		}
	}
}
