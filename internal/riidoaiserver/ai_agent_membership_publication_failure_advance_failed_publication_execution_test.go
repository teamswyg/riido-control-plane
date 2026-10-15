package riidoaiserver

import (
	"context"
	"testing"
)

func advanceFailedPublicationExecution(t *testing.T, store *Store, assignment Assignment, state string) {
	t.Helper()
	ctx := context.Background()
	poll, err := store.PollAgent(ctx, assignment.AgentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != assignment.ID {
		t.Fatalf("pending poll=%+v err=%v", poll, err)
	}
	request := AgentEventRequest{AssignmentID: assignment.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentRunning, EventType: EventAssignmentRunning}
	mustRecordActorEvent(t, store, ctx, assignment.AgentID, request)
	if state == "completed" {
		request.State, request.EventType = AssignmentCompleted, EventAssignmentCompleted
	}
	if state == "failed" {
		request.State, request.EventType = AssignmentFailed, EventAssignmentFailed
	}
	if state != "running" {
		request.Message = "preserve terminal publication failure evidence"
		request.Metadata = map[string]string{"result": "preserve-me", "failure_code": "fixture-quota"}
		mustRecordActorEvent(t, store, ctx, assignment.AgentID, request)
	}
}
