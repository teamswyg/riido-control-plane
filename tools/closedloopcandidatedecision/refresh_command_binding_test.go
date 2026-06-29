package main

import (
	"reflect"
	"slices"
	"testing"
)

func TestRefreshCommandBindingsCoverCommandFields(t *testing.T) {
	bound := map[string]bool{}
	for _, binding := range refreshCommandBindings {
		bound[binding.CommandJSON] = true
	}
	for _, field := range jsonFieldNames(selectedRefreshCommand{}) {
		if !bound[field] {
			t.Fatalf("selectedRefreshCommand field %q has no binding", field)
		}
	}
}

func TestRefreshCommandBindingsPointToArtifactFields(t *testing.T) {
	artifactFields := jsonFieldNames(decisionArtifactEvidence{})
	for _, binding := range refreshCommandBindings {
		if binding.Source == refreshBindingSubject {
			continue
		}
		if !slices.Contains(artifactFields, binding.ArtifactJSON) {
			t.Fatalf("artifact field %q missing for command field %q", binding.ArtifactJSON, binding.CommandJSON)
		}
	}
}

func TestRefreshCommandBindingsPreserveValues(t *testing.T) {
	result := ignoredCommandTemplateResult(t, "go test ./tools/looprefreshdispatch")
	commandJSON := jsonObject(t, newRefreshCommandEvidence(result).Commands[0])
	artifactJSON := jsonObject(t, result.DecisionArtifacts[0])
	for _, binding := range refreshCommandBindings {
		if !refreshBindingMatches(binding, commandJSON, artifactJSON) {
			t.Fatalf("binding mismatch %+v: command=%v artifact=%v",
				binding,
				commandJSON[binding.CommandJSON],
				artifactJSON[binding.ArtifactJSON],
			)
		}
	}
}

func refreshBindingMatches(
	binding refreshCommandBinding,
	commandJSON map[string]any,
	artifactJSON map[string]any,
) bool {
	switch binding.Source {
	case refreshBindingCopy:
		return reflect.DeepEqual(commandJSON[binding.CommandJSON], artifactJSON[binding.ArtifactJSON])
	case refreshBindingDerive:
		command, ok := artifactJSON[binding.ArtifactJSON].(string)
		return ok && commandJSON[binding.CommandJSON] == refreshCommandKind(command)
	case refreshBindingSubject:
		return commandJSON[binding.CommandJSON] == "loop_refresh_ignored_command"
	default:
		return false
	}
}
