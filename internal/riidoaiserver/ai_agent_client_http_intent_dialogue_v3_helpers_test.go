package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func newIntentDialogueHTTPTestServer(t *testing.T) (http.Handler, *DevelopmentAIAgentClientStore, *Store) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "agent:*:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiStore})
	t.Cleanup(func() { assignmentStore.Close() })
	server := NewServer(ServerConfig{
		AIAgentClient: aiStore, Assignment: assignmentStore,
		TaskContext: &assignmentHTTPRequestTaskContextReader{contextSnapshot: marketingTaskContextFixture()},
		Authorizer:  authorizer,
	}).Handler()
	return server, aiStore, assignmentStore
}

func postIntentDialogueAssignment(t *testing.T, handler http.Handler, agentID string) AIAgentTaskActionResponse {
	t.Helper()
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-copy/agent-assignments"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	return decodeAIAgentTaskActionResponse(t, resp.Body.Bytes())
}

func postIntentDialogueThreadMessage(t *testing.T, handler http.Handler, base AIAgentTaskActionResponse, body string) AIAgentTaskActionResponse {
	t.Helper()
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/" + base.TaskID + "/threads/" + base.ThreadID + "/messages"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"body":`+strconv.Quote(body)+`}`))
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("message status=%d body=%s", resp.Code, resp.Body.String())
	}
	return decodeAIAgentTaskActionResponse(t, resp.Body.Bytes())
}

func getIntentDialogueV3History(t *testing.T, handler http.Handler, taskID string) AIAgentTaskThreadHistoryCollectionResponse {
	t.Helper()
	path := "/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks/" + taskID + "/threads"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", resp.Code, resp.Body.String())
	}
	var history AIAgentTaskThreadHistoryCollectionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &history); err != nil {
		t.Fatalf("history json: %v", err)
	}
	return history
}
