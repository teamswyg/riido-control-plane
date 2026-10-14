package main

import "testing"

func TestVerifierCommandsPreferClaimBoundGoPackages(t *testing.T) {
	claim := claimBinding{
		ID:        "claim",
		Verifiers: []string{"TestGeneratedDocMatchesManifest"},
		Files: []string{
			"tools/aiagentclientapi/types_thread_history_v3.go",
			"docs/30-architecture/loop-registry.md",
		},
	}
	tests := map[string][]string{
		"TestGeneratedDocMatchesManifest": {
			"./tools/aiagentclientapi",
			"./tools/configreference",
		},
	}
	commands := verifierCommandsForClaim(claim, nil, tests)
	want := "go test ./tools/aiagentclientapi -run '^(TestGeneratedDocMatchesManifest)$' -count=1"
	if len(commands) != 1 || commands[0] != want {
		t.Fatalf("commands = %#v, want %q", commands, want)
	}
}
