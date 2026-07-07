package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAgentHeartbeatRecordsReturnedEvents(t *testing.T) {
	event := TaskEvent{Seq: 3, TaskID: "task-a", AssignmentID: "asn-a", Type: EventAssignmentRunning, State: AssignmentRunning}
	store := &handlerHeartbeatEventStore{
		handlerAssignmentStore: &handlerAssignmentStore{},
		heartbeatEvents: func(_ context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []TaskEvent, error) {
			if agentID != "agent-a" || req.DaemonID != "daemon-a" {
				t.Fatalf("heartbeat target agent=%s req=%+v", agentID, req)
			}
			return AgentHeartbeatResponse{SchemaVersion: SchemaVersion}, []TaskEvent{event}, nil
		},
	}
	recorder := &handlerRecorderStore{}
	server := Server{
		assignment: store,
		aiAgent:    recorder,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:heartbeat"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"daemon_id":"daemon-a","device_id":"device-a","runtime_id":"runtime-a"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentHeartbeat(resp, req, "agent-a")
	if resp.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", resp.Code, resp.Body.String())
	}
	if recorder.eventCalls != 1 || recorder.eventReq.DaemonID != "daemon-a" || recorder.event.State != AssignmentRunning {
		t.Fatalf("recorder = calls:%d event:%+v req:%+v", recorder.eventCalls, recorder.event, recorder.eventReq)
	}
}

func TestHandleAgentHeartbeatFallsBackWithoutEventStore(t *testing.T) {
	store := &handlerAssignmentStore{
		heartbeat: func(context.Context, string, AgentHeartbeatRequest) (AgentHeartbeatResponse, error) {
			return AgentHeartbeatResponse{SchemaVersion: SchemaVersion}, nil
		},
	}
	server := Server{
		assignment: store,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:heartbeat"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentHeartbeat(resp, req, "agent-a")
	if resp.Code != http.StatusOK {
		t.Fatalf("heartbeat fallback status=%d body=%s", resp.Code, resp.Body.String())
	}
}
