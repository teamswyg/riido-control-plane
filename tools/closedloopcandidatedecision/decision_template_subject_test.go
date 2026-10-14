package main

import (
	"encoding/json"
	"testing"
)

func ignoredCommandSubject(t *testing.T) rawSubject {
	t.Helper()
	body, err := json.Marshal(map[string]string{
		"kind":          "loop_refresh_ignored_command",
		"next_artifact": "verifier",
	})
	if err != nil {
		t.Fatal(err)
	}
	return rawSubject(body)
}

func ignoredCommandAdoptionPlan(command string) []adoptionStep {
	return []adoptionStep{
		{
			Artifact: "claim_binding",
			Command:  "go run ./tools/loopregistry -check-doc",
		},
		{
			Artifact: "verifier",
			Command:  command,
		},
	}
}
