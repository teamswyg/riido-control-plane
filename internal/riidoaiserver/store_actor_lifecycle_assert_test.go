package riidoaiserver

import (
	"context"
	"testing"
)

func assertStoreActorLifecycleMetrics(t *testing.T, store *Store, ctx context.Context, completed AgentEventResponse) {
	t.Helper()
	if completed.Assignment == nil || completed.Assignment.State != AssignmentCompleted {
		t.Fatalf("completed event = %+v", completed)
	}
	metrics, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if metrics.TasksTotal != 1 || metrics.AssignmentsTotal != 1 || metrics.AssignmentsByState[AssignmentCompleted] != 1 {
		t.Fatalf("metrics assignments = %+v", metrics)
	}
	if metrics.PollRequestsTotal != 1 || metrics.PollActionsTotal[PollStart] != 1 {
		t.Fatalf("metrics poll = %+v", metrics)
	}
	if metrics.AgentEventsTotal != 2 || metrics.TaskEventsTotal != 4 {
		t.Fatalf("metrics events = %+v", metrics)
	}
}
