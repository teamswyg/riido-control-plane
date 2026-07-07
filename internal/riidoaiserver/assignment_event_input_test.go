package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestNewAssignmentEventInputPrefersEventValuesAndDefaultsRunning(t *testing.T) {
	at := time.Date(2026, 7, 8, 3, 10, 0, 0, time.UTC)
	input, err := newAssignmentEventInput(" agent-1 ", AgentEventRequest{
		TaskID:       " task-from-request ",
		AssignmentID: " asn-from-request ",
		State:        AssignmentQueued,
		Message:      " request message ",
	}, TaskEvent{
		TaskID:       " task-from-event ",
		AssignmentID: " asn-from-event ",
		Type:         " riido_log ",
		Message:      " event message ",
		Metadata:     map[string]string{"k": "v"},
		At:           at,
	})
	if err != nil {
		t.Fatalf("newAssignmentEventInput: %v", err)
	}
	if input.AgentID != "agent-1" || input.TaskID != "task-from-event" ||
		input.AssignmentID != "asn-from-event" || input.State != AssignmentQueued ||
		input.Type != EventRiidoLog || input.Message != "event message" ||
		input.Metadata["k"] != "v" || !input.At.Equal(at) {
		t.Fatalf("input = %+v", input)
	}
}

func TestNewAssignmentEventInputFallsBackAndRequiresIdentities(t *testing.T) {
	input, err := newAssignmentEventInput("agent-1", AgentEventRequest{
		TaskID:       "task-from-request",
		AssignmentID: "asn-from-request",
		Message:      "request message",
	}, TaskEvent{})
	if err != nil {
		t.Fatalf("newAssignmentEventInput fallback: %v", err)
	}
	if input.State != AssignmentRunning || input.Message != "request message" {
		t.Fatalf("fallback input = %+v", input)
	}

	for _, tt := range []struct {
		name    string
		agentID string
		req     AgentEventRequest
		wantErr string
	}{
		{"agent", " ", AgentEventRequest{TaskID: "task", AssignmentID: "asn"}, "agent_id"},
		{"task", "agent", AgentEventRequest{AssignmentID: "asn"}, "task_id"},
		{"assignment", "agent", AgentEventRequest{TaskID: "task"}, "assignment_id"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newAssignmentEventInput(tt.agentID, tt.req, TaskEvent{}); err == nil ||
				!strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
