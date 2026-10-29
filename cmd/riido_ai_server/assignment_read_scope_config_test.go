package main

import (
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestAssignmentReadScopeConfigDefaultsOffAndRequiresAtomicPrivateCapture(t *testing.T) {
	t.Setenv(envAssignmentReadScope, "")
	if enabled, err := assignmentReadScopeFromEnv(nil, nil); err != nil || enabled {
		t.Fatalf("default=%v %v", enabled, err)
	}
	t.Setenv(envAssignmentReadScope, "on")
	reader, err := riidoaiserver.NewAIAgentPrivateTaskContextClient(riidoaiserver.AIAgentPrivateTaskContextClientConfig{BaseURL: "https://example.test", AssignmentReadScope: true})
	if err != nil {
		t.Fatal(err)
	}
	store := riidoaiserver.NewMemoryAssignmentContextStore()
	if enabled, err := assignmentReadScopeFromEnv(reader, store); err != nil || !enabled {
		t.Fatalf("configured capture=%v %v", enabled, err)
	}
	if _, err := assignmentReadScopeFromEnv(reader, nil); err == nil {
		t.Fatal("enabled without context storage")
	}
	if _, err := assignmentReadScopeFromEnv(nil, store); err == nil {
		t.Fatal("enabled without requester-private reader")
	}
	t.Setenv(envAssignmentReadScope, "maybe")
	if _, err := assignmentReadScopeFromEnv(reader, store); err == nil {
		t.Fatal("invalid capture gate accepted")
	}
}
