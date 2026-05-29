package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	thumbnailURL := "https://cdn.riido.io/mock/ai-agents/updated-claude.png"
	description := strings.Repeat("설", AgentDescriptionMaxCharacters)
	instruction := strings.Repeat("지", AgentInstructionMaxCharacters)
	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:                "신규 코리",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-mock",
		ProfileThumbnailURL: &thumbnailURL,
		Description:         &description,
		Instruction:         &instruction,
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(createBody)))
	createReq.Header.Set("Authorization", "Bearer owner-token")
	createResp := httptest.NewRecorder()
	server.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("create json: %v", err)
	}
	if created.Agent.OwnerPrincipalID != "user-1" ||
		!created.Agent.IsOwnedByViewer ||
		created.Agent.Name != "신규 코리" ||
		created.Agent.RuntimeKind != RuntimeKindCursor ||
		created.Agent.WorkStatus != AgentWorkStatusIdle ||
		created.Agent.Editability != AgentEditabilityEditable ||
		created.Agent.AssignedTaskCount != 0 ||
		created.Agent.ProfileThumbnailURL != thumbnailURL ||
		created.Agent.Description != description ||
		created.Agent.Instruction != instruction ||
		created.Agent.UpdatedAt.IsZero() {
		t.Fatalf("created agent = %+v", created.Agent)
	}

	patchBody, err := json.Marshal(UpdateAgentConfigurationRequest{
		Name:                "같은 이름 가능",
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-cursor-mock",
		ProfileThumbnailURL: &thumbnailURL,
		Description:         &description,
		Instruction:         &instruction,
	})
	if err != nil {
		t.Fatalf("marshal patch body: %v", err)
	}
	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(string(patchBody)))
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
	if patched.Agent.Name != "같은 이름 가능" ||
		patched.Agent.Visibility != AgentVisibilityPublic ||
		patched.Agent.RuntimeKind != RuntimeKindCursor ||
		patched.Agent.ProfileThumbnailURL != thumbnailURL ||
		patched.Agent.Description != description ||
		patched.Agent.Instruction != instruction {
		t.Fatalf("patched agent = %+v", patched.Agent)
	}
	if patched.Agent.UpdatedAt.IsZero() {
		t.Fatalf("patched agent updated_at is zero: %+v", patched.Agent)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer owner-token")
	bootstrapResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	updated, ok := findAIAgent(bootstrap.Agents, "agent-owned-claude")
	if !ok || updated.ProfileThumbnailURL != thumbnailURL || updated.Description != description || updated.Instruction != instruction || !updated.UpdatedAt.Equal(patched.Agent.UpdatedAt) {
		t.Fatalf("bootstrap updated agent = %+v found=%v", updated, ok)
	}
	createdAgain, ok := findAIAgent(bootstrap.Agents, created.Agent.AgentID)
	if !ok || createdAgain.OwnerPrincipalID != "user-1" || createdAgain.RuntimeID != "runtime-cursor-mock" {
		t.Fatalf("bootstrap created agent = %+v found=%v", createdAgain, ok)
	}
	if !runtimeHasAssignedAgent(bootstrap.Devices, "runtime-cursor-mock") {
		t.Fatalf("bootstrap runtime-cursor-mock was not marked assigned: %+v", bootstrap.Devices)
	}

	tooLongDescription := strings.Repeat("가", AgentDescriptionMaxCharacters+1)
	tooLongDescriptionBody, err := json.Marshal(UpdateAgentConfigurationRequest{Description: &tooLongDescription})
	if err != nil {
		t.Fatalf("marshal too-long description patch body: %v", err)
	}
	tooLongDescriptionReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(string(tooLongDescriptionBody)))
	tooLongDescriptionReq.Header.Set("Authorization", "Bearer owner-token")
	tooLongDescriptionResp := httptest.NewRecorder()
	server.ServeHTTP(tooLongDescriptionResp, tooLongDescriptionReq)
	if tooLongDescriptionResp.Code != http.StatusBadRequest {
		t.Fatalf("too-long description patch status=%d body=%s", tooLongDescriptionResp.Code, tooLongDescriptionResp.Body.String())
	}

	tooLongInstruction := strings.Repeat("가", AgentInstructionMaxCharacters+1)
	tooLongBody, err := json.Marshal(UpdateAgentConfigurationRequest{Instruction: &tooLongInstruction})
	if err != nil {
		t.Fatalf("marshal too-long patch body: %v", err)
	}
	tooLongReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(string(tooLongBody)))
	tooLongReq.Header.Set("Authorization", "Bearer owner-token")
	tooLongResp := httptest.NewRecorder()
	server.ServeHTTP(tooLongResp, tooLongReq)
	if tooLongResp.Code != http.StatusBadRequest {
		t.Fatalf("too-long patch status=%d body=%s", tooLongResp.Code, tooLongResp.Body.String())
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

func TestHTTPAIAgentThreadProgressBatchIngestsAssignmentAndClientEvent(t *testing.T) {
	ctx := context.Background()
	assignmentStore := NewStore()
	defer assignmentStore.Close()
	assignment, err := assignmentStore.AssignTask(ctx, "task-1", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-owned-codex",
		RuntimeProvider: "codex",
		Prompt:          "summarize team projects",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := assignmentStore.PollAgent(ctx, "agent-owned-codex", PollRequest{DaemonID: "daemon-1", RuntimeID: "runtime-1"}); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon-1",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}, {
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:stream"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: NewMockAIAgentClientStore(),
		Authorizer:    authorizer,
	}).Handler()

	body := `{"assignment_id":"` + assignment.ID + `","task_id":"task-1","daemon_id":"daemon-1","runtime_id":"runtime-1","run_id":"run-1","lines":[{"seq":1,"message":"생각 중..."},{"seq":2,"message":"팀 프로젝트 수집 중 - 팀의 프로젝트 목록을 조회 중."}]}`
	ingestReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/thread-progress", strings.NewReader(body))
	ingestReq.Header.Set("Authorization", "Bearer daemon-token")
	ingestResp := httptest.NewRecorder()
	handler.ServeHTTP(ingestResp, ingestReq)
	if ingestResp.Code != http.StatusAccepted {
		t.Fatalf("ingest status=%d body=%s", ingestResp.Code, ingestResp.Body.String())
	}
	var response AgentThreadProgressBatchResponse
	if err := json.Unmarshal(ingestResp.Body.Bytes(), &response); err != nil {
		t.Fatalf("ingest json: %v", err)
	}
	if response.AcceptedLines != 2 || response.Event.EventType != AgentClientEventThreadProgress || len(response.Event.Lines) != 2 {
		t.Fatalf("ingest response = %+v", response)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	handler.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	message := eventsResp.Body.String()
	if !strings.Contains(message, "event: agent_thread_progress\n") || !strings.Contains(message, "팀 프로젝트 수집 중") {
		t.Fatalf("events body = %q", message)
	}
}

func TestMockAIAgentClientStoreThreadProgressFanout(t *testing.T) {
	store := NewMockAIAgentClientStore()
	history, events, cancel, err := store.SubscribeAIAgentClientEvents(context.Background(), AuthorizationResult{PrincipalID: "user-1"})
	if err != nil {
		t.Fatalf("SubscribeAIAgentClientEvents: %v", err)
	}
	defer cancel()
	if len(history) == 0 {
		t.Fatal("expected replay history")
	}
	if _, err := store.RecordAIAgentThreadProgress(context.Background(), "agent-owned-codex", AgentThreadProgressBatchRequest{
		AssignmentID: "asn-1",
		TaskID:       "task-1",
		RunID:        "run-1",
		Lines:        []AgentThreadProgressLine{{Seq: 1, Message: "웹 검색 실행 중"}},
	}); err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	select {
	case event := <-events:
		progress, ok := event.Payload.(AgentThreadProgressEvent)
		if !ok || progress.EventType != AgentClientEventThreadProgress || progress.Lines[0].Message != "웹 검색 실행 중" {
			t.Fatalf("fanout event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for progress fanout")
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

func findAIAgent(agents []AgentClientRecord, id string) (AgentClientRecord, bool) {
	for _, agent := range agents {
		if agent.AgentID == id {
			return agent, true
		}
	}
	return AgentClientRecord{}, false
}

func runtimeHasAssignedAgent(devices []DeviceRecord, runtimeID string) bool {
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return runtime.HasAssignedAgent
			}
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
