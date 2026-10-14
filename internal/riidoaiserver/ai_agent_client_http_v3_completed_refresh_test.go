package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientCompletedRefreshIgnoresLateProgress(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1", Token: "owner-token", Scopes: []string{"ai-agent:*", "agent:*:poll"},
	}, {
		PrincipalID: "daemon-dev-macbook", Token: "daemon-token", Scopes: []string{"agent:*:events:write"},
	}})
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	taskID := "task-v3-completed-refresh"
	assignedBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/assignment",
		"owner-token", `{"agent_id":"agent-owned-codex"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, assignedBytes, &assigned)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/poll",
		"owner-token", `{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`,
		http.StatusOK)
	runningBody := `{"assignment_id":"` + assigned.AssignmentID + `","task_id":"` + taskID +
		`","daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev",` +
		`"state":"running","event_type":"riido_log","message":"작업 중"}`
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/events",
		"daemon-token", runningBody, http.StatusOK)

	completedBody := `{"assignment_id":"` + assigned.AssignmentID + `","task_id":"` + taskID +
		`","daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev",` +
		`"state":"completed","event_type":"assignment_completed","message":"작업 완료"}`
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/events",
		"daemon-token", completedBody, http.StatusOK)
	lateBody := `{"assignment_id":"` + assigned.AssignmentID + `","task_id":"` + taskID +
		`","thread_id":"` + assigned.ThreadID + `","daemon_id":"daemon-dev-macbook",` +
		`"device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev","run_id":"` + assigned.RunID +
		`","lines":[{"seq":99,"message":"late progress after completion"}]}`
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/thread-progress",
		"daemon-token", lateBody, http.StatusBadRequest)

	threadsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/threads",
		"owner-token", "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, threadsBytes, &threads)
	if threads.ActiveStream != nil || len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateCompleted ||
		strings.Contains(string(threadsBytes), "late progress after completion") {
		t.Fatalf("completed v2 refresh resurrected or leaked late progress: %+v", threads)
	}

	historyBytes := aiAgentSmokeRequest(t, server, http.MethodGet,
		"/v3/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/tasks/"+taskID+"/threads",
		"owner-token", "", http.StatusOK)
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, historyBytes, &history)
	if history.ActiveStream != nil || len(history.Threads) != 1 ||
		history.Threads[0].AssignmentState != AgentAssignmentStateCompleted ||
		strings.Contains(string(historyBytes), "late progress after completion") {
		t.Fatalf("completed v3 refresh resurrected or leaked late progress: %+v", history)
	}

	subBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/thread-stream-subscription",
		"owner-token", "", http.StatusOK)
	var sub AIAgentTaskThreadStreamSubscriptionResponse
	aiAgentSmokeDecode(t, subBytes, &sub)
	events := string(aiAgentSmokeRequest(t, server, http.MethodGet, base+"/events?replay=1",
		"owner-token", "", http.StatusOK))
	if len(sub.ActiveThreadFilters) != 0 || strings.Contains(events, "late progress after completion") {
		t.Fatalf("completed subscription/events leaked active work: filters=%+v events=%q", sub.ActiveThreadFilters, events)
	}
}
