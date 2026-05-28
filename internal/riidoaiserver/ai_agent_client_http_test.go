package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientMockBootstrapAndAssignableAgents(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-1:read"},
	}})

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap?client_kind=desktop_webview", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer user-token")
	bootstrapResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	if bootstrap.SchemaVersion != SchemaVersion || bootstrap.ClientKind != ClientKindDesktopWebview || bootstrap.WorkspaceID == "" {
		t.Fatalf("bootstrap = %+v", bootstrap)
	}
	if got, want := aiAgentIDs(bootstrap.Agents), []string{"agent-owned-claude", "agent-owned-codex", "agent-public-openclaw"}; !sameStrings(got, want) {
		t.Fatalf("bootstrap agents = %v, want %v", got, want)
	}
	if containsString(aiAgentIDs(bootstrap.Agents), "agent-private-cursor") {
		t.Fatalf("bootstrap leaked other private agent: %+v", bootstrap.Agents)
	}

	assignableReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-1/assignable-agents", nil)
	assignableReq.Header.Set("Authorization", "Bearer user-token")
	assignableResp := httptest.NewRecorder()
	server.ServeHTTP(assignableResp, assignableReq)
	if assignableResp.Code != http.StatusOK {
		t.Fatalf("assignable status=%d body=%s", assignableResp.Code, assignableResp.Body.String())
	}
	var assignable AgentClientListResponse
	if err := json.Unmarshal(assignableResp.Body.Bytes(), &assignable); err != nil {
		t.Fatalf("assignable json: %v", err)
	}
	if len(assignable.Agents) != 3 {
		t.Fatalf("assignable agents = %+v", assignable.Agents)
	}
	if !assignable.Agents[0].IsOwnedByViewer || !assignable.Agents[1].IsOwnedByViewer || assignable.Agents[2].IsOwnedByViewer {
		t.Fatalf("assignable owned-first flags = %+v", assignable.Agents)
	}
	if assignable.Agents[0].Name > assignable.Agents[1].Name {
		t.Fatalf("owned agents should be ordered by name: %+v", assignable.Agents[:2])
	}
}

func TestHTTPAIAgentClientMockDevicesAndEditability(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer user-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if len(devices.Devices) != 1 || len(devices.Devices[0].Runtimes) != 3 {
		t.Fatalf("devices = %+v", devices)
	}

	editReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-owned-codex/editability", nil)
	editReq.Header.Set("Authorization", "Bearer user-token")
	editResp := httptest.NewRecorder()
	server.ServeHTTP(editResp, editReq)
	if editResp.Code != http.StatusOK {
		t.Fatalf("editability status=%d body=%s", editResp.Code, editResp.Body.String())
	}
	var edit AgentEditabilityResponse
	if err := json.Unmarshal(editResp.Body.Bytes(), &edit); err != nil {
		t.Fatalf("editability json: %v", err)
	}
	if edit.Editability != AgentEditabilityBlockedAssignedTasks || edit.AssignedTaskCount != 1 || edit.Reason == "" {
		t.Fatalf("editability = %+v", edit)
	}
}

