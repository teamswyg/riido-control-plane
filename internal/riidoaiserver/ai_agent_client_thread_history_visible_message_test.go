package riidoaiserver

import "testing"

func TestAIAgentTaskThreadHistoryLocalizesStoppedContextCanceled(t *testing.T) {
	t.Parallel()
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

func TestAIAgentTaskThreadHistoryLocalizesStoredAgentContextCanceled(t *testing.T) {
	t.Parallel()
	messages := copyTaskThreadHistoryMessages([]AIAgentTaskThreadHistoryMessage{
		{
			MessageID:     "msg-agent-stopped",
			Role:          AIAgentTaskThreadMessageRoleAgent,
			CommentKind:   AgentTaskCommentStoppedByUserRequest,
			Body:          "context canceled",
			ResultMessage: "context canceled",
		},
		{
			MessageID: "msg-user-literal",
			Role:      AIAgentTaskThreadMessageRoleUser,
			Body:      "context canceled",
		},
	})
	if messages[0].Body != clientMessageTaskStopped {
		t.Fatalf("stored agent body = %q, want %q", messages[0].Body, clientMessageTaskStopped)
	}
	if messages[0].ResultMessage != clientMessageTaskStopped {
		t.Fatalf("stored agent result_message = %q, want %q", messages[0].ResultMessage, clientMessageTaskStopped)
	}
	if messages[1].Body != "context canceled" {
		t.Fatalf("user body = %q, want literal user text", messages[1].Body)
	}
}

func TestAIAgentTaskThreadHistoryProgressUsesClientVisibleText(t *testing.T) {
	t.Parallel()
	messages := taskThreadProgressMessages(AIAgentTaskThreadRecord{
		ThreadID:     "thread-progress-visible",
		AssignmentID: "asn-progress-visible",
		RunID:        "run-progress-visible",
		Lines: []AgentThreadProgressLine{
			{Seq: 1, Message: "생각 중...\n<ri"},
			{Seq: 2, Message: `<riido_log>{"code":1001,"args"`},
		},
	})
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Body != "생각 중..." {
		t.Fatalf("progress body = %q, want %q", messages[0].Body, "생각 중...")
	}
}
