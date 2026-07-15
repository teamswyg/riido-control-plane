package riidoaiserver

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestAgentEventErrorCategory(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{err: errors.New("assignment asn-1 not found"), want: agentEventErrorCategoryAssignmentNotFound},
		{err: errors.New("assignment asn-1 belongs to agent agent-a"), want: agentEventErrorCategoryAgentMismatch},
		{err: errors.New("assignment asn-1 belongs to task task-a"), want: agentEventErrorCategoryTaskMismatch},
		{err: errors.New("invalid assignment transition completed -> running"), want: agentEventErrorCategoryInvalidTransition},
		{err: newAgentBindingValidationErrorf("agent agent-a is not registered"), want: agentEventErrorCategoryBindingValidation},
		{err: errors.New("dynamodb put assignment operation: unavailable"), want: agentEventErrorCategoryStoreFailure},
		{err: errors.New("assignment_id is required"), want: agentEventErrorCategoryBadRequest},
	}
	for _, tt := range tests {
		if got := agentEventErrorCategory(tt.err); got != tt.want {
			t.Fatalf("agentEventErrorCategory() = %q, want %q", got, tt.want)
		}
	}
}

func TestLogAgentEventRejectedExcludesSensitivePayload(t *testing.T) {
	var buf bytes.Buffer
	previousWriter, previousFlags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
	})

	logAgentEventRejected(" agent-a ", AgentEventRequest{
		AssignmentID: " asn-1 ",
		EventType:    EventRiidoLog,
		State:        AssignmentRunning,
		DaemonID:     "daemon-private",
		RuntimeID:    "runtime-private",
		Message:      "prompt-or-provider-output",
		Metadata:     map[string]string{"token": "super-secret"},
	}, errors.New("invalid assignment transition completed -> running"))

	output := buf.String()
	for _, want := range []string{
		"event=agent_event_rejected",
		`route=/v1/agents/{agent_id}/events`,
		`agent_id="agent-a"`,
		`assignment_id="asn-1"`,
		`event_type="riido_log"`,
		`requested_state="running"`,
		`event_error_category="invalid_state_transition"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("log output %q does not contain %q", output, want)
		}
	}
	for _, forbidden := range []string{
		"daemon-private",
		"runtime-private",
		"prompt-or-provider-output",
		"super-secret",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("log output leaked %q: %s", forbidden, output)
		}
	}
}
