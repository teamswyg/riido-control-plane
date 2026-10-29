package main

import (
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestAssignmentContextStoreFromEnv(t *testing.T) {
	t.Setenv(envAssignmentContext, "")
	if store, err := assignmentContextStoreFromEnv(nil); store != nil || err != nil {
		t.Fatalf("default must be off, got %T %v", store, err)
	}
	t.Setenv(envAssignmentContext, "on")
	if store, err := assignmentContextStoreFromEnv(nil); err != nil {
		t.Fatal(err)
	} else if _, ok := store.(*riidoaiserver.MemoryAssignmentContextStore); !ok {
		t.Fatalf("without DynamoDB the store must be in memory, got %T", store)
	}
	t.Setenv(envAssignmentContext, "maybe")
	if _, err := assignmentContextStoreFromEnv(nil); err == nil {
		t.Fatal("invalid value must fail startup")
	}
}
