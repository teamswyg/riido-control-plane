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
	if len(bootstrap.AgentTemplates) != 4 {
		t.Fatalf("bootstrap templates = %+v", bootstrap.AgentTemplates)
	}
	wantTemplates := []struct {
		templateID string
		name       string
		roleLabel  string
	}{
		{templateID: "riido_pm", name: "리도", roleLabel: "PM Agent"},
		{templateID: "yeongsil_backend", name: "영실", roleLabel: "Backend Agent"},
		{templateID: "hongdo_frontend", name: "홍도", roleLabel: "Frontend Agent"},
		{templateID: "jiwon_research", name: "지원", roleLabel: "Research Agent"},
	}
	for i, want := range wantTemplates {
		got := bootstrap.AgentTemplates[i]
		if got.TemplateID != want.templateID ||
			got.Name != want.name ||
			got.RoleLabel != want.roleLabel ||
			got.Description == "" ||
			got.Instruction == "" ||
			got.ProfileThumbnailURL == "" {
			t.Fatalf("bootstrap template[%d] = %+v, want %+v with copy fields", i, got, want)
		}
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

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-1/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.TaskID != "task-1" || len(threads.Threads) != 2 {
		t.Fatalf("threads response = %+v", threads)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != "thread-task-1-codex-2" || threads.ActiveStream.Href != "/v1/client/ai-agent/events" {
		t.Fatalf("active stream = %+v", threads.ActiveStream)
	}
}

func TestHTTPAIAgentClientMockAcceptsExplicitAIAgentTokenHeader(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:read"},
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	req.Header.Set("Authorization", "Bearer wrong-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPAIAgentClientMockAssignableAgentsUseStableIDTieBreak(t *testing.T) {
	store := NewMockAIAgentClientStore()
	ownedCodex := store.agents["agent-owned-codex"]
	ownedClaude := store.agents["agent-owned-claude"]
	ownedCodex.Name = "중복 이름"
	ownedClaude.Name = "중복 이름"
	store.agents[ownedCodex.AgentID] = ownedCodex
	store.agents[ownedClaude.AgentID] = ownedClaude

	publicA := store.agents["agent-public-openclaw"]
	publicA.AgentID = "agent-public-alpha"
	publicA.Name = "공개 중복"
	publicZ := publicA
	publicZ.AgentID = "agent-public-zeta"
	store.agents[publicA.AgentID] = publicA
	store.agents[publicZ.AgentID] = publicZ

	handler := NewServer(ServerConfig{
		AIAgentClient: store,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*", "task:task-1:read"}, "user-1"),
	}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-1/assignable-agents", nil)
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assignable status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AgentClientListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("assignable json: %v", err)
	}
	if got, want := aiAgentIDsWithName(out.Agents, "중복 이름"), []string{"agent-owned-claude", "agent-owned-codex"}; !sameStrings(got, want) {
		t.Fatalf("owned duplicate-name order = %v, want %v; all agents=%+v", got, want, out.Agents)
	}
	if got, want := aiAgentIDsWithName(out.Agents, "공개 중복"), []string{"agent-public-alpha", "agent-public-zeta"}; !sameStrings(got, want) {
		t.Fatalf("public duplicate-name order = %v, want %v; all agents=%+v", got, want, out.Agents)
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
	if len(devices.Devices) != 2 {
		t.Fatalf("devices = %+v", devices)
	}
	ownedDevice, ok := findDevice(devices.Devices, "device-mock-macbook")
	if !ok || len(ownedDevice.Runtimes) != 3 {
		t.Fatalf("owned device = %+v, ok=%v", ownedDevice, ok)
	}
	sharedDevice, ok := findDevice(devices.Devices, "device-shared-studio")
	if !ok || len(sharedDevice.Runtimes) != 1 || sharedDevice.Runtimes[0].RuntimeID != "runtime-openclaw-shared" {
		t.Fatalf("shared public-agent device = %+v, ok=%v", sharedDevice, ok)
	}
	cursorRuntime := ownedDevice.Runtimes[2]
	if cursorRuntime.RuntimeID != "runtime-cursor-mock" || len(cursorRuntime.Models) != 2 || cursorRuntime.Models[0].ModelID != "cursor-auto" || !cursorRuntime.Models[0].IsDefault {
		t.Fatalf("cursor runtime models = %+v", cursorRuntime)
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

func TestHTTPAIAgentClientMockDoesNotExposeWaitlistMarketingMutation(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	paths := []string{
		"/v1/client/ai-agent/waitlist",
		"/v1/client/ai-agent/runtime-waitlist",
		"/v1/client/ai-agent/marketing-consent",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"viewer@example.com"}`))
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("%s should not be exposed by AI Agent client API, status=%d body=%s", path, resp.Code, resp.Body.String())
		}
	}
}

func TestHTTPAIAgentClientMockDeviceDaemonDetailAndControl(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{
		{
			PrincipalID: "user-1",
			Token:       "user-token",
			Scopes:      []string{"ai-agent:*"},
		},
		{
			PrincipalID: "admin-user",
			Token:       "admin-token",
			Scopes:      []string{"ai-agent:*"},
			Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
		},
	})

	detailReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-owned-codex/daemon", nil)
	detailReq.Header.Set("Authorization", "Bearer user-token")
	detailResp := httptest.NewRecorder()
	server.ServeHTTP(detailResp, detailReq)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("daemon detail status=%d body=%s", detailResp.Code, detailResp.Body.String())
	}
	var detail DeviceDaemonDetailResponse
	if err := json.Unmarshal(detailResp.Body.Bytes(), &detail); err != nil {
		t.Fatalf("daemon detail json: %v", err)
	}
	if detail.Daemon.Availability != DaemonAvailabilityOnline || detail.Daemon.PID != 5111 || detail.Daemon.Profile != "desktop-api.riido.ai" {
		t.Fatalf("daemon detail = %+v", detail.Daemon)
	}
	if !sameDaemonActions(detail.Daemon.SupportedActions, []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop}) {
		t.Fatalf("daemon supported actions = %+v", detail.Daemon.SupportedActions)
	}

	publicDetailReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon", nil)
	publicDetailReq.Header.Set("Authorization", "Bearer user-token")
	publicDetailResp := httptest.NewRecorder()
	server.ServeHTTP(publicDetailResp, publicDetailReq)
	if publicDetailResp.Code != http.StatusOK {
		t.Fatalf("public agent daemon detail status=%d body=%s", publicDetailResp.Code, publicDetailResp.Body.String())
	}
	var publicDetail DeviceDaemonDetailResponse
	if err := json.Unmarshal(publicDetailResp.Body.Bytes(), &publicDetail); err != nil {
		t.Fatalf("public daemon detail json: %v", err)
	}
	if publicDetail.Daemon.DeviceID != "device-shared-studio" || publicDetail.Daemon.OwnerPrincipalID != "user-2" {
		t.Fatalf("public agent daemon detail = %+v", publicDetail.Daemon)
	}

	privateDeniedReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-private-cursor/daemon", nil)
	privateDeniedReq.Header.Set("Authorization", "Bearer user-token")
	privateDeniedResp := httptest.NewRecorder()
	server.ServeHTTP(privateDeniedResp, privateDeniedReq)
	if privateDeniedResp.Code != http.StatusNotFound {
		t.Fatalf("private agent daemon detail for non-admin status=%d body=%s", privateDeniedResp.Code, privateDeniedResp.Body.String())
	}

	adminPrivateReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-private-cursor/daemon", nil)
	adminPrivateReq.Header.Set("Authorization", "Bearer admin-token")
	adminPrivateResp := httptest.NewRecorder()
	server.ServeHTTP(adminPrivateResp, adminPrivateReq)
	if adminPrivateResp.Code != http.StatusOK {
		t.Fatalf("private agent daemon detail for admin status=%d body=%s", adminPrivateResp.Code, adminPrivateResp.Body.String())
	}

	restartReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/restart", strings.NewReader(`{"reason":"settings page restart"}`))
	restartReq.Header.Set("Authorization", "Bearer user-token")
	restartResp := httptest.NewRecorder()
	server.ServeHTTP(restartResp, restartReq)
	if restartResp.Code != http.StatusAccepted {
		t.Fatalf("daemon restart status=%d body=%s", restartResp.Code, restartResp.Body.String())
	}
	var restart DeviceDaemonCommandResponse
	if err := json.Unmarshal(restartResp.Body.Bytes(), &restart); err != nil {
		t.Fatalf("daemon restart json: %v", err)
	}
	if restart.Action != DaemonControlActionRestart || restart.ControlState != DaemonControlStateRestarting {
		t.Fatalf("daemon restart = %+v", restart)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/stop", strings.NewReader(`{"reason":"settings page stop"}`))
	stopReq.Header.Set("Authorization", "Bearer user-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code != http.StatusAccepted {
		t.Fatalf("daemon stop status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	var stop DeviceDaemonCommandResponse
	if err := json.Unmarshal(stopResp.Body.Bytes(), &stop); err != nil {
		t.Fatalf("daemon stop json: %v", err)
	}
	if stop.Action != DaemonControlActionStop || stop.Availability != DaemonAvailabilityOffline || stop.ControlState != DaemonControlStateStopping {
		t.Fatalf("daemon stop = %+v", stop)
	}

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer user-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices after daemon stop status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	var sharedDevice DeviceRecord
	for _, device := range devices.Devices {
		if device.DeviceID == "device-shared-studio" {
			sharedDevice = device
			break
		}
	}
	if sharedDevice.DeviceID == "" || len(sharedDevice.Runtimes) != 1 {
		t.Fatalf("shared public-agent device should be visible with one public runtime: %+v", devices.Devices)
	}
	for _, runtime := range sharedDevice.Runtimes {
		if runtime.RuntimeID != "runtime-openclaw-shared" || runtime.Availability != RuntimeAvailabilityOffline {
			t.Fatalf("public runtime should be offline after daemon stop: %+v", runtime)
		}
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK || !strings.Contains(eventsResp.Body.String(), AgentClientEventDeviceDaemonStatus) {
		t.Fatalf("events should include daemon status, status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}

	privateStopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents/agent-private-cursor/daemon/stop", nil)
	privateStopReq.Header.Set("Authorization", "Bearer user-token")
	privateStopResp := httptest.NewRecorder()
	server.ServeHTTP(privateStopResp, privateStopReq)
	if privateStopResp.Code != http.StatusNotFound {
		t.Fatalf("private agent daemon control for non-admin status=%d body=%s", privateStopResp.Code, privateStopResp.Body.String())
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
	if stopped.ThreadID == "" {
		t.Fatalf("stop response missing thread_id: %+v", stopped)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-1/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream != nil {
		t.Fatalf("stopped thread collection should not advertise active stream: %+v", threads.ActiveStream)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if body := eventsResp.Body.String(); !strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) || !strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}

func TestHTTPAIAgentClientMockTaskAssignmentAndParticipantRemoval(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign", "task:task-new:comment", "task:task-new:stop"},
	}})

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assigned.TaskID != "task-new" ||
		assigned.AgentID != "agent-public-openclaw" ||
		assigned.AssignmentState != AgentAssignmentStateRunning ||
		assigned.CommentKind != AgentTaskCommentAssignmentStarted ||
		assigned.ThreadID == "" {
		t.Fatalf("assign response = %+v", assigned)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 || threads.ActiveStream == nil || threads.ActiveStream.ThreadID != assigned.ThreadID {
		t.Fatalf("threads after assign = %+v", threads)
	}

	messageReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/threads/"+assigned.ThreadID+"/messages", strings.NewReader(`{"body":"다음 작업을 이어서 진행해 주세요.","source_message_id":"message-next-1"}`))
	messageReq.Header.Set("Authorization", "Bearer user-token")
	messageResp := httptest.NewRecorder()
	server.ServeHTTP(messageResp, messageReq)
	if messageResp.Code != http.StatusAccepted {
		t.Fatalf("thread message status=%d body=%s", messageResp.Code, messageResp.Body.String())
	}
	var message AIAgentTaskActionResponse
	if err := json.Unmarshal(messageResp.Body.Bytes(), &message); err != nil {
		t.Fatalf("thread message json: %v", err)
	}
	if message.ThreadID != assigned.ThreadID ||
		message.AssignmentState != AgentAssignmentStateRunning ||
		message.CommentKind != AgentTaskCommentRuntimeProgress {
		t.Fatalf("thread message response = %+v", message)
	}

	threadsAfterMessageReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsAfterMessageReq.Header.Set("Authorization", "Bearer user-token")
	threadsAfterMessageResp := httptest.NewRecorder()
	server.ServeHTTP(threadsAfterMessageResp, threadsAfterMessageReq)
	if threadsAfterMessageResp.Code != http.StatusOK {
		t.Fatalf("threads after message status=%d body=%s", threadsAfterMessageResp.Code, threadsAfterMessageResp.Body.String())
	}
	var threadsAfterMessage AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsAfterMessageResp.Body.Bytes(), &threadsAfterMessage); err != nil {
		t.Fatalf("threads after message json: %v", err)
	}
	if len(threadsAfterMessage.Threads) != 1 || threadsAfterMessage.Threads[0].SourceMessageID != "message-next-1" {
		t.Fatalf("threads after message = %+v", threadsAfterMessage)
	}

	unassignReq := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"removed from participants"}`))
	unassignReq.Header.Set("Authorization", "Bearer user-token")
	unassignResp := httptest.NewRecorder()
	server.ServeHTTP(unassignResp, unassignReq)
	if unassignResp.Code != http.StatusAccepted {
		t.Fatalf("unassign status=%d body=%s", unassignResp.Code, unassignResp.Body.String())
	}
	var unassigned AIAgentTaskActionResponse
	if err := json.Unmarshal(unassignResp.Body.Bytes(), &unassigned); err != nil {
		t.Fatalf("unassign json: %v", err)
	}
	if unassigned.ThreadID != assigned.ThreadID ||
		unassigned.AssignmentState != AgentAssignmentStateStopped ||
		unassigned.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("unassign response = %+v", unassigned)
	}

	threadsAfterReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsAfterReq.Header.Set("Authorization", "Bearer user-token")
	threadsAfterResp := httptest.NewRecorder()
	server.ServeHTTP(threadsAfterResp, threadsAfterReq)
	if threadsAfterResp.Code != http.StatusOK {
		t.Fatalf("threads after stop status=%d body=%s", threadsAfterResp.Code, threadsAfterResp.Body.String())
	}
	var threadsAfter AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsAfterResp.Body.Bytes(), &threadsAfter); err != nil {
		t.Fatalf("threads after stop json: %v", err)
	}
	if threadsAfter.ActiveStream != nil || len(threadsAfter.Threads) != 1 || threadsAfter.Threads[0].CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("threads after unassign = %+v", threadsAfter)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if body := eventsResp.Body.String(); !strings.Contains(body, string(AgentTaskCommentAssignmentStarted)) || !strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}

func TestHTTPAIAgentClientMockTaskThreadColdCollectionAfterViewerAwayAssignment(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-viewer-away:read", "task:task-viewer-away:comment"},
	}})

	commentReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-viewer-away/comments", strings.NewReader(`{"agent_id":"agent-public-openclaw","body":"Start while the viewer is looking elsewhere","source_comment_id":"comment-viewer-away"}`))
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
	if comment.ThreadID == "" || comment.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("comment response = %+v", comment)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-viewer-away/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].ThreadID != comment.ThreadID {
		t.Fatalf("threads did not return persisted viewer-away thread: %+v", threads)
	}
	if threads.Threads[0].SourceCommentID != "comment-viewer-away" {
		t.Fatalf("source comment id = %q", threads.Threads[0].SourceCommentID)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != comment.ThreadID || threads.ActiveStream.TaskID != "task-viewer-away" {
		t.Fatalf("active stream = %+v", threads.ActiveStream)
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
		created.Agent.ModelID != "cursor-auto" ||
		created.Agent.ModelLabel != "Cursor Auto" ||
		created.Agent.WorkStatus != AgentWorkStatusIdle ||
		created.Agent.Editability != AgentEditabilityEditable ||
		created.Agent.AssignedTaskCount != 0 ||
		created.Agent.ProfileThumbnailURL != thumbnailURL ||
		created.Agent.Description != description ||
		created.Agent.Instruction != instruction ||
		created.Agent.CreatedAt.IsZero() ||
		created.Agent.UpdatedAt.IsZero() {
		t.Fatalf("created agent = %+v", created.Agent)
	}

	patchBody, err := json.Marshal(UpdateAgentConfigurationRequest{
		Name:                "같은 이름 가능",
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-cursor-mock",
		ModelID:             stringPtr("cursor-fast"),
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
		patched.Agent.ModelID != "cursor-fast" ||
		patched.Agent.ModelLabel != "Cursor Fast" ||
		patched.Agent.ProfileThumbnailURL != thumbnailURL ||
		patched.Agent.Description != description ||
		patched.Agent.Instruction != instruction {
		t.Fatalf("patched agent = %+v", patched.Agent)
	}
	if patched.Agent.UpdatedAt.IsZero() {
		t.Fatalf("patched agent updated_at is zero: %+v", patched.Agent)
	}
	if patched.Agent.CreatedAt.IsZero() || !patched.Agent.CreatedAt.Before(patched.Agent.UpdatedAt) {
		t.Fatalf("patched agent created_at must be preserved and before updated_at: %+v", patched.Agent)
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
	if !ok || updated.ProfileThumbnailURL != thumbnailURL || updated.Description != description || updated.Instruction != instruction || !updated.CreatedAt.Equal(patched.Agent.CreatedAt) || !updated.UpdatedAt.Equal(patched.Agent.UpdatedAt) {
		t.Fatalf("bootstrap updated agent = %+v found=%v", updated, ok)
	}
	createdAgain, ok := findAIAgent(bootstrap.Agents, created.Agent.AgentID)
	if !ok || createdAgain.OwnerPrincipalID != "user-1" || createdAgain.RuntimeID != "runtime-cursor-mock" || !createdAgain.CreatedAt.Equal(created.Agent.CreatedAt) || !createdAgain.UpdatedAt.Equal(created.Agent.UpdatedAt) {
		t.Fatalf("bootstrap created agent = %+v found=%v", createdAgain, ok)
	}
	if !runtimeHasAssignedAgent(bootstrap.Devices, "runtime-cursor-mock") {
		t.Fatalf("bootstrap runtime-cursor-mock was not marked assigned: %+v", bootstrap.Devices)
	}

	invalidModelBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "잘못된 모델",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-mock",
		ModelID:    stringPtr("claude-opus-4-7"),
	})
	if err != nil {
		t.Fatalf("marshal invalid model body: %v", err)
	}
	invalidModelReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(invalidModelBody)))
	invalidModelReq.Header.Set("Authorization", "Bearer owner-token")
	invalidModelResp := httptest.NewRecorder()
	server.ServeHTTP(invalidModelResp, invalidModelReq)
	if invalidModelResp.Code != http.StatusBadRequest {
		t.Fatalf("invalid model create status=%d body=%s", invalidModelResp.Code, invalidModelResp.Body.String())
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

func TestHTTPAIAgentClientMockAdminCreateUsesAuthorizedWorkspaceRuntime(t *testing.T) {
	adminServer := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "admin-1",
		Token:       "admin-token",
		Scopes:      []string{"ai-agent:*"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}})

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set("Authorization", "Bearer admin-token")
	devicesResp := httptest.NewRecorder()
	adminServer.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("admin devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("admin devices json: %v", err)
	}
	if len(devices.Devices) != 2 {
		t.Fatalf("admin devices = %+v", devices.Devices)
	}
	ownedDevice, ok := findDevice(devices.Devices, "device-mock-macbook")
	if !ok || ownedDevice.OwnerPrincipalID != "user-1" || len(ownedDevice.Runtimes) != 3 {
		t.Fatalf("admin owned device = %+v, ok=%v", ownedDevice, ok)
	}
	sharedDevice, ok := findDevice(devices.Devices, "device-shared-studio")
	if !ok || sharedDevice.OwnerPrincipalID != "user-2" || len(sharedDevice.Runtimes) != 2 {
		t.Fatalf("admin shared device = %+v, ok=%v", sharedDevice, ok)
	}

	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "관리자 생성 에이전트",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-codex-mock",
	})
	if err != nil {
		t.Fatalf("marshal admin create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(createBody)))
	createReq.Header.Set("Authorization", "Bearer admin-token")
	createResp := httptest.NewRecorder()
	adminServer.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("admin create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("admin create json: %v", err)
	}
	if created.Agent.OwnerPrincipalID != "admin-1" ||
		created.Agent.RuntimeID != "runtime-codex-mock" ||
		created.Agent.RuntimeKind != RuntimeKindCodex ||
		!created.Agent.IsOwnedByViewer {
		t.Fatalf("admin created agent = %+v", created.Agent)
	}

	nonAdminServer := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-2",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	deniedReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(createBody)))
	deniedReq.Header.Set("Authorization", "Bearer user-token")
	deniedResp := httptest.NewRecorder()
	nonAdminServer.ServeHTTP(deniedResp, deniedReq)
	if deniedResp.Code != http.StatusBadRequest {
		t.Fatalf("non-admin create status=%d body=%s", deniedResp.Code, deniedResp.Body.String())
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

	body := `{"assignment_id":"` + assignment.ID + `","task_id":"task-1","thread_id":"thread-task-1-codex-live","daemon_id":"daemon-1","runtime_id":"runtime-1","run_id":"run-1","lines":[{"seq":1,"message":"생각 중..."},{"seq":2,"message":"팀 프로젝트 수집 중 - 팀의 프로젝트 목록을 조회 중."}]}`
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
	if response.AcceptedLines != 2 || response.Event.EventType != AgentClientEventThreadProgress || response.Event.ThreadID != "thread-task-1-codex-live" || len(response.Event.Lines) != 2 {
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

func aiAgentIDsWithName(agents []AgentClientRecord, name string) []string {
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		if agent.Name == name {
			ids = append(ids, agent.AgentID)
		}
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

func findDevice(devices []DeviceRecord, id string) (DeviceRecord, bool) {
	for _, device := range devices {
		if device.DeviceID == id {
			return device, true
		}
	}
	return DeviceRecord{}, false
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

func stringPtr(value string) *string {
	return &value
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameDaemonActions(got, want []DaemonControlAction) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
