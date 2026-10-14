package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func newAssignableRepairServer(t *testing.T) (http.Handler, *DevelopmentAIAgentClientStore, *Store) {
	t.Helper()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-assignable-repair:read", "task:task-assignable-repair:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler(), aiAgentStore, assignmentStore
}

func pollAssignableRepairAssignment(t *testing.T, ctx context.Context, store *Store) PollResponse {
	t.Helper()
	poll, err := store.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil {
		t.Fatalf("poll response missing assignment: %+v", poll)
	}
	return poll
}
