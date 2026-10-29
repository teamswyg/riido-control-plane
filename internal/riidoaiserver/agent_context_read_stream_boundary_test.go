package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestContextReadStreamBoundsUpstream(t *testing.T) {
	s, reader, auth, event := contextReadStreamServer(t)
	reader.err = errors.New("upstream unavailable")
	checked := false
	s.config.Authorizer = authorizationFunc(func(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
		if req.TaskID != "" {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > DefaultAIAgentTaskContextClientTimeout {
				t.Fatal("read authorization and upstream lookup need a bounded context")
			}
			checked = true
		}
		return auth.Authorize(ctx, token, req)
	})
	body := contextReadStreamRequest(s, "live", event).Body.String()
	if !checked || body != "" {
		t.Fatalf("failed metadata read leaked: %s", body)
	}
}

func TestContextReadStreamPreservesDeletedAgentThreadVisibility(t *testing.T) {
	s, _, _, event := contextReadStreamServer(t)
	store := s.aiAgent.(*DevelopmentAIAgentClientStore)
	store.visibleTaskThreadLocked(AuthorizationResult{PrincipalID: "viewer", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task", "thread")
	delete(store.agents, "agent-owned-codex")
	body := contextReadStreamRequest(s, "replay", event).Body.String()
	if !strings.Contains(body, "agent_context_read") {
		t.Fatal("deleting a public agent must not hide its visible thread history")
	}
}

func TestPersistentContextReadDeliveryUsesExistingAgentPolicy(t *testing.T) {
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	event := ClientStreamEvent{EventType: AgentClientEventContextRead, Payload: AgentContextReadEvent{AgentID: "agent-owned-codex"}}
	saves := snapshots.saves
	if _, allowed := store.FilterAIAgentClientEvent(t.Context(), AuthorizationResult{PrincipalID: "user-1"}, event); !allowed {
		t.Fatal("owner must retain existing agent visibility")
	}
	if _, allowed := store.FilterAIAgentClientEvent(t.Context(), AuthorizationResult{PrincipalID: "viewer"}, event); allowed {
		t.Fatal("another viewer must not see a private agent")
	}
	if snapshots.saves != saves {
		t.Fatal("delivery visibility must not write the snapshot")
	}
}
