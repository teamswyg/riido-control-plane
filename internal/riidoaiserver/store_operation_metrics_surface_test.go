package riidoaiserver

import (
	"context"
	"testing"
)

func TestStoreMetricsCaptureAssignmentOperationSurface(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil {
		t.Fatalf("poll = %+v", poll)
	}
	_, err = store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}

	snapshot, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	for _, operation := range assignmentOperationSurface() {
		metric := storeOperationMetricsByName(snapshot.StoreOperations)[operation.String()]
		if metric.CallsTotal != 1 || metric.ErrorsTotal != 0 || metric.LatencySamplesTotal != 1 {
			t.Fatalf("store operation %s metric = %+v", operation, metric)
		}
	}
	if snapshot.StoreOperationCallsTotal != 5 || snapshot.StoreOperationLatencySamplesTotal != 5 {
		t.Fatalf("store operation aggregate = %+v", snapshot)
	}
}

func assignmentOperationSurface() []StoreOperationName {
	return []StoreOperationName{
		StoreOperationCreateTask,
		StoreOperationCreateAssignment,
		StoreOperationPollAssignment,
		StoreOperationLeaseAssignment,
		StoreOperationAppendEvent,
	}
}
