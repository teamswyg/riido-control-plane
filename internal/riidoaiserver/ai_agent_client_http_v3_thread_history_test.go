package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientV3ThreadHistory(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	assignReq := httptest.NewRequest(http.MethodPost, base+"/tasks/task-v3-history/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "owner-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	historyReq := httptest.NewRequest(http.MethodGet, "/v3/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/tasks/task-v3-history/threads", nil)
	historyReq.Header.Set(aiAgentTokenHeader, "owner-token")
	historyResp := httptest.NewRecorder()
	server.ServeHTTP(historyResp, historyReq)
	if historyResp.Code != http.StatusOK {
		t.Fatalf("v3 history status=%d body=%s", historyResp.Code, historyResp.Body.String())
	}
	var history AIAgentTaskThreadHistoryCollectionResponse
	if err := json.Unmarshal(historyResp.Body.Bytes(), &history); err != nil {
		t.Fatalf("v3 history json: %v", err)
	}
	if len(history.Threads) != 1 || len(history.Threads[0].Messages) == 0 {
		t.Fatalf("v3 history = %+v", history)
	}
	if history.Threads[0].AgentSnapshotID == "" || len(history.AgentSnapshots) != 1 {
		t.Fatalf("v3 history snapshots = %+v", history)
	}
}
