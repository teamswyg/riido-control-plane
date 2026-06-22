package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientProviderMultiAssignmentReplaysEachThread(t *testing.T) {
	const token = "owner-token"
	taskID := "task-provider-multi-sse"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	cursorID := createProviderMultiCursorAgent(t, server, base, token)
	first := createProviderMultiAssignment(t, server, base, token, taskID, "agent-public-openclaw")
	second := createProviderMultiAssignment(t, server, base, token, taskID, cursorID)
	if first.AssignmentID == second.AssignmentID || first.ThreadID == second.ThreadID {
		t.Fatalf("multi assignment collapsed: first=%+v second=%+v", first, second)
	}
	assertProviderMultiPollStart(t, server, token, first, "daemon-shared-studio", "device-shared-studio", "runtime-openclaw-shared")
	assertProviderMultiPollStart(t, server, token, second, "daemon-dev-macbook", "device-dev-macbook", "runtime-cursor-dev")
	threadsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, threadsBytes, &threads)
	if len(threads.Threads) != 2 {
		t.Fatalf("threads = %+v", threads)
	}
	taskThreadByAssignment(t, threads.Threads, first.AssignmentID)
	taskThreadByAssignment(t, threads.Threads, second.AssignmentID)
	subBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/thread-stream-subscription", token, "", http.StatusOK)
	var sub AIAgentTaskThreadStreamSubscriptionResponse
	aiAgentSmokeDecode(t, subBytes, &sub)
	if sub.Stream.Href != base+"/events" ||
		!hasThreadFilter(sub.ActiveThreadFilters, first.AgentID, first.ThreadID, first.RunID) ||
		!hasThreadFilter(sub.ActiveThreadFilters, second.AgentID, second.ThreadID, second.RunID) {
		t.Fatalf("stream subscription = %+v", sub)
	}
	events := string(aiAgentSmokeRequest(t, server, http.MethodGet, base+"/events?replay=1", token, "", http.StatusOK))
	if !strings.Contains(events, first.AssignmentID) || !strings.Contains(events, second.AssignmentID) {
		t.Fatalf("sse replay missed multi assignments: %s", events)
	}
	messagePath := base + "/tasks/" + taskID + "/threads/" + first.ThreadID + "/messages"
	messageBody := `{"body":"follow up first thread","source_message_id":"provider-multi-message-1"}`
	messageBytes := aiAgentSmokeRequest(t, server, http.MethodPost, messagePath, token, messageBody, http.StatusAccepted)
	var message AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, messageBytes, &message)
	if message.ThreadID != first.ThreadID || message.AgentID != first.AgentID {
		t.Fatalf("thread message targeted wrong thread: %+v first=%+v", message, first)
	}
}

func createProviderMultiCursorAgent(t *testing.T, server http.Handler, base, token string) string {
	t.Helper()
	body := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
		Name: "multi cursor", Visibility: AgentVisibilityPrivate, RuntimeID: "runtime-cursor-dev",
	})
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents", token, body, http.StatusCreated)
	var created AgentClientRecordResponse
	aiAgentSmokeDecode(t, bytes, &created)
	return created.Agent.AgentID
}

func createProviderMultiAssignment(t *testing.T, server http.Handler, base, token, taskID, agentID string) AIAgentTaskActionResponse {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments", token, `{"agent_id":"`+agentID+`"}`, http.StatusAccepted)
	var out AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &out)
	return out
}
