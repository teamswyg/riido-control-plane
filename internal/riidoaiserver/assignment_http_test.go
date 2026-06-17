package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
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

	eventReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/events", bytes.NewReader([]byte(`{"assignment_id":"`+pollOut.Assignment.ID+`","daemon_id":"daemon-a","runtime_id":"runtime-a","state":"failed","event_type":"assignment_failed","message":"approval_timeout: no headless approval path","metadata":{"`+metadatakeys.AssignmentResultStatus.String()+`":"blocked","`+metadatakeys.AssignmentFailureCategory.String()+`":"provider_blocked"}}`)))
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
	if eventOut.Assignment == nil || eventOut.Assignment.State != AssignmentFailed || eventOut.Event.Type != EventAssignmentFailed {
		t.Fatalf("event response = %+v", eventOut)
	}
	if got := eventOut.Event.Metadata[metadatakeys.AssignmentResultStatus.String()]; got != "blocked" {
		t.Fatalf("event result status metadata = %q", got)
	}
	if got := eventOut.Event.Metadata[metadatakeys.AssignmentFailureCategory.String()]; got != "provider_blocked" {
		t.Fatalf("event failure category metadata = %q", got)
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

func TestHTTPAssignmentComposesPromptFromTaskContext(t *testing.T) {
	store := NewStore()
	defer store.Close()
	taskContext := &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}
	server := NewServer(ServerConfig{
		Assignment:  store,
		TaskContext: taskContext,
		Authorizer:  assignmentHTTPAuthorizer(t, []string{"component-task:task-a:assign"}),
	}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out struct {
		SchemaVersion string     `json:"schema_version"`
		Assignment    Assignment `json:"assignment"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if got, want := taskContext.componentIDs, []string{"component-a"}; !sameStrings(got, want) {
		t.Fatalf("task context component ids = %v, want %v", got, want)
	}
	if out.Assignment.ComponentID != "component-a" {
		t.Fatalf("component id = %q", out.Assignment.ComponentID)
	}
	for _, want := range []string{
		"# Riido AI Agent Assignment",
		"branch_name: RIID-4800-server-task-context-http-client-assignment-prompt-wiring",
		"full_name: teamswyg/riido-control-plane",
	} {
		if !strings.Contains(out.Assignment.Prompt, want) {
			t.Fatalf("assignment prompt missing %q:\n%s", want, out.Assignment.Prompt)
		}
	}
}

func TestHTTPAssignmentTaskContextFailureFailsClosed(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{
		Assignment:  store,
		TaskContext: &assignmentHTTPTaskContextReader{err: errors.New("task context unavailable")},
		Authorizer:  assignmentHTTPAuthorizer(t, []string{"component-task:task-a:assign"}),
	}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	if strings.Contains(resp.Body.String(), "prompt is required") {
		t.Fatalf("should fail on task context before assignment store: %s", resp.Body.String())
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

type assignmentHTTPTaskContextReader struct {
	contextSnapshot AIAgentTaskContext
	err             error
	componentIDs    []string
}

func (r *assignmentHTTPTaskContextReader) GetAIAgentTaskContext(ctx context.Context, componentID string) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	r.componentIDs = append(r.componentIDs, componentID)
	if r.err != nil {
		return AIAgentTaskContext{}, r.err
	}
	return r.contextSnapshot, nil
}

type assignmentHTTPRequestTaskContextReader struct {
	contextSnapshot AIAgentTaskContext
	err             error
	requests        []AIAgentTaskContextRequest
}

func (r *assignmentHTTPRequestTaskContextReader) GetAIAgentTaskContext(ctx context.Context, componentID string) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	return r.contextSnapshot, r.err
}

func (r *assignmentHTTPRequestTaskContextReader) GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	r.requests = append(r.requests, req)
	if r.err != nil {
		return AIAgentTaskContext{}, r.err
	}
	return r.contextSnapshot, nil
}
