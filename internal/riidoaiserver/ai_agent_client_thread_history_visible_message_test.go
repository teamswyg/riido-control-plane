package riidoaiserver

import "testing"

func TestAIAgentTaskThreadHistoryLocalizesStoppedContextCanceled(t *testing.T) {
	message, ok := taskThreadProjectionMessage(AIAgentTaskThreadRecord{
		ThreadID:        "thread-stopped-context",
		TaskID:          "task-stopped-context",
		AssignmentID:    "asn-stopped-context",
		AgentID:         "agent-public-openclaw",
		RunID:           "run-stopped-context",
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     AgentTaskCommentStoppedByUserRequest,
		Message:         "context canceled",
		ResultMessage:   "context canceled",
	})
	if !ok {
		t.Fatal("expected stopped projection message")
	}
	if message.Body != clientMessageTaskStopped {
		t.Fatalf("body = %q, want %q", message.Body, clientMessageTaskStopped)
	}
	if message.ResultMessage != clientMessageTaskStopped {
		t.Fatalf("result_message = %q, want %q", message.ResultMessage, clientMessageTaskStopped)
	}
}
