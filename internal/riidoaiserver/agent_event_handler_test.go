package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAgentEventRecordsReadModelEvent(t *testing.T) {
	event := TaskEvent{Seq: 7, TaskID: "task-a", AssignmentID: "asn-a", AgentID: "agent-a", Type: EventAssignmentRunning}
	store := &handlerAssignmentStore{
		record: func(_ context.Context, agentID string, req AgentEventRequest) (AgentEventResponse, error) {
			if agentID != "agent-a" || req.AssignmentID != "asn-a" {
				t.Fatalf("record target agent=%s req=%+v", agentID, req)
			}
			return AgentEventResponse{SchemaVersion: SchemaVersion, Event: event}, nil
		},
	}
	recorder := &handlerRecorderStore{}
	server := Server{
		assignment: store,
		aiAgent:    recorder,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:events:write"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"assignment_id":"asn-a","task_id":"task-a","state":"running"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentEvent(resp, req, "agent-a")
	if resp.Code != http.StatusOK {
		t.Fatalf("event status=%d body=%s", resp.Code, resp.Body.String())
	}
	if recorder.eventCalls != 1 || recorder.event.Seq != 7 || recorder.eventReq.AssignmentID != "asn-a" {
		t.Fatalf("recorder = calls:%d event:%+v req:%+v", recorder.eventCalls, recorder.event, recorder.eventReq)
	}
}

func TestHandleAgentEventStoreErrorFailsClosed(t *testing.T) {
	store := &handlerAssignmentStore{
		record: func(context.Context, string, AgentEventRequest) (AgentEventResponse, error) {
			return AgentEventResponse{}, errors.New("assignment not found")
		},
	}
	server := Server{
		assignment: store,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:events:write"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"assignment_id":"missing"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentEvent(resp, req, "agent-a")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("event error status=%d body=%s", resp.Code, resp.Body.String())
	}
}
