package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func (f *coverageLifecycleFixture) assign(t *testing.T) {
	t.Helper()
	if _, err := f.cli.actor.AssignTask(context.Background(), "coverage-task", AssignRequest{ComponentID: "coverage-component", AgentID: f.binding.AgentID, RuntimeProvider: f.binding.RuntimeProvider, Prompt: "synthetic lifecycle", CreatedBy: "non-owner"}); err != nil {
		t.Fatal(err)
	}
}

func (f *coverageLifecycleFixture) restart(t *testing.T) {
	t.Helper()
	f.cli.actor.Close()
	f.devices = openCoverageEnrollmentFixture(t, f.endpoint, false)
	operations, err := NewExecutionCoverageAssignmentStore(f.devices.snapshotStore, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = operations.Close() })
	f.operations = operations
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{f.binding})
	if err != nil {
		t.Fatal(err)
	}
	f.cli.actor = NewStoreWithConfig(StoreConfig{OperationStore: operations, AgentRegistry: registry})
	t.Cleanup(f.cli.actor.Close)
	f.cli.handler = NewServer(ServerConfig{Assignment: f.cli.actor, AIAgentClient: f.devices, Authorizer: f.authorizer, CLIInstallationService: f.cli.service}).Handler()
}
