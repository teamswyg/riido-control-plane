package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientWorkspaceAgentStopClosesVisibleThreads(t *testing.T) {
	const token = "owner-token"
	const agentID = "agent-public-openclaw"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	first := createProviderMultiAssignment(t, server, base, token, "task-workspace-stop-a", agentID)
	second := createProviderMultiAssignment(t, server, base, token, "task-workspace-stop-b", agentID)

	stopBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agent-assignments/"+agentID+"/stop", token, `{"reason":"user_request"}`, http.StatusAccepted)
	var stopped AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, stopBytes, &stopped)
	if stopped.AgentID != agentID || stopped.WorkStatus != AgentWorkStatusIdle || stopped.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("workspace stop response = %+v", stopped)
	}

	assertWorkspaceStopThreadClosed(t, server, base, token, first)
	assertWorkspaceStopThreadClosed(t, server, base, token, second)
}

func assertWorkspaceStopThreadClosed(t *testing.T, server http.Handler, base, token string, assigned AIAgentTaskActionResponse) {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+assigned.TaskID+"/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, bytes, &threads)
	thread := taskThreadByAssignment(t, threads.Threads, assigned.AssignmentID)
	if taskThreadHasActiveStream(thread) || thread.WorkStatus != AgentWorkStatusIdle || thread.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("thread should be strongly stopped: %+v", thread)
	}
}
