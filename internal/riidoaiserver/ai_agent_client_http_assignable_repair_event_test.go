package riidoaiserver

import (
	"context"
	"testing"
)

func recordAssignableRepairCompleted(t *testing.T, ctx context.Context, store *Store, assignmentID string) {
	t.Helper()
	for _, event := range []AgentEventRequest{
		{
			AssignmentID: assignmentID,
			TaskID:       "task-assignable-repair",
			DaemonID:     "daemon-shared-studio",
			DeviceID:     "device-shared-studio",
			RuntimeID:    "runtime-openclaw-shared",
			State:        AssignmentRunning,
			EventType:    EventAssignmentRunning,
			Message:      "running before terminal projection",
		},
		{
			AssignmentID: assignmentID,
			TaskID:       "task-assignable-repair",
			DaemonID:     "daemon-shared-studio",
			DeviceID:     "device-shared-studio",
			RuntimeID:    "runtime-openclaw-shared",
			State:        AssignmentCompleted,
			EventType:    EventAssignmentCompleted,
			Message:      "completed before assignable-agents repaired the client read-model",
		},
	} {
		if _, err := store.RecordAgentEvent(ctx, "agent-public-openclaw", event); err != nil {
			t.Fatalf("RecordAgentEvent %s: %v", event.EventType, err)
		}
	}
}
