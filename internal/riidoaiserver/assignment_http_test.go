package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAssignmentAssignPollHeartbeatAndEvent(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{
		"component-task:task-a:assign",
		"agent:agent-a:poll",
		"agent:agent-a:heartbeat",
		"agent:agent-a:events:write",
	})}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex","prompt":"ship it"}`))
	assignReq.Header.Set("Authorization", "Bearer assignment-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusOK {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assignOut struct {
		SchemaVersion string     `json:"schema_version"`
		Assignment    Assignment `json:"assignment"`
	}
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assignOut); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assignOut.SchemaVersion != SchemaVersion || assignOut.Assignment.ID == "" || assignOut.Assignment.State != AssignmentQueued {
		t.Fatalf("assign response = %+v", assignOut)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", strings.NewReader(`{"daemon_id":"daemon-a","runtime_id":"runtime-a"}`))
	pollReq.Header.Set("Authorization", "Bearer assignment-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var pollOut PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &pollOut); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if pollOut.Action != PollStart || pollOut.Assignment == nil || pollOut.Assignment.ID != assignOut.Assignment.ID {
		t.Fatalf("poll response = %+v", pollOut)
	}

	heartbeatReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/heartbeat", strings.NewReader(`{"daemon_id":"daemon-a","runtime_id":"runtime-a","active_assignment_ids":["`+pollOut.Assignment.ID+`"]}`))
	heartbeatReq.Header.Set("Authorization", "Bearer assignment-token")
	heartbeatResp := httptest.NewRecorder()
	server.ServeHTTP(heartbeatResp, heartbeatReq)
	if heartbeatResp.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", heartbeatResp.Code, heartbeatResp.Body.String())
	}
	var heartbeatOut AgentHeartbeatResponse
	if err := json.Unmarshal(heartbeatResp.Body.Bytes(), &heartbeatOut); err != nil {
		t.Fatalf("heartbeat json: %v", err)
	}
	if len(heartbeatOut.RefreshedAssignments) != 1 || heartbeatOut.RefreshedAssignments[0].ID != pollOut.Assignment.ID {
		t.Fatalf("heartbeat response = %+v", heartbeatOut)
	}

	eventReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/events", bytes.NewReader([]byte(`{"assignment_id":"`+pollOut.Assignment.ID+`","daemon_id":"daemon-a","runtime_id":"runtime-a","state":"ready","event_type":"assignment_ready","message":"ready"}`)))
	eventReq.Header.Set("Authorization", "Bearer assignment-token")
	eventResp := httptest.NewRecorder()
	server.ServeHTTP(eventResp, eventReq)
	if eventResp.Code != http.StatusOK {
		t.Fatalf("event status=%d body=%s", eventResp.Code, eventResp.Body.String())
	}
	var eventOut AgentEventResponse
	if err := json.Unmarshal(eventResp.Body.Bytes(), &eventOut); err != nil {
		t.Fatalf("event json: %v", err)
	}
	if eventOut.Assignment == nil || eventOut.Assignment.State != AssignmentReady || eventOut.Event.Type != EventAssignmentReady {
		t.Fatalf("event response = %+v", eventOut)
	}
}

func TestHTTPAssignmentRequiresScopedAuthorization(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-b:poll"})}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", strings.NewReader(`{"daemon_id":"daemon-a","runtime_id":"runtime-a"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("other agent poll status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPAssignmentRejectsUnknownPrivateFields(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-a:assign"})}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex","prompt":"ship it","provider_executable_path":"/private/bin/codex"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("private field status=%d body=%s", resp.Code, resp.Body.String())
	}
	if strings.Contains(resp.Body.String(), "/private/bin/codex") {
		t.Fatalf("response leaked private path: %s", resp.Body.String())
	}
}

func TestHTTPAssignmentStoreNotConfiguredFailsClosed(t *testing.T) {
	server := NewServer(ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"component-task:task-a:assign"})}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex","prompt":"ship it"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured assignment store status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func assignmentHTTPAuthorizer(t *testing.T, scopes []string) RequestAuthorizer {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "assignment-test",
		Token:       "assignment-token",
		Scopes:      scopes,
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return authorizer
}
