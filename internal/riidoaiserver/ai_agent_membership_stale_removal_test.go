package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMembershipStaleRemovalPreservesNewExecution(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	old := assignConversationThread(t, s, ctx, p, "membership-stale", "asn-old")
	markConversationThreadCompleted(s, "membership-stale", old.ThreadID, old.AgentID)
	newer, err := s.CreateAIAgentTaskAgentAssignment(ctx, p, "membership-stale", AssignAIAgentTaskRequest{AgentID: old.AgentID, AssignmentID: "asn-new", durableState: AssignmentRunning})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DeleteAIAgentTaskAgentAssignment(ctx, p, "membership-stale", old.AgentID, AgentAssignmentActionRequest{AssignmentID: old.AssignmentID}); !errors.Is(err, ErrAIAgentMembershipChanged) {
		t.Fatalf("stale delete error = %v", err)
	}
	// The HTTP fence must run before a durable cancellation, including when
	// the old durable assignment is absent and cancelling it would fail.
	durable := NewStoreWithConfig(StoreConfig{AgentRegistry: s})
	defer durable.Close()
	server := NewServer(ServerConfig{AIAgentClient: s, Assignment: durable, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, p.PrincipalID)}).Handler()
	for _, path := range []string{"/v1/client/ai-agent/tasks/membership-stale/assignment", "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/membership-stale/agent-assignments/" + old.AgentID} {
		body := `{"assignment_id":"asn-old"}`
		if strings.Contains(path, "/tasks/membership-stale/assignment") {
			body = `{"agent_id":"agent-owned-codex","assignment_id":"asn-old"}`
		}
		req := httptest.NewRequest(http.MethodDelete, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer ai-agent-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusConflict {
			t.Fatalf("stale HTTP delete = %d %s", resp.Code, resp.Body.String())
		}
	}
	if ids := membershipAgentIDs(t, s, p, "membership-stale"); len(ids) != 1 {
		t.Fatalf("stale delete removed member: %v", ids)
	}
	thread, ok := s.taskThreadForAssignmentLocked("membership-stale", old.AgentID, newer.AssignmentID)
	if !ok || !taskThreadHasActiveStream(thread) {
		t.Fatal("stale delete stopped new execution")
	}
	if _, err = s.DeleteAIAgentTaskAgentAssignment(ctx, p, "membership-stale", old.AgentID, AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	if ids := membershipAgentIDs(t, s, p, "membership-stale"); len(ids) != 0 {
		t.Fatalf("unscoped explicit removal = %v", ids)
	}
}
