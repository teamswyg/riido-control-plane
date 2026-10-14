package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type membershipRemovalHTTPGate struct {
	*DevelopmentAIAgentClientStore
	entered chan struct{}
	resume  chan struct{}
}

func (s *membershipRemovalHTTPGate) UnassignAIAgentTask(ctx context.Context, p AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	close(s.entered)
	<-s.resume
	return s.DevelopmentAIAgentClientStore.UnassignAIAgentTask(ctx, p, taskID, req)
}

func (s *membershipRemovalHTTPGate) DeleteAIAgentTaskAgentAssignment(ctx context.Context, p AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	close(s.entered)
	<-s.resume
	return s.DevelopmentAIAgentClientStore.DeleteAIAgentTaskAgentAssignment(ctx, p, taskID, agentID, req)
}

func TestHTTPMembershipRemovalRejectsInterleavedAssignment(t *testing.T) {
	for _, version := range []string{"v1", "v2"} {
		for _, history := range []string{"queued", "observed", "absent"} {
			t.Run(version+"/"+history, func(t *testing.T) {
				checkMembershipRemovalHTTPInterleave(t, version, history)
			})
		}
	}
}

func checkMembershipRemovalHTTPInterleave(t *testing.T, version, history string) {
	t.Helper()
	const taskID, agentID = "task-removal-fence", "agent-public-openclaw"
	ctx := context.Background()
	client := &membershipRemovalHTTPGate{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), entered: make(chan struct{}), resume: make(chan struct{})}
	assignment := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	defer assignment.Close()
	auth := aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")
	// The standard helper uses this credential; keep the public HTTP path real.
	const requestToken = "ai-agent-token"
	handler := NewServer(ServerConfig{AIAgentClient: client, Assignment: assignment, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: auth}).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	var observed string
	if history != "absent" {
		observed = createProviderMultiAssignment(t, handler, base, requestToken, taskID, agentID).AssignmentID
	} else {
		// Membership can outlive retained execution history.
		client.mu.Lock()
		client.assignTaskAgentMemberLocked(taskID, client.agents[agentID])
		client.mu.Unlock()
	}
	path := base + "/tasks/" + taskID + "/agent-assignments/" + agentID
	body := `{"reason":"participant_removed"}`
	if version == "v1" {
		path = "/v1/client/ai-agent/tasks/" + taskID + "/assignment"
		body = `{"agent_id":"` + agentID + `","reason":"participant_removed"}`
	}
	if history == "observed" {
		body = strings.TrimSuffix(body, "}") + `,"assignment_id":"` + observed + `"}`
	}
	req := httptest.NewRequest(http.MethodDelete, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+requestToken)
	recorder := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { handler.ServeHTTP(recorder, req); close(finished) }()
	resumed := false
	defer func() {
		if !resumed {
			close(client.resume)
		}
	}()
	select {
	case <-client.entered:
	case <-finished:
		t.Fatalf("DELETE failed before cancellation: status=%d body=%s", recorder.Code, recorder.Body.String())
	case <-time.After(3 * time.Second):
		t.Fatalf("DELETE did not reach the post-cancellation barrier: status=%d", recorder.Code)
	}
	next := createProviderMultiAssignment(t, handler, base, requestToken, taskID, agentID)
	close(client.resume)
	resumed = true
	<-finished
	if recorder.Code != http.StatusConflict {
		t.Fatalf("interleaved execution must reject removal: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if got := membershipAgentIDs(t, client, p, taskID); len(got) != 1 || got[0] != agentID {
		t.Fatalf("conflict removed the new participant: %v", got)
	}
	thread := taskThreadByAssignment(t, client.taskThreads[taskID], next.AssignmentID)
	if !taskThreadHasActiveStream(thread) {
		t.Fatalf("conflict falsely stopped the new execution: %+v", thread)
	}
	poll, err := assignment.PollAgent(ctx, agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID {
		t.Fatalf("failed removal should preserve the new execution: poll=%+v err=%v", poll, err)
	}
}
