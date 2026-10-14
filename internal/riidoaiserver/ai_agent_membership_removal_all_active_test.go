package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type membershipRemovalCancelRecorder struct {
	*Store
	ids []string
}

func (s *membershipRemovalCancelRecorder) CancelAssignment(_ context.Context, _ string, req CancelAssignmentRequest) (Assignment, error) {
	s.ids = append(s.ids, req.AssignmentID)
	return Assignment{ID: req.AssignmentID, State: AssignmentCancelled}, nil
}

func TestHTTPMembershipRemovalObservedIDCancelsAllRetainedActiveExecutions(t *testing.T) {
	client := NewDevelopmentAIAgentClientStore()
	ctx := context.Background()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	const taskID = "task-restored-active-membership"
	first := assignConversationThread(t, client, ctx, p, taskID, "asn-first")
	// Restored history can contain multiple active executions for one member.
	client.mu.Lock()
	latest := client.taskThreads[taskID][0]
	latest.ThreadID = "thread-latest"
	latest.AssignmentID = "asn-latest"
	peer := latest
	peer.ThreadID = "thread-peer"
	peer.AssignmentID = "asn-peer"
	peer.AgentID = "agent-public-openclaw"
	client.taskThreads[taskID] = append(client.taskThreads[taskID], latest, peer)
	client.assignTaskAgentMemberLocked(taskID, client.agents[peer.AgentID])
	client.taskThreads["other-task"] = []AIAgentTaskThreadRecord{latest}
	client.mu.Unlock()
	canceller := &membershipRemovalCancelRecorder{Store: NewStore()}
	defer canceller.Close()
	handler := NewServer(ServerConfig{AIAgentClient: client, Assignment: canceller, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, p.PrincipalID)}).Handler()
	path := "/v2/client/workspaces/" + p.WorkspaceID + "/ai-agent/tasks/" + taskID + "/agent-assignments/" + first.AgentID
	response := serveUnassignTaskBoundary(handler, path, "ai-agent-token", `{"assignment_id":"asn-latest","reason":"participant_removed"}`)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if len(canceller.ids) != 2 || canceller.ids[0] != "asn-latest" || canceller.ids[1] != "asn-first" {
		t.Fatalf("cancelled %v; want both member executions only", canceller.ids)
	}
	for _, thread := range client.taskThreads[taskID] {
		if thread.AgentID == first.AgentID && taskThreadHasActiveStream(thread) {
			t.Fatalf("removed member still active: %+v", thread)
		}
		if thread.AgentID == peer.AgentID && !taskThreadHasActiveStream(thread) {
			t.Fatal("other member was stopped")
		}
	}
	if !taskThreadHasActiveStream(client.taskThreads["other-task"][0]) {
		t.Fatal("other task was stopped")
	}
	if ids := membershipAgentIDs(t, client, p, taskID); len(ids) != 1 || ids[0] != peer.AgentID {
		t.Fatalf("remaining members=%v", ids)
	}
}
