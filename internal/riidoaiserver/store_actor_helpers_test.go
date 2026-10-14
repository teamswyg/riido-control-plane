package riidoaiserver

import (
	"context"
	"testing"
)

func daemonPollRequest() PollRequest {
	return PollRequest{DaemonID: "daemon-1", RuntimeID: "runtime-1"}
}

func mustAssignActorTask(t *testing.T, store *Store, ctx context.Context, taskID string, req AssignRequest) Assignment {
	t.Helper()
	assignment, err := store.AssignTask(ctx, taskID, req)
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	return assignment
}

func mustAssignActorTaskAdditive(t *testing.T, store *Store, ctx context.Context, taskID string, req AssignRequest) Assignment {
	t.Helper()
	assignment, err := store.AssignTaskAdditive(ctx, taskID, req)
	if err != nil {
		t.Fatalf("AssignTaskAdditive: %v", err)
	}
	return assignment
}

func mustPollActor(t *testing.T, store *Store, ctx context.Context, agentID string) PollResponse {
	t.Helper()
	poll, err := store.PollAgent(ctx, agentID, daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent %s: %v", agentID, err)
	}
	return poll
}

func mustRecordActorEvent(t *testing.T, store *Store, ctx context.Context, agentID string, req AgentEventRequest) AgentEventResponse {
	t.Helper()
	response, err := store.RecordAgentEvent(ctx, agentID, req)
	if err != nil {
		t.Fatalf("RecordAgentEvent %s: %v", req.EventType, err)
	}
	return response
}

func mustCancelActorAssignment(t *testing.T, store *Store, ctx context.Context, taskID string, req CancelAssignmentRequest) Assignment {
	t.Helper()
	assignment, err := store.CancelAssignment(ctx, taskID, req)
	if err != nil {
		t.Fatalf("CancelAssignment: %v", err)
	}
	return assignment
}

func mustLoadActorProjection(t *testing.T, store *Store, ctx context.Context, assignmentID string) AssignmentProjection {
	t.Helper()
	projection, ok, err := store.LoadAssignmentProjection(ctx, assignmentID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection %s: %v", assignmentID, err)
	}
	if !ok {
		t.Fatalf("LoadAssignmentProjection %s: not found", assignmentID)
	}
	return projection
}
