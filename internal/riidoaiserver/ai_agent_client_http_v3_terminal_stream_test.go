package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientV3TerminalHistoryOmitsActiveStream(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	taskID := "task-v3-terminal-stream"
	assignedBytes := aiAgentSmokeRequest(
		t,
		server,
		http.MethodPost,
		base+"/tasks/"+taskID+"/assignment",
		"owner-token",
		`{"agent_id":"agent-owned-codex"}`,
		http.StatusAccepted,
	)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, assignedBytes, &assigned)
	if assigned.ActiveStream == nil {
		t.Fatalf("test setup assignment missing active stream: %+v", assigned)
	}

	aiAgentSmokeRequest(
		t,
		server,
		http.MethodPost,
		base+"/tasks/"+taskID+"/agent-assignments/"+assigned.AgentID+"/stop",
		"owner-token",
		`{"assignment_id":"`+assigned.AssignmentID+`","reason":"user_request"}`,
		http.StatusAccepted,
	)
	historyBytes := aiAgentSmokeRequest(
		t,
		server,
		http.MethodGet,
		"/v3/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/tasks/"+taskID+"/threads",
		"owner-token",
		"",
		http.StatusOK,
	)
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, historyBytes, &history)
	if history.ActiveStream != nil {
		t.Fatalf("terminal-only v3 history must not expose active_stream: %+v", history.ActiveStream)
	}
	if len(history.Threads) != 1 || history.Threads[0].ActiveStream != nil {
		t.Fatalf("terminal thread must not carry per-thread active_stream: %+v", history.Threads)
	}
	if history.Threads[0].AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("terminal history state = %+v", history.Threads[0])
	}
}
