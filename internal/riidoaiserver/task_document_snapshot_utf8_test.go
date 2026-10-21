package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func invalidUTF8SnapshotPayload() []byte {
	raw := append([]byte(`{"agent_id":"agent-owned-codex","request_id":"invalid-body","document_snapshot":{"content":"`), 0xff)
	return append(raw, []byte(`","content_format":"html"}}`)...)
}

func TestTaskDocumentSnapshotRejectsInvalidRawUTF8BeforeJSONReplacement(t *testing.T) {
	var request AssignAIAgentTaskRequest
	err := json.Unmarshal(invalidUTF8SnapshotPayload(), &request)
	if err == nil || !strings.Contains(err.Error(), "document_snapshot") {
		t.Fatalf("invalid UTF-8 was replaced instead of rejected: %+v error=%v", request, err)
	}
}

func TestHTTPTaskDocumentSnapshotRejectsInvalidUTF8WithoutExecution(t *testing.T) {
	fixture := newNativeResumeHTTPFixture(t)
	req := httptest.NewRequest(http.MethodPost, fixture.base+"/tasks/task-new/agent-assignments", bytes.NewReader(invalidUTF8SnapshotPayload()))
	req.Header.Set(aiAgentTokenHeader, fixture.token)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "document_snapshot") {
		t.Fatalf("invalid raw body HTTP=%d %s", response.Code, response.Body.String())
	}
	metrics, err := fixture.store.Metrics(t.Context())
	if err != nil || metrics.AssignmentsTotal != 0 {
		t.Fatalf("rejected input created execution: %+v error=%v", metrics, err)
	}
}