func TestHTTPAIAgentClientMockTaskCommentAndStop(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-1:comment", "task:task-1:stop"},
	}})

	commentReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-1/comments", strings.NewReader(`{"agent_id":"agent-owned-codex","body":"Please continue from the latest design"}`))
	commentReq.Header.Set("Authorization", "Bearer user-token")
	commentResp := httptest.NewRecorder()
	server.ServeHTTP(commentResp, commentReq)
	if commentResp.Code != http.StatusAccepted {
		t.Fatalf("comment status=%d body=%s", commentResp.Code, commentResp.Body.String())
	}
	var comment AIAgentTaskActionResponse
	if err := json.Unmarshal(commentResp.Body.Bytes(), &comment); err != nil {
		t.Fatalf("comment json: %v", err)
	}
	if comment.AssignmentState != AgentAssignmentStateQueued || comment.CommentKind != AgentTaskCommentQueuedByBusyAgent {
		t.Fatalf("comment response = %+v", comment)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-1/stop", strings.NewReader(`{"agent_id":"agent-owned-codex","reason":"user requested stop"}`))
	stopReq.Header.Set("Authorization", "Bearer user-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code != http.StatusAccepted {
		t.Fatalf("stop status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	var stopped AIAgentTaskActionResponse
	if err := json.Unmarshal(stopResp.Body.Bytes(), &stopped); err != nil {
		t.Fatalf("stop json: %v", err)
	}
	if stopped.AssignmentState != AgentAssignmentStateStopped || stopped.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("stop response = %+v", stopped)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if body := eventsResp.Body.String(); !strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) || !strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}

func TestHTTPAIAgentClientMockMutationAndDeletion(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(`{"name":"같은 이름 가능","visibility":"public","runtime_id":"runtime-cursor-mock"}`))
	patchReq.Header.Set("Authorization", "Bearer owner-token")
	patchResp := httptest.NewRecorder()
	server.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patchResp.Code, patchResp.Body.String())
	}
	var patched AgentClientRecordResponse
	if err := json.Unmarshal(patchResp.Body.Bytes(), &patched); err != nil {
		t.Fatalf("patch json: %v", err)
	}
	if patched.Agent.Name != "같은 이름 가능" || patched.Agent.Visibility != AgentVisibilityPublic || patched.Agent.RuntimeKind != RuntimeKindCursor {
		t.Fatalf("patched agent = %+v", patched.Agent)
	}

	assignedPatchReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-codex", strings.NewReader(`{"name":"blocked"}`))
	assignedPatchReq.Header.Set("Authorization", "Bearer owner-token")
	assignedPatchResp := httptest.NewRecorder()
	server.ServeHTTP(assignedPatchResp, assignedPatchReq)
	if assignedPatchResp.Code != http.StatusConflict {
		t.Fatalf("assigned patch status=%d body=%s", assignedPatchResp.Code, assignedPatchResp.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/agents/agent-owned-codex", nil)
	deleteReq.Header.Set("Authorization", "Bearer owner-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	var deleted DeleteAgentResponse
	if err := json.Unmarshal(deleteResp.Body.Bytes(), &deleted); err != nil {
		t.Fatalf("delete json: %v", err)
	}
	if deleted.RunningTasksForceStopped != 1 || deleted.AgentID != "agent-owned-codex" {
		t.Fatalf("delete response = %+v", deleted)
	}
}

func TestHTTPAIAgentClientMockSSEReplaysTypedCommentStatus(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:stream"},
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("Content-Type = %q", got)
	}
	message := resp.Body.String()
	if !strings.Contains(message, "event: agent_work_status_changed\n") || !strings.Contains(message, string(AgentTaskCommentQueuedByBusyAgent)) {
		t.Fatalf("sse message = %q", message)
	}
}

func TestHTTPAIAgentClientMockRequiresConfigurationAndAuthorization(t *testing.T) {
	unconfigured := NewServer(ServerConfig{Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	unconfigured.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured status=%d body=%s", resp.Code, resp.Body.String())
	}

	configured := NewServer(ServerConfig{AIAgentClient: NewMockAIAgentClientStore(), Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:stream"}, "user-1")}).Handler()
	forbiddenReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(`{"name":"nope"}`))
	forbiddenReq.Header.Set("Authorization", "Bearer ai-agent-token")
	forbiddenResp := httptest.NewRecorder()
	configured.ServeHTTP(forbiddenResp, forbiddenReq)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("forbidden patch status=%d body=%s", forbiddenResp.Code, forbiddenResp.Body.String())
	}
}

func newAIAgentClientHTTPTestServer(t *testing.T, credentials []StaticTokenCredential) http.Handler {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer(credentials)
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{AIAgentClient: NewMockAIAgentClientStore(), Authorizer: authorizer}).Handler()
}

func aiAgentClientHTTPAuthorizer(t *testing.T, scopes []string, principalID string) RequestAuthorizer {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: principalID,
		Token:       "ai-agent-token",
		Scopes:      scopes,
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return authorizer
}

func aiAgentIDs(agents []AgentClientRecord) []string {
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		ids = append(ids, agent.AgentID)
	}
	return ids
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
