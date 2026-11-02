package riidoaiserver

import (
	"context"
	"testing"
)

func replicaUserMessage(t *testing.T, store *PersistentAIAgentClientStore, owner AuthorizationResult, body string) (string, string) {
	t.Helper()
	history, err := store.ListAIAgentTaskThreadHistory(context.Background(), owner, "task-replica")
	if err != nil {
		t.Fatal(err)
	}
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			if message.Role == AIAgentTaskThreadMessageRoleUser && message.Body == body {
				return thread.ThreadID, message.MessageID
			}
		}
	}
	t.Fatalf("missing user message %q", body)
	return "", ""
}
