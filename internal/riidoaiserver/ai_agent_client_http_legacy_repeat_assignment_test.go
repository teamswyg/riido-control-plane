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
	if !taskThreadHasActiveStream(secondThread) ||
		secondThread.WorkStatus != AgentWorkStatusIdle ||
		secondThread.AssignmentState != AgentAssignmentStateQueued ||
		secondThread.CommentKind != "" {
		t.Fatalf("legacy replacement thread should preserve queued lifecycle: %+v", secondThread)
	}
}

func TestHTTPAIAgentClientLegacyAssignmentHidesBlockedRepeatQueuedProjection(t *testing.T) {
	const token = "owner-token"
	const taskID = "task-legacy-repeat-blocked"
	const agentID = "agent-public-openclaw"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	first := legacyAssignSameAgent(t, server, base, token, taskID, agentID)
	assertProviderMultiPollStart(t, server, token, first, "daemon-shared-studio", "device-shared-studio", "runtime-openclaw-shared")
	second := legacyAssignSameAgent(t, server, base, token, taskID, agentID)
	if second.AssignmentID == first.AssignmentID || second.ThreadID == first.ThreadID {
		t.Fatalf("blocked legacy repeat collapsed: first=%+v second=%+v", first, second)
	}
	if second.WorkStatus != AgentWorkStatusIdle ||
		second.AssignmentState != AgentAssignmentStateQueued ||
		second.CommentKind != "" ||
		second.Message != "" ||
		second.ActiveStream == nil {
		t.Fatalf("blocked repeat should hide queued client projection: %+v", second)
	}
}

func legacyAssignSameAgent(t *testing.T, server http.Handler, base, token, taskID, agentID string) AIAgentTaskActionResponse {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/assignment", token, `{"agent_id":"`+agentID+`"}`, http.StatusAccepted)
	var out AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &out)
	return out
}
