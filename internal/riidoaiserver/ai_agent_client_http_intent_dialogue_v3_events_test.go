package riidoaiserver

import (
	"context"
	"testing"
)

func recordIntentDialogueRunning(
	t *testing.T,
	assignmentStore *Store,
	aiStore *DevelopmentAIAgentClientStore,
	action AIAgentTaskActionResponse,
) {
	t.Helper()
	recordIntentDialogueDurableEvent(t, assignmentStore, aiStore, action, AssignmentRunning, EventAssignmentRunning, "agent work is running")
}

func recordIntentDialogueCompleted(
	t *testing.T,
	assignmentStore *Store,
	aiStore *DevelopmentAIAgentClientStore,
	action AIAgentTaskActionResponse,
	message string,
) {
	t.Helper()
	recordIntentDialogueDurableEvent(t, assignmentStore, aiStore, action, AssignmentCompleted, EventAssignmentCompleted, message)
}

func recordIntentDialogueFailed(
	t *testing.T,
	assignmentStore *Store,
	aiStore *DevelopmentAIAgentClientStore,
	action AIAgentTaskActionResponse,
	message string,
) {
	t.Helper()
	recordIntentDialogueDurableEvent(t, assignmentStore, aiStore, action, AssignmentFailed, EventAssignmentFailed, message)
}

func recordIntentDialogueDurableEvent(
	t *testing.T,
	assignmentStore *Store,
	aiStore *DevelopmentAIAgentClientStore,
	action AIAgentTaskActionResponse,
	state AssignmentState,
	eventType string,
	message string,
) {
	t.Helper()
	req := AgentEventRequest{
		AssignmentID: action.AssignmentID,
		TaskID:       action.TaskID,
		DaemonID:     "daemon-dev-macbook",
		DeviceID:     "device-dev-macbook",
		RuntimeID:    "runtime-codex-dev",
		State:        state,
		EventType:    eventType,
		Message:      message,
	}
	response, err := assignmentStore.RecordAgentEvent(context.Background(), action.AgentID, req)
	if err != nil {
		t.Fatalf("RecordAgentEvent %s: %v", eventType, err)
	}
	if err := aiStore.RecordAIAgentAssignmentEvent(context.Background(), action.AgentID, req, response.Event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent %s: %v", eventType, err)
	}
}
