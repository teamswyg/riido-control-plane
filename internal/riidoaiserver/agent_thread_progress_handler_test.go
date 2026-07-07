package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAgentThreadProgressRecordsEventsAndFallback(t *testing.T) {
	var recorded []AgentEventRequest
	store := &handlerAssignmentStore{
		record: func(_ context.Context, _ string, req AgentEventRequest) (AgentEventResponse, error) {
			recorded = append(recorded, req)
			return AgentEventResponse{SchemaVersion: SchemaVersion}, nil
		},
	}
	server := Server{
		assignment: store,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:events:write"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(progressBody(`"run_id":"run-a",`)))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentThreadProgress(resp, req, "agent-a")
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body %s", resp.Code, resp.Body.String())
	}
	if len(recorded) != 2 || recorded[0].EventType != EventRiidoLog || recorded[0].Message != "thinking" {
		t.Fatalf("recorded events = %#v", recorded)
	}
	var out AgentThreadProgressBatchResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.AcceptedLines != 2 || out.Event.RunID != "run-a" || out.Event.EventType != AgentClientEventThreadProgress {
		t.Fatalf("response = %#v", out)
	}
}

func TestHandleAgentThreadProgressRecorderErrorFailsClosed(t *testing.T) {
	store := &handlerAssignmentStore{}
	recorder := &handlerRecorderStore{progressErr: ErrAIAgentNotFound}
	server := Server{
		assignment: store,
		aiAgent:    recorder,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:events:write"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(progressBody("")))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentThreadProgress(resp, req, "agent-a")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body %s", resp.Code, resp.Body.String())
	}
	if recorder.progressCalls != 1 || recorder.progressReq.RunID != "run-asn-a" {
		t.Fatalf("recorder calls = %d req = %#v", recorder.progressCalls, recorder.progressReq)
	}
}

func TestHandleAgentThreadProgressRejectsEmptyLines(t *testing.T) {
	server := Server{
		assignment: &handlerAssignmentStore{},
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:events:write"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"assignment_id":"asn-a","task_id":"task-a","lines":[]}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentThreadProgress(resp, req, "agent-a")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body %s", resp.Code, resp.Body.String())
	}
}
