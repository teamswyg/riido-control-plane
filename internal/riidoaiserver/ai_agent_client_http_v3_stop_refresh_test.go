package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientV3StopRefreshIgnoresLateProgress(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}, {
		PrincipalID: "daemon-dev-macbook",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	taskID := "task-v3-stop-refresh"
	assignedBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/assignment",
		"owner-token", `{"agent_id":"agent-owned-codex"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, assignedBytes, &assigned)

	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+assigned.AgentID+"/stop",
		"owner-token", `{"assignment_id":"`+assigned.AssignmentID+`","reason":"user_request"}`, http.StatusAccepted)
	progressBody := `{"assignment_id":"` + assigned.AssignmentID +
		`","task_id":"` + taskID +
		`","thread_id":"` + assigned.ThreadID +
		`","daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev","run_id":"` + assigned.RunID +
		`","lines":[{"seq":99,"message":"late progress after stop"}]}`
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/thread-progress",
		"daemon-token", progressBody, http.StatusBadRequest)

	historyPath := "/v3/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/" + taskID + "/threads"
	historyBytes := aiAgentSmokeRequest(t, server, http.MethodGet, historyPath, "owner-token", "", http.StatusOK)
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, historyBytes, &history)
	if history.ActiveStream != nil || len(history.Threads) != 1 {
		t.Fatalf("stopped refresh history should not expose active stream: %+v", history)
	}
	thread := history.Threads[0]
	if thread.AssignmentState != AgentAssignmentStateStopped || thread.WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("stopped thread resurrected after refresh: %+v", thread)
	}
	if len(thread.Messages) != 1 || thread.Messages[0].Body == "late progress after stop" {
		t.Fatalf("late progress leaked into stopped history: %+v", thread.Messages)
	}
}
