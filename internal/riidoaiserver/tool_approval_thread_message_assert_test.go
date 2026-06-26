package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func assertApprovalChatWaitApproved(t *testing.T, waitDone <-chan ToolApprovalWaitResponse) {
	t.Helper()
	select {
	case waited := <-waitDone:
		if waited.Result.Status != ApprovalApproved || waited.Decision == nil {
			t.Fatalf("wait response = %+v", waited)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for approval decision")
	}
}

func assertApprovalChatHistoryMessage(t *testing.T, server http.Handler, threadID, body string) {
	t.Helper()
	req := httptest.NewRequest(
		http.MethodGet,
		"/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/threads",
		nil,
	)
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AIAgentTaskThreadHistoryCollectionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("history json: %v", err)
	}
	thread := historyThreadByID(t, out, threadID)
	if !historyMessagesContainUserBody(thread.Messages, body) {
		t.Fatalf("history missed approval reply: %+v", thread.Messages)
	}
}

func assertApprovalChatHistoryBody(t *testing.T, server http.Handler, threadID, body string) {
	t.Helper()
	req := httptest.NewRequest(
		http.MethodGet,
		"/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/threads",
		nil,
	)
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AIAgentTaskThreadHistoryCollectionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("history json: %v", err)
	}
	thread := historyThreadByID(t, out, threadID)
	if !historyMessagesContainBody(thread.Messages, body) {
		t.Fatalf("history missed body %q: %+v", body, thread.Messages)
	}
}
