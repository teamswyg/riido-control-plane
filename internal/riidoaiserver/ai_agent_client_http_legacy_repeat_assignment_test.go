package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientLegacyAssignmentRepeatsSameAgentAsNewThread(t *testing.T) {
	const token = "owner-token"
	const taskID = "task-legacy-repeat-agent"
	const agentID = "agent-public-openclaw"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	first := legacyAssignSameAgent(t, server, base, token, taskID, agentID)
	second := legacyAssignSameAgent(t, server, base, token, taskID, agentID)
	if first.AssignmentID == second.AssignmentID || first.ThreadID == second.ThreadID {
		t.Fatalf("legacy repeated assignment collapsed: first=%+v second=%+v", first, second)
	}

	threadsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, threadsBytes, &threads)
	if len(threads.Threads) != 2 {
		t.Fatalf("threads = %+v", threads)
	}
	firstThread := taskThreadByAssignment(t, threads.Threads, first.AssignmentID)
	secondThread := taskThreadByAssignment(t, threads.Threads, second.AssignmentID)
	if taskThreadHasActiveStream(firstThread) || firstThread.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("legacy previous thread should remain visible as stopped: %+v", firstThread)
	}
	if !taskThreadHasActiveStream(secondThread) {
		t.Fatalf("legacy replacement thread should remain active or queued: %+v", secondThread)
	}
	switch secondThread.AssignmentState {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning:
	default:
		t.Fatalf("legacy replacement state = %+v", secondThread)
	}
}

func legacyAssignSameAgent(t *testing.T, server http.Handler, base, token, taskID, agentID string) AIAgentTaskActionResponse {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/assignment", token, `{"agent_id":"`+agentID+`"}`, http.StatusAccepted)
	var out AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &out)
	return out
}
