package main

import (
	"reflect"
	"slices"
	"testing"
)

func TestIgnoredCommandSubjectBindingsCoverCommandFields(t *testing.T) {
	bound := map[string]bool{}
	for _, binding := range ignoredCommandSubjectBindings {
		bound[binding.CommandJSON] = true
	}
	for _, field := range jsonFieldNames(selectedRefreshCommand{}) {
		if !bound[field] {
			t.Fatalf("selectedRefreshCommand field %q has no subject binding", field)
		}
	}
}

func TestIgnoredCommandSubjectBindingsPointToSubjectFields(t *testing.T) {
	subjectFields := jsonFieldNames(candidateSubject{})
	for _, binding := range ignoredCommandSubjectBindings {
		if !slices.Contains(subjectFields, binding.SubjectJSON) {
			t.Fatalf("subject field %q missing for command field %q", binding.SubjectJSON, binding.CommandJSON)
		}
	}
}

func TestIgnoredCommandSubjectBindingsPreserveValues(t *testing.T) {
	command := selectedRefreshCommand{
		LoopID:                      "loop_one",
		Kind:                        "target_verifier",
		Command:                     "go test ./tools/looprefreshdispatch",
		CandidateID:                 "candidate_one",
		SubjectKind:                 "loop_refresh_ignored_command",
		DecisionSource:              "template",
		DecisionTemplateSubjectKind: "loop_refresh_ignored_command",
		ClaimIDs:                    []string{"claim_one"},
		EvidenceChainIDs:            []string{"chain_one"},
	}
	commandJSON := jsonObject(t, command)
	subjectJSON := jsonObject(t, ignoredCommandSubject(command))
	for _, binding := range ignoredCommandSubjectBindings {
		if !reflect.DeepEqual(commandJSON[binding.CommandJSON], subjectJSON[binding.SubjectJSON]) {
			t.Fatalf("%s -> %s mismatch: %v != %v",
				binding.CommandJSON,
				binding.SubjectJSON,
				commandJSON[binding.CommandJSON],
				subjectJSON[binding.SubjectJSON],
			)
		}
	}
}
