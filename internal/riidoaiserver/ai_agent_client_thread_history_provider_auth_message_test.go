package riidoaiserver

import "testing"

func TestAIAgentTaskThreadHistoryLocalizesStoredProviderAuthFailure(t *testing.T) {
	t.Parallel()
	raw := "Failed to authenticate. API Error: 401 Invalid authentication credentials"
	messages := copyTaskThreadHistoryMessages([]AIAgentTaskThreadHistoryMessage{
		{
			MessageID:     "msg-agent-auth",
			Role:          AIAgentTaskThreadMessageRoleAgent,
			CommentKind:   AgentTaskCommentTaskFailed,
			Body:          raw,
			ResultMessage: raw,
		},
	})
	if messages[0].Body != clientMessageProviderAuthFailed {
		t.Fatalf("stored agent body = %q, want %q", messages[0].Body, clientMessageProviderAuthFailed)
	}
	if messages[0].ResultMessage != clientMessageProviderAuthFailed {
		t.Fatalf("stored agent result_message = %q, want %q", messages[0].ResultMessage, clientMessageProviderAuthFailed)
	}
}
