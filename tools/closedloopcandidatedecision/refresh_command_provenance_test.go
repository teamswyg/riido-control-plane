package main

import "testing"

func TestCandidateDecisionRefreshCommandsCarryDecisionProvenance(t *testing.T) {
	command := "go test ./tools/looprefreshdispatch"
	result := ignoredCommandTemplateResult(t, command)
	got := newRefreshCommandEvidence(result)
	if got.CommandCount != 1 {
		t.Fatalf("command count = %d", got.CommandCount)
	}
	item := got.Commands[0]
	if item.DecisionSource != decisionSourceTemplate ||
		item.DecisionTemplateSubjectKind != "loop_refresh_ignored_command" {
		t.Fatalf("command provenance = %+v", item)
	}
	if item.Command != command {
		t.Fatalf("command = %+v", item)
	}
}
