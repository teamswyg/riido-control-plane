package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestHTTPAIAgentClientDevelopmentBootstrapAndAssignableAgents(t *testing.T) {
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

	fixturesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/onboarding/fixtures", nil)
	fixturesReq.Header.Set(aiAgentTokenHeader, "user-token")
	fixturesResp := httptest.NewRecorder()
	server.ServeHTTP(fixturesResp, fixturesReq)
	if fixturesResp.Code != http.StatusOK {
		t.Fatalf("fixtures status=%d body=%s", fixturesResp.Code, fixturesResp.Body.String())
	}
	var fixtures AgentOnboardingFixtureListResponse
	if err := json.Unmarshal(fixturesResp.Body.Bytes(), &fixtures); err != nil {
		t.Fatalf("fixtures json: %v", err)
	}
	if len(fixtures.Fixtures) != 4 {
		t.Fatalf("fixtures = %+v", fixtures.Fixtures)
	}
	wantFixtures := []struct {
		fixtureID              string
		name                   string
		roleLabel              string
		defaultVisibility      AgentVisibility
		recommendedRuntimeKind RuntimeKind
	}{
		{fixtureID: "riido_pm", name: "리도", roleLabel: "PM Agent", defaultVisibility: AgentVisibilityPrivate, recommendedRuntimeKind: RuntimeKindCodex},
		{fixtureID: "yeongsil_backend", name: "영실", roleLabel: "Backend Agent", defaultVisibility: AgentVisibilityPrivate, recommendedRuntimeKind: RuntimeKindClaudeCode},
		{fixtureID: "hongdo_frontend", name: "홍도", roleLabel: "Frontend Agent", defaultVisibility: AgentVisibilityPrivate, recommendedRuntimeKind: RuntimeKindCursor},
		{fixtureID: "jiwon_research", name: "지원", roleLabel: "Research Agent", defaultVisibility: AgentVisibilityPrivate, recommendedRuntimeKind: RuntimeKindOpenClaw},
	}
	for i, want := range wantFixtures {
		got := fixtures.Fixtures[i]
		if got.FixtureID != want.fixtureID ||
			got.Name != want.name ||
			got.RoleLabel != want.roleLabel ||
			got.DefaultVisibility != want.defaultVisibility ||
			got.RecommendedRuntimeKind != want.recommendedRuntimeKind ||
			got.TmpColor == "" ||
			got.Description == "" ||
			got.Instruction == "" ||
			got.ProfileThumbnailURL == "" {
			t.Fatalf("fixture[%d] = %+v, want %+v with copy fields", i, got, want)
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

func TestHTTPAIAgentClientDevelopmentWorkspaceAssignedAgentProfiles(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assigned profiles status=%d body=%s", resp.Code, resp.Body.String())
	}
	var profiles AssignedAgentProfileMapResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &profiles); err != nil {
		t.Fatalf("assigned profiles json: %v", err)
	}
	if profiles.SchemaVersion != SchemaVersion || profiles.WorkspaceID != "workspace-dev-riid" {
		t.Fatalf("assigned profiles response = %+v", profiles)
	}
	active := profiles.AssignedAgentProfiles["task-1"]
	if active.AvatarURL == "" {
		t.Fatalf("task-1 active assigned profile missing avatar: %+v", profiles.AssignedAgentProfiles)
	}
	if _, ok := profiles.AssignedAgentProfiles["task-completed-only"]; ok {
		t.Fatalf("completed-only task leaked into assigned profile map: %+v", profiles.AssignedAgentProfiles)
	}

	fixtureBody := `{"name":"홍도","visibility":"private","runtime_id":"runtime-cursor-dev","model_id":"cursor-auto"}`
	fixtureCreateReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/onboarding/fixtures/hongdo_frontend/agents", strings.NewReader(fixtureBody))
	fixtureCreateReq.Header.Set("Authorization", "Bearer user-token")
	fixtureCreateReq.Header.Set("Content-Type", "application/json")
	fixtureCreateResp := httptest.NewRecorder()
	server.ServeHTTP(fixtureCreateResp, fixtureCreateReq)
	if fixtureCreateResp.Code != http.StatusCreated {
		t.Fatalf("fixture create status=%d body=%s", fixtureCreateResp.Code, fixtureCreateResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(fixtureCreateResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("fixture create json: %v", err)
	}
	if created.Agent.TmpColor != "#B87EAD" {
		t.Fatalf("fixture-created tmp_color = %+v", created.Agent)
	}

	assignReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/23958923859/assignment", strings.NewReader(`{"agent_id":"`+created.Agent.AgentID+`"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignReq.Header.Set("Content-Type", "application/json")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}

	nextReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
	nextReq.Header.Set("Authorization", "Bearer user-token")
	nextResp := httptest.NewRecorder()
	server.ServeHTTP(nextResp, nextReq)
	if nextResp.Code != http.StatusOK {
		t.Fatalf("assigned profiles after assign status=%d body=%s", nextResp.Code, nextResp.Body.String())
	}
	var next AssignedAgentProfileMapResponse
	if err := json.Unmarshal(nextResp.Body.Bytes(), &next); err != nil {
		t.Fatalf("assigned profiles after assign json: %v", err)
	}
	if got := next.AssignedAgentProfiles["23958923859"]; got.TmpColor != "#B87EAD" {
		t.Fatalf("component keyed profile = %+v, all=%+v", got, next.AssignedAgentProfiles)
	}
}

func TestHTTPAIAgentClientDevelopmentAcceptsExplicitAIAgentTokenHeader(t *testing.T) {
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

func TestHTTPDesktopDeviceEnrollmentAndDaemonCredentialAuthorization(t *testing.T) {
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStore()
	defer assignmentStore.Close()
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:create", "ai-agent:device:read"}, "user-1"),
	}).Handler()

	enrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"display_name":"JY MacBook","platform":"darwin","app_version":"0.0.0"}`))
	enrollReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	enrollResp := httptest.NewRecorder()
	server.ServeHTTP(enrollResp, enrollReq)
	if enrollResp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollResp.Code, enrollResp.Body.String())
	}
	var enrollment EnrollDeviceResponse
	if err := json.Unmarshal(enrollResp.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}
	if enrollment.SchemaVersion != DeviceCredentialSchemaVersion ||
		enrollment.DeviceID == "" ||
		enrollment.DeviceSecret == "" ||
		enrollment.OwnerPrincipalID != "user-1" ||
		enrollment.WorkspaceID != "workspace-alpha" ||
		enrollment.DisplayName != "JY MacBook" {
		t.Fatalf("enrollment = %+v", enrollment)
	}
	firstDeviceSecret := enrollment.DeviceSecret

	rotateReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"device_id":"`+enrollment.DeviceID+`","display_name":"JY MacBook","platform":"darwin","app_version":"0.0.1"}`))
	rotateReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	rotateResp := httptest.NewRecorder()
	server.ServeHTTP(rotateResp, rotateReq)
	if rotateResp.Code != http.StatusCreated {
		t.Fatalf("rotate status=%d body=%s", rotateResp.Code, rotateResp.Body.String())
	}
	var rotated EnrollDeviceResponse
	if err := json.Unmarshal(rotateResp.Body.Bytes(), &rotated); err != nil {
		t.Fatalf("rotate json: %v", err)
	}
	if rotated.DeviceID != enrollment.DeviceID ||
		rotated.DeviceSecret == "" ||
		rotated.DeviceSecret == firstDeviceSecret ||
		rotated.OwnerPrincipalID != enrollment.OwnerPrincipalID ||
		rotated.WorkspaceID != enrollment.WorkspaceID {
		t.Fatalf("rotated enrollment = %+v original=%+v", rotated, enrollment)
	}
	if _, err := aiAgentStore.AuthorizeDeviceCredential(context.Background(), enrollment.DeviceID, firstDeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll}); !errors.Is(err, ErrAuthorizationUnauthenticated) {
		t.Fatalf("old rotated device secret must be rejected, got %v", err)
	}
	enrollment = rotated

	devicesReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/devices", nil)
	devicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if device, ok := findDevice(devices.Devices, enrollment.DeviceID); !ok || device.OwnerPrincipalID != "user-1" {
		t.Fatalf("enrolled device missing from devices response: %+v", devices.Devices)
	}
	devicePrincipal, err := aiAgentStore.AuthorizeDeviceCredential(context.Background(), enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll})
	if err != nil {
		t.Fatalf("AuthorizeDeviceCredential: %v", err)
	}
	if devicePrincipal.PrincipalID != "user-1" || devicePrincipal.WorkspaceID != "workspace-alpha" {
		t.Fatalf("device principal = %+v", devicePrincipal)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-enrolled","device_id":"`+enrollment.DeviceID+`","runtime_id":"runtime-codex-dev"}`))
	pollReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	pollReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}

	badPollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-enrolled","device_id":"`+enrollment.DeviceID+`","runtime_id":"runtime-codex-dev"}`))
	badPollReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	badPollReq.Header.Set(deviceSecretHeader, "wrong-secret")
	badPollResp := httptest.NewRecorder()
	server.ServeHTTP(badPollResp, badPollReq)
	if badPollResp.Code != http.StatusUnauthorized {
		t.Fatalf("bad poll status=%d body=%s", badPollResp.Code, badPollResp.Body.String())
	}

	codexRuntimeID := enrollment.DeviceID + ":codex"
	cursorRuntimeID := enrollment.DeviceID + ":cursor"
	codexSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-enrolled","runtimes":[{"runtime_id":"`+codexRuntimeID+`","kind":"codex","requires_experimental_opt_in":true}]}`))
	codexSnapshotReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	codexSnapshotReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	codexSnapshotResp := httptest.NewRecorder()
	server.ServeHTTP(codexSnapshotResp, codexSnapshotReq)
	if codexSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("codex snapshot status=%d body=%s", codexSnapshotResp.Code, codexSnapshotResp.Body.String())
	}

	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "enrolled codex",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  codexRuntimeID,
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-alpha/ai-agent/agents", strings.NewReader(string(createBody)))
	createReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	createResp := httptest.NewRecorder()
	server.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("create json: %v", err)
	}
	if created.Agent.RuntimeID != codexRuntimeID || created.Agent.RuntimeKind != RuntimeKindCodex {
		t.Fatalf("created agent = %+v", created.Agent)
	}

	cursorSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-enrolled","runtimes":[{"runtime_id":"`+cursorRuntimeID+`","kind":"cursor"}]}`))
	cursorSnapshotReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	cursorSnapshotReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	cursorSnapshotResp := httptest.NewRecorder()
	server.ServeHTTP(cursorSnapshotResp, cursorSnapshotReq)
	if cursorSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("cursor snapshot status=%d body=%s", cursorSnapshotResp.Code, cursorSnapshotResp.Body.String())
	}

	mergedDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	mergedDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	mergedDevicesResp := httptest.NewRecorder()
	server.ServeHTTP(mergedDevicesResp, mergedDevicesReq)
	if mergedDevicesResp.Code != http.StatusOK {
		t.Fatalf("merged devices status=%d body=%s", mergedDevicesResp.Code, mergedDevicesResp.Body.String())
	}
	var mergedDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(mergedDevicesResp.Body.Bytes(), &mergedDevices); err != nil {
		t.Fatalf("merged devices json: %v", err)
	}
	mergedDevice, ok := findDevice(mergedDevices.Devices, enrollment.DeviceID)
	if !ok {
		t.Fatalf("merged enrolled device missing: %+v", mergedDevices.Devices)
	}
	if !runtimeHasAssignedAgent(mergedDevices.Devices, codexRuntimeID) {
		t.Fatalf("codex runtime lost assigned-agent flag after cursor snapshot: %+v", mergedDevice.Runtimes)
	}
	codexRuntime, ok := findRuntime(mergedDevice.Runtimes, codexRuntimeID)
	if !ok || !codexRuntime.RequiresExperimentalOptIn {
		t.Fatalf("codex runtime opt-in fact missing after second snapshot: %+v", mergedDevice.Runtimes)
	}
	if _, ok := findRuntime(mergedDevice.Runtimes, cursorRuntimeID); !ok {
		t.Fatalf("cursor runtime missing after second snapshot: %+v", mergedDevice.Runtimes)
	}

	bindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	bindingsReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	bindingsReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	bindingsResp := httptest.NewRecorder()
	server.ServeHTTP(bindingsResp, bindingsReq)
	if bindingsResp.Code != http.StatusOK {
		t.Fatalf("bindings status=%d body=%s", bindingsResp.Code, bindingsResp.Body.String())
	}
	var bindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(bindingsResp.Body.Bytes(), &bindings); err != nil {
		t.Fatalf("bindings json: %v", err)
	}
	if len(bindings.Bindings) != 1 ||
		bindings.Bindings[0].AgentID != created.Agent.AgentID ||
		bindings.Bindings[0].RuntimeID != codexRuntimeID ||
		bindings.Bindings[0].RuntimeProvider != "codex" {
		t.Fatalf("bindings after second runtime snapshot = %+v", bindings.Bindings)
	}

	staleAt := time.Now().UTC().Add(-AIAgentDeviceRuntimeSnapshotStaleAfter - time.Second)
	aiAgentStore.mu.Lock()
	for deviceIndex := range aiAgentStore.devices {
		if aiAgentStore.devices[deviceIndex].DeviceID == enrollment.DeviceID {
			aiAgentStore.devices[deviceIndex].DaemonLastSeenAt = staleAt
			break
		}
	}
	if daemon, ok := aiAgentStore.daemons[enrollment.DeviceID]; ok {
		daemon.LastSeenAt = staleAt
		aiAgentStore.daemons[enrollment.DeviceID] = daemon
	}
	aiAgentStore.mu.Unlock()

	staleDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	staleDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	staleDevicesResp := httptest.NewRecorder()
	server.ServeHTTP(staleDevicesResp, staleDevicesReq)
	if staleDevicesResp.Code != http.StatusOK {
		t.Fatalf("stale devices status=%d body=%s", staleDevicesResp.Code, staleDevicesResp.Body.String())
	}
	var staleDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(staleDevicesResp.Body.Bytes(), &staleDevices); err != nil {
		t.Fatalf("stale devices json: %v", err)
	}
	staleDevice, ok := findDevice(staleDevices.Devices, enrollment.DeviceID)
	if !ok {
		t.Fatalf("stale device missing: %+v", staleDevices.Devices)
	}
	staleCodex, ok := findRuntime(staleDevice.Runtimes, codexRuntimeID)
	if !ok || staleCodex.Availability != RuntimeAvailabilityOffline || staleCodex.DetectionState != RuntimeDetectionStateMissing {
		t.Fatalf("stale codex runtime should project offline/missing: %+v", staleDevice.Runtimes)
	}

	staleDaemonReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/agents/"+created.Agent.AgentID+"/daemon", nil)
	staleDaemonReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	staleDaemonResp := httptest.NewRecorder()
	server.ServeHTTP(staleDaemonResp, staleDaemonReq)
	if staleDaemonResp.Code != http.StatusOK {
		t.Fatalf("stale daemon detail status=%d body=%s", staleDaemonResp.Code, staleDaemonResp.Body.String())
	}
	var staleDaemon DeviceDaemonDetailResponse
	if err := json.Unmarshal(staleDaemonResp.Body.Bytes(), &staleDaemon); err != nil {
		t.Fatalf("stale daemon json: %v", err)
	}
	if staleDaemon.Daemon.Availability != DaemonAvailabilityOffline ||
		staleDaemon.Daemon.ControlState != DaemonControlStateIdle ||
		!daemonSupportsAction(staleDaemon.Daemon, DaemonControlActionStart) ||
		daemonSupportsAction(staleDaemon.Daemon, DaemonControlActionStop) {
		t.Fatalf("stale daemon should project offline/start-only: %+v", staleDaemon.Daemon)
	}

	staleBindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	staleBindingsReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	staleBindingsReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	staleBindingsResp := httptest.NewRecorder()
	server.ServeHTTP(staleBindingsResp, staleBindingsReq)
	if staleBindingsResp.Code != http.StatusOK {
		t.Fatalf("stale bindings status=%d body=%s", staleBindingsResp.Code, staleBindingsResp.Body.String())
	}
	var staleBindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(staleBindingsResp.Body.Bytes(), &staleBindings); err != nil {
		t.Fatalf("stale bindings json: %v", err)
	}
	if len(staleBindings.Bindings) != 0 {
		t.Fatalf("stale runtime must be excluded from daemon bindings: %+v", staleBindings.Bindings)
	}

	replacementReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"display_name":"Replacement Mac","platform":"darwin","app_version":"0.0.2"}`))
	replacementReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	replacementResp := httptest.NewRecorder()
	server.ServeHTTP(replacementResp, replacementReq)
	if replacementResp.Code != http.StatusCreated {
		t.Fatalf("replacement enroll status=%d body=%s", replacementResp.Code, replacementResp.Body.String())
	}
	var replacement EnrollDeviceResponse
	if err := json.Unmarshal(replacementResp.Body.Bytes(), &replacement); err != nil {
		t.Fatalf("replacement json: %v", err)
	}
	if replacement.DeviceID == "" || replacement.DeviceID == enrollment.DeviceID {
		t.Fatalf("replacement enrollment = %+v", replacement)
	}
	moveSnapshotReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{"daemon_id":"daemon-replacement","runtimes":[{"runtime_id":"`+codexRuntimeID+`","kind":"codex","requires_experimental_opt_in":true}]}`))
	moveSnapshotReq.Header.Set(deviceIDHeader, replacement.DeviceID)
	moveSnapshotReq.Header.Set(deviceSecretHeader, replacement.DeviceSecret)
	moveSnapshotResp := httptest.NewRecorder()
	server.ServeHTTP(moveSnapshotResp, moveSnapshotReq)
	if moveSnapshotResp.Code != http.StatusAccepted {
		t.Fatalf("move snapshot status=%d body=%s", moveSnapshotResp.Code, moveSnapshotResp.Body.String())
	}
	afterMoveDevicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices", nil)
	afterMoveDevicesReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	afterMoveDevicesResp := httptest.NewRecorder()
	server.ServeHTTP(afterMoveDevicesResp, afterMoveDevicesReq)
	if afterMoveDevicesResp.Code != http.StatusOK {
		t.Fatalf("after move devices status=%d body=%s", afterMoveDevicesResp.Code, afterMoveDevicesResp.Body.String())
	}
	var afterMoveDevices DeviceRuntimeListResponse
	if err := json.Unmarshal(afterMoveDevicesResp.Body.Bytes(), &afterMoveDevices); err != nil {
		t.Fatalf("after move devices json: %v", err)
	}
	originalAfterMove, ok := findDevice(afterMoveDevices.Devices, enrollment.DeviceID)
	if !ok {
		t.Fatalf("original device missing after move: %+v", afterMoveDevices.Devices)
	}
	if _, ok := findRuntime(originalAfterMove.Runtimes, codexRuntimeID); ok {
		t.Fatalf("moved runtime must be removed from original device: %+v", originalAfterMove.Runtimes)
	}
	replacementAfterMove, ok := findDevice(afterMoveDevices.Devices, replacement.DeviceID)
	if !ok {
		t.Fatalf("replacement device missing after move: %+v", afterMoveDevices.Devices)
	}
	if _, ok := findRuntime(replacementAfterMove.Runtimes, codexRuntimeID); !ok {
		t.Fatalf("moved runtime missing from replacement device: %+v", replacementAfterMove.Runtimes)
	}
	replacementBindingsReq := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	replacementBindingsReq.Header.Set(deviceIDHeader, replacement.DeviceID)
	replacementBindingsReq.Header.Set(deviceSecretHeader, replacement.DeviceSecret)
	replacementBindingsResp := httptest.NewRecorder()
	server.ServeHTTP(replacementBindingsResp, replacementBindingsReq)
	if replacementBindingsResp.Code != http.StatusOK {
		t.Fatalf("replacement bindings status=%d body=%s", replacementBindingsResp.Code, replacementBindingsResp.Body.String())
	}
	var replacementBindings AgentRuntimeBindingListResponse
	if err := json.Unmarshal(replacementBindingsResp.Body.Bytes(), &replacementBindings); err != nil {
		t.Fatalf("replacement bindings json: %v", err)
	}
	if len(replacementBindings.Bindings) != 1 ||
		replacementBindings.Bindings[0].AgentID != created.Agent.AgentID ||
		replacementBindings.Bindings[0].DeviceID != replacement.DeviceID ||
		replacementBindings.Bindings[0].RuntimeID != codexRuntimeID {
		t.Fatalf("replacement bindings = %+v", replacementBindings.Bindings)
	}
}

func TestHTTPAIAgentClientExternalWorkspaceHidesDevelopmentSeedDevices(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "external-admin",
		Token:       "external-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}, {
		PrincipalID: "other-user",
		Token:       "other-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	enrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-real/devices/enroll", strings.NewReader(`{"display_name":"Real Mac","platform":"darwin","app_version":"0.0.0"}`))
	enrollReq.Header.Set(aiAgentTokenHeader, "external-token")
	enrollResp := httptest.NewRecorder()
	server.ServeHTTP(enrollResp, enrollReq)
	if enrollResp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollResp.Code, enrollResp.Body.String())
	}
	var enrollment EnrollDeviceResponse
	if err := json.Unmarshal(enrollResp.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}

	otherEnrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-other/devices/enroll", strings.NewReader(`{"display_name":"Other Mac","platform":"darwin","app_version":"0.0.0"}`))
	otherEnrollReq.Header.Set(aiAgentTokenHeader, "other-token")
	otherEnrollResp := httptest.NewRecorder()
	server.ServeHTTP(otherEnrollResp, otherEnrollReq)
	if otherEnrollResp.Code != http.StatusCreated {
		t.Fatalf("other enroll status=%d body=%s", otherEnrollResp.Code, otherEnrollResp.Body.String())
	}
	var otherEnrollment EnrollDeviceResponse
	if err := json.Unmarshal(otherEnrollResp.Body.Bytes(), &otherEnrollment); err != nil {
		t.Fatalf("other enroll json: %v", err)
	}

	syncBody := `{"daemon_id":"daemon-real","runtimes":[{"runtime_id":"agentd-real:codex","kind":"codex","availability":"online","detection_state":"detected","models":[{"model_id":"gpt-5.5","label":"gpt-5.5","is_default":true}]}]}`
	syncReq := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(syncBody))
	syncReq.Header.Set(deviceIDHeader, enrollment.DeviceID)
	syncReq.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	syncResp := httptest.NewRecorder()
	server.ServeHTTP(syncResp, syncReq)
	if syncResp.Code != http.StatusAccepted {
		t.Fatalf("runtime sync status=%d body=%s", syncResp.Code, syncResp.Body.String())
	}

	devicesReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-real/ai-agent/devices", nil)
	devicesReq.Header.Set(aiAgentTokenHeader, "external-token")
	devicesResp := httptest.NewRecorder()
	server.ServeHTTP(devicesResp, devicesReq)
	if devicesResp.Code != http.StatusOK {
		t.Fatalf("devices status=%d body=%s", devicesResp.Code, devicesResp.Body.String())
	}
	var devices DeviceRuntimeListResponse
	if err := json.Unmarshal(devicesResp.Body.Bytes(), &devices); err != nil {
		t.Fatalf("devices json: %v", err)
	}
	if len(devices.Devices) != 1 {
		t.Fatalf("external workspace devices = %+v, want only enrolled device", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, "device-dev-macbook"); ok {
		t.Fatalf("external workspace leaked device-dev-macbook: %+v", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, "device-shared-studio"); ok {
		t.Fatalf("external workspace leaked device-shared-studio: %+v", devices.Devices)
	}
	if _, ok := findDevice(devices.Devices, otherEnrollment.DeviceID); ok {
		t.Fatalf("external workspace leaked other enrolled device: %+v", devices.Devices)
	}
	realDevice, ok := findDevice(devices.Devices, enrollment.DeviceID)
	if !ok || realDevice.OwnerPrincipalID != "external-admin" || len(realDevice.Runtimes) != 1 || realDevice.Runtimes[0].RuntimeID != "agentd-real:codex" {
		t.Fatalf("external workspace real device = %+v, ok=%v", realDevice, ok)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-real/ai-agent/bootstrap?client_kind=web", nil)
	bootstrapReq.Header.Set(aiAgentTokenHeader, "external-token")
	bootstrapResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	if len(bootstrap.Devices) != 1 || bootstrap.Devices[0].DeviceID != enrollment.DeviceID {
		t.Fatalf("external workspace bootstrap devices = %+v", bootstrap.Devices)
	}
}

func TestHTTPAIAgentClientDevelopmentV2WorkspaceScopedCreateAndThreadStream(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	body, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "워크스페이스 A 에이전트",
		Visibility: AgentVisibilityPublic,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-fast"),
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-a/ai-agent/agents", strings.NewReader(string(body)))
	createReq.Header.Set(aiAgentTokenHeader, "owner-token")
	createResp := httptest.NewRecorder()
	server.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("v2 create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("v2 create json: %v", err)
	}
	if created.Agent.WorkspaceID != "workspace-a" ||
		created.Agent.OwnerPrincipalID != "user-1" ||
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.ModelID != "cursor-fast" {
		t.Fatalf("v2 created agent = %+v", created.Agent)
	}

	bootstrapAReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-a/ai-agent/bootstrap?client_kind=web", nil)
	bootstrapAReq.Header.Set(aiAgentTokenHeader, "owner-token")
	bootstrapAResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapAResp, bootstrapAReq)
	if bootstrapAResp.Code != http.StatusOK {
		t.Fatalf("v2 workspace-a bootstrap status=%d body=%s", bootstrapAResp.Code, bootstrapAResp.Body.String())
	}
	var bootstrapA ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapAResp.Body.Bytes(), &bootstrapA); err != nil {
		t.Fatalf("v2 workspace-a bootstrap json: %v", err)
	}
	if bootstrapA.WorkspaceID != "workspace-a" || bootstrapA.ClientKind != ClientKindWeb {
		t.Fatalf("v2 workspace-a bootstrap = %+v", bootstrapA)
	}
	if got, ok := findAIAgent(bootstrapA.Agents, created.Agent.AgentID); !ok || got.WorkspaceID != "workspace-a" {
		t.Fatalf("v2 workspace-a created agent = %+v, ok=%v", got, ok)
	}
	if _, ok := findAIAgent(bootstrapA.Agents, "agent-owned-codex"); ok {
		t.Fatalf("v2 workspace-a leaked default workspace agent: %+v", bootstrapA.Agents)
	}

	assignReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-a/ai-agent/tasks/task-v2/assignment", strings.NewReader(`{"agent_id":"`+created.Agent.AgentID+`"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "owner-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("v2 assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("v2 assign json: %v", err)
	}
	if assigned.AgentID != created.Agent.AgentID || assigned.ThreadID == "" {
		t.Fatalf("v2 assign response = %+v", assigned)
	}
	if assigned.ActiveStream == nil ||
		assigned.ActiveStream.Href != "/v2/client/workspaces/workspace-a/ai-agent/events" ||
		assigned.ActiveStream.ThreadID != assigned.ThreadID ||
		assigned.ActiveStream.RunID != assigned.RunID {
		t.Fatalf("v2 assign active_stream = %+v, assigned=%+v", assigned.ActiveStream, assigned)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/"+created.Agent.AgentID+"/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-cursor-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer owner-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("v2 assignment poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var poll PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &poll); err != nil {
		t.Fatalf("v2 assignment poll json: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil ||
		poll.Assignment.TaskID != "task-v2" ||
		poll.Assignment.ComponentID != "task-v2" ||
		poll.Assignment.AgentID != created.Agent.AgentID ||
		poll.Assignment.RuntimeProvider != "cursor" ||
		!strings.Contains(poll.Assignment.AgentInstruction, created.Agent.Instruction) ||
		!strings.Contains(poll.Assignment.AgentInstruction, "한국어") ||
		!strings.Contains(poll.Assignment.Prompt, "branch_name: RIID-4800-server-task-context-http-client-assignment-prompt-wiring") {
		if poll.Assignment != nil {
			t.Fatalf("v2 assignment poll action=%s assignment=%+v", poll.Action, *poll.Assignment)
		}
		t.Fatalf("v2 assignment poll = %+v", poll)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-a/ai-agent/tasks/task-v2/threads", nil)
	threadsReq.Header.Set(aiAgentTokenHeader, "owner-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("v2 threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("v2 threads json: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].AgentID != created.Agent.AgentID {
		t.Fatalf("v2 threads = %+v", threads)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.Href != "/v2/client/workspaces/workspace-a/ai-agent/events" {
		t.Fatalf("v2 active stream = %+v", threads.ActiveStream)
	}

	bootstrapBReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-b/ai-agent/bootstrap", nil)
	bootstrapBReq.Header.Set(aiAgentTokenHeader, "owner-token")
	bootstrapBResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapBResp, bootstrapBReq)
	if bootstrapBResp.Code != http.StatusOK {
		t.Fatalf("v2 workspace-b bootstrap status=%d body=%s", bootstrapBResp.Code, bootstrapBResp.Body.String())
	}
	var bootstrapB ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapBResp.Body.Bytes(), &bootstrapB); err != nil {
		t.Fatalf("v2 workspace-b bootstrap json: %v", err)
	}
	if _, ok := findAIAgent(bootstrapB.Agents, created.Agent.AgentID); ok {
		t.Fatalf("v2 workspace-b leaked workspace-a agent: %+v", bootstrapB.Agents)
	}
}

func TestHTTPAIAgentClientAssignUsesRequestScopedTaskContext(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-private:assign", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	aiAgentStore.devices[0].Runtimes[0].RequiresExperimentalOptIn = true
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	taskContext := &assignmentHTTPRequestTaskContextReader{
		contextSnapshot: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-private",
				ComponentType: "task",
				Title:         "JWT로 task context 조회",
				KeyNumber:     "RIID-4873",
				BranchName:    "RIID-4964-agent-profile-upload",
			},
			Document: AIAgentTaskContextDocument{
				Content:       "<p>private JWT component body</p>",
				ContentFormat: "html",
			},
			Hierarchy: AIAgentTaskContextHierarchy{
				Project:   AIAgentTaskContextReference{ID: "project-a", Title: "AI Agent", KeyNumber: "RIID-4800"},
				Milestone: AIAgentTaskContextReference{ID: "milestone-a", Title: "JWT task context", KeyNumber: "RIID-4872"},
			},
			Repositories: []AIAgentTaskContextRepository{{
				FullName:      "teamswyg/riido-daemon",
				RepositoryURL: "https://github.com/teamswyg/riido-daemon",
				Source:        TaskContextRepositorySourceConnectedPullRequest,
			}},
		},
	}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   taskContext,
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-private/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	if len(taskContext.requests) != 1 {
		t.Fatalf("task context requests = %+v", taskContext.requests)
	}
	gotContextReq := taskContext.requests[0]
	if gotContextReq.ComponentID != "task-private" ||
		gotContextReq.WorkspaceID != "workspace-dev-riid" ||
		gotContextReq.BearerToken != "user-token" {
		t.Fatalf("task context request = %+v", gotContextReq)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer user-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var pollOut PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &pollOut); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if pollOut.Action != PollStart || pollOut.Assignment == nil {
		t.Fatalf("poll response = %+v", pollOut)
	}
	if !strings.Contains(pollOut.Assignment.Prompt, "private JWT component body") ||
		!strings.Contains(pollOut.Assignment.Prompt, "JWT로 task context 조회") {
		t.Fatalf("assignment prompt = %s", pollOut.Assignment.Prompt)
	}
	if !pollOut.Assignment.AllowExperimentalRuntime {
		t.Fatalf("assignment experimental opt-in = false, assignment=%+v", *pollOut.Assignment)
	}
	if pollOut.Assignment.ModelID != "codex-default" {
		t.Fatalf("assignment model_id = %q, want codex-default", pollOut.Assignment.ModelID)
	}
	if pollOut.Assignment.Worktree == nil ||
		pollOut.Assignment.Worktree.RepositoryFullName != "teamswyg/riido-daemon" ||
		pollOut.Assignment.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" ||
		pollOut.Assignment.Worktree.BranchName != "RIID-4964-agent-profile-upload" ||
		pollOut.Assignment.Worktree.Source != TaskContextRepositorySourceConnectedPullRequest {
		t.Fatalf("assignment worktree = %+v", pollOut.Assignment.Worktree)
	}
}

func TestHTTPAIAgentClientAssignFallsBackWhenTaskContextUnavailable(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-smoke:assign", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	taskContext := &assignmentHTTPRequestTaskContextReader{err: errors.New("private task context returned HTTP 401")}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   taskContext,
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-smoke/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	if len(taskContext.requests) != 1 {
		t.Fatalf("task context requests = %+v", taskContext.requests)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer user-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var pollOut PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &pollOut); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if pollOut.Action != PollStart || pollOut.Assignment == nil {
		t.Fatalf("poll response = %+v", pollOut)
	}
	if !strings.Contains(pollOut.Assignment.Prompt, "Task context was not available when this assignment was created.") ||
		!strings.Contains(pollOut.Assignment.Prompt, "- task_id: task-smoke") {
		t.Fatalf("assignment prompt = %s", pollOut.Assignment.Prompt)
	}
	if strings.Contains(pollOut.Assignment.Prompt, "private task context returned HTTP 401") {
		t.Fatalf("assignment prompt leaked task context error: %s", pollOut.Assignment.Prompt)
	}
}

func TestHTTPAIAgentClientDevelopmentAssignableAgentsUseStableIDTieBreak(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
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

func TestHTTPAIAgentClientDevelopmentDevicesAndEditability(t *testing.T) {
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
	ownedDevice, ok := findDevice(devices.Devices, "device-dev-macbook")
	if !ok || len(ownedDevice.Runtimes) != 3 {
		t.Fatalf("owned device = %+v, ok=%v", ownedDevice, ok)
	}
	sharedDevice, ok := findDevice(devices.Devices, "device-shared-studio")
	if !ok || len(sharedDevice.Runtimes) != 1 || sharedDevice.Runtimes[0].RuntimeID != "runtime-openclaw-shared" {
		t.Fatalf("shared public-agent device = %+v, ok=%v", sharedDevice, ok)
	}
	cursorRuntime := ownedDevice.Runtimes[2]
	if cursorRuntime.RuntimeID != "runtime-cursor-dev" || len(cursorRuntime.Models) != 2 || cursorRuntime.Models[0].ModelID != "cursor-auto" || !cursorRuntime.Models[0].IsDefault {
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

func TestHTTPAIAgentClientDevelopmentDoesNotExposeWaitlistMarketingMutation(t *testing.T) {
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

func TestHTTPAIAgentClientDevelopmentDeviceDaemonDetailAndControl(t *testing.T) {
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
	if detail.Runtime == nil ||
		detail.Runtime.RuntimeID != "runtime-codex-dev" ||
		detail.Runtime.ProviderVersion != "codex-cli 0.133.0" {
		t.Fatalf("daemon detail runtime = %+v", detail.Runtime)
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
	if publicDetail.Runtime == nil ||
		publicDetail.Runtime.RuntimeID != "runtime-openclaw-shared" ||
		publicDetail.Runtime.ProviderVersion != "openclaw 0.1.0" {
		t.Fatalf("public daemon detail runtime = %+v", publicDetail.Runtime)
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

func TestHTTPAIAgentClientDevelopmentTaskCommentAndStop(t *testing.T) {
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
	if comment.ActiveStream == nil || comment.ActiveStream.Href != "/v1/client/ai-agent/events" {
		t.Fatalf("comment active_stream = %+v", comment.ActiveStream)
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
	if stopped.ActiveStream != nil {
		t.Fatalf("terminal stop must omit active_stream: %+v", stopped.ActiveStream)
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

func TestHTTPAIAgentClientDevelopmentTaskAssignmentAndParticipantRemoval(t *testing.T) {
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

func TestHTTPAIAgentClientStopCancelsDurableAssignmentForDaemonPoll(t *testing.T) {
	ctx := context.Background()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-stop-durable:read", "task:task-stop-durable:assign", "task:task-stop-durable:stop"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-durable/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
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
	if assigned.AssignmentID == "" || assigned.AgentID != "agent-public-openclaw" {
		t.Fatalf("assign response = %+v", assigned)
	}
	openClawPollRequest := PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	}
	pollStart, err := assignmentStore.PollAgent(ctx, assigned.AgentID, openClawPollRequest)
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if pollStart.Action != PollStart || pollStart.Assignment == nil || pollStart.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll start = %+v, assigned=%+v", pollStart, assigned)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-durable/stop", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"user requested stop"}`))
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
	if stopped.AssignmentID != assigned.AssignmentID ||
		stopped.AssignmentState != AgentAssignmentStateStopping ||
		stopped.WorkStatus != AgentWorkStatusRunning ||
		stopped.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("stop response = %+v, assigned=%+v", stopped, assigned)
	}
	projection, ok, err := assignmentStore.LoadAssignmentProjection(ctx, assigned.AssignmentID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection: %v", err)
	}
	if !ok || projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("projection after stop = %+v ok=%v", projection, ok)
	}
	pollCancel, err := assignmentStore.PollAgent(ctx, assigned.AgentID, openClawPollRequest)
	if err != nil {
		t.Fatalf("PollAgent cancel: %v", err)
	}
	if pollCancel.Action != PollCancel || pollCancel.Assignment == nil || pollCancel.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll cancel = %+v", pollCancel)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-stop-durable/threads", nil)
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
	if threads.ActiveStream == nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateStopping {
		t.Fatalf("threads after stop = %+v", threads)
	}

	cancelledEvent, err := assignmentStore.RecordAgentEvent(ctx, assigned.AgentID, AgentEventRequest{
		AssignmentID: assigned.AssignmentID,
		TaskID:       assigned.TaskID,
		DaemonID:     openClawPollRequest.DaemonID,
		DeviceID:     openClawPollRequest.DeviceID,
		RuntimeID:    openClawPollRequest.RuntimeID,
		State:        AssignmentCancelled,
		EventType:    EventAssignmentCancelled,
		Message:      "provider cancelled after client stop",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent cancel: %v", err)
	}
	if err := aiAgentStore.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, cancelledEvent.Event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent cancel: %v", err)
	}
	threadsResp = httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads after cancel status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	threads = AIAgentTaskThreadCollectionResponse{}
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads after cancel json: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("threads after cancel = %+v", threads)
	}
}

func TestHTTPAIAgentClientStopDoesNotMutateThreadWhenDurableCancelFails(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := aiAgentStore.AssignAIAgentTask(ctx, principal, "task-stop-cancel-fails", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-missing-durable",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if assigned.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup assignment should be running: %+v", assigned)
	}

	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, principal.PrincipalID),
	}).Handler()
	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-cancel-fails/stop", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"user requested stop"}`))
	stopReq.Header.Set("Authorization", "Bearer ai-agent-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code == http.StatusAccepted {
		t.Fatalf("stop should fail when durable cancel fails: status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	if !strings.Contains(stopResp.Body.String(), "assignment asn-missing-durable not found") {
		t.Fatalf("stop error should expose durable cancel failure, status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}

	threads, err := aiAgentStore.ListAIAgentTaskThreads(ctx, principal, "task-stop-cancel-fails")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads)
	}
	thread := threads.Threads[0]
	if !taskThreadHasActiveStream(thread) ||
		thread.AssignmentState != AgentAssignmentStateRunning ||
		thread.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("failed durable cancel must not pre-stop read model thread: %+v", thread)
	}
}

func TestHTTPAIAgentClientDevelopmentTaskThreadColdCollectionAfterViewerAwayAssignment(t *testing.T) {
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

func TestHTTPAIAgentClientDevelopmentMutationAndDeletion(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	thumbnailURL := "https://cdn.riido.io/dev/ai-agents/updated-claude.png"
	description := strings.Repeat("설", AgentDescriptionMaxCharacters)
	instruction := strings.Repeat("지", AgentInstructionMaxCharacters)
	createBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:                "신규 코리",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-dev",
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

	fixtureDescription := "서버 구조를 설계하고, API와 데이터 흐름을 안정적으로 구현합니다."
	fixtureInstruction := "fixture 선택 후에도 client가 agent 생성에 들어가는 값을 모두 담아 보냅니다."
	fixtureBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:                "영실",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-fast"),
		ProfileThumbnailURL: stringPtr("https://cdn.riido.io/dev/ai-agent-fixtures/yeongsil-backend.png"),
		Description:         &fixtureDescription,
		Instruction:         &fixtureInstruction,
	})
	if err != nil {
		t.Fatalf("marshal fixture create body: %v", err)
	}
	duplicateRuntimeReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/yeongsil_backend/agents", strings.NewReader(string(fixtureBody)))
	duplicateRuntimeReq.Header.Set(aiAgentTokenHeader, "owner-token")
	duplicateRuntimeResp := httptest.NewRecorder()
	server.ServeHTTP(duplicateRuntimeResp, duplicateRuntimeReq)
	if duplicateRuntimeResp.Code != http.StatusConflict {
		t.Fatalf("duplicate runtime status=%d body=%s", duplicateRuntimeResp.Code, duplicateRuntimeResp.Body.String())
	}
	deleteCreatedReq := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/agents/"+created.Agent.AgentID, nil)
	deleteCreatedReq.Header.Set(aiAgentTokenHeader, "owner-token")
	deleteCreatedResp := httptest.NewRecorder()
	server.ServeHTTP(deleteCreatedResp, deleteCreatedReq)
	if deleteCreatedResp.Code != http.StatusOK {
		t.Fatalf("delete created status=%d body=%s", deleteCreatedResp.Code, deleteCreatedResp.Body.String())
	}
	fixtureCreateReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/yeongsil_backend/agents", strings.NewReader(string(fixtureBody)))
	fixtureCreateReq.Header.Set(aiAgentTokenHeader, "owner-token")
	fixtureCreateResp := httptest.NewRecorder()
	server.ServeHTTP(fixtureCreateResp, fixtureCreateReq)
	if fixtureCreateResp.Code != http.StatusCreated {
		t.Fatalf("fixture create status=%d body=%s", fixtureCreateResp.Code, fixtureCreateResp.Body.String())
	}
	var fixtureCreated AgentClientRecordResponse
	if err := json.Unmarshal(fixtureCreateResp.Body.Bytes(), &fixtureCreated); err != nil {
		t.Fatalf("fixture create json: %v", err)
	}
	if fixtureCreated.Agent.Name != "영실" ||
		fixtureCreated.Agent.RuntimeKind != RuntimeKindCursor ||
		fixtureCreated.Agent.ModelID != "cursor-fast" ||
		fixtureCreated.Agent.Description != fixtureDescription ||
		fixtureCreated.Agent.Instruction != fixtureInstruction ||
		fixtureCreated.Agent.ProfileThumbnailURL == "" {
		t.Fatalf("fixture created agent = %+v", fixtureCreated.Agent)
	}

	duplicateFixtureReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/yeongsil_backend/agents", strings.NewReader(string(fixtureBody)))
	duplicateFixtureReq.Header.Set(aiAgentTokenHeader, "owner-token")
	duplicateFixtureResp := httptest.NewRecorder()
	server.ServeHTTP(duplicateFixtureResp, duplicateFixtureReq)
	if duplicateFixtureResp.Code != http.StatusConflict {
		t.Fatalf("duplicate fixture create status=%d body=%s", duplicateFixtureResp.Code, duplicateFixtureResp.Body.String())
	}

	missingFixtureReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/missing_fixture/agents", strings.NewReader(string(fixtureBody)))
	missingFixtureReq.Header.Set(aiAgentTokenHeader, "owner-token")
	missingFixtureResp := httptest.NewRecorder()
	server.ServeHTTP(missingFixtureResp, missingFixtureReq)
	if missingFixtureResp.Code != http.StatusNotFound {
		t.Fatalf("missing fixture status=%d body=%s", missingFixtureResp.Code, missingFixtureResp.Body.String())
	}

	patchBody, err := json.Marshal(UpdateAgentConfigurationRequest{
		Name:                "같은 이름 가능",
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-fast"),
		ProfileThumbnailURL: &thumbnailURL,
		Description:         &description,
		Instruction:         &instruction,
	})
	if err != nil {
		t.Fatalf("marshal patch body: %v", err)
	}
	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/"+fixtureCreated.Agent.AgentID, strings.NewReader(string(patchBody)))
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
	updated, ok := findAIAgent(bootstrap.Agents, fixtureCreated.Agent.AgentID)
	if !ok || updated.ProfileThumbnailURL != thumbnailURL || updated.Description != description || updated.Instruction != instruction || !updated.CreatedAt.Equal(patched.Agent.CreatedAt) || !updated.UpdatedAt.Equal(patched.Agent.UpdatedAt) {
		t.Fatalf("bootstrap updated agent = %+v found=%v", updated, ok)
	}
	if !runtimeHasAssignedAgent(bootstrap.Devices, "runtime-cursor-dev") {
		t.Fatalf("bootstrap runtime-cursor-dev was not marked assigned: %+v", bootstrap.Devices)
	}

	invalidModelBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "잘못된 모델",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
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

	trackedThumbnailURL := "https://cdn.riido.io/dev/ai-agents/tracked.png?token=secret"
	trackedThumbnailBody, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:                "추적 URL",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-dev",
		ProfileThumbnailURL: &trackedThumbnailURL,
	})
	if err != nil {
		t.Fatalf("marshal tracked thumbnail body: %v", err)
	}
	trackedThumbnailReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(trackedThumbnailBody)))
	trackedThumbnailReq.Header.Set("Authorization", "Bearer owner-token")
	trackedThumbnailResp := httptest.NewRecorder()
	server.ServeHTTP(trackedThumbnailResp, trackedThumbnailReq)
	if trackedThumbnailResp.Code != http.StatusBadRequest {
		t.Fatalf("tracked thumbnail create status=%d body=%s", trackedThumbnailResp.Code, trackedThumbnailResp.Body.String())
	}

	fragmentThumbnailURL := "https://cdn.riido.io/dev/ai-agents/fragment.png#secret"
	fragmentThumbnailBody, err := json.Marshal(UpdateAgentConfigurationRequest{ProfileThumbnailURL: &fragmentThumbnailURL})
	if err != nil {
		t.Fatalf("marshal fragment thumbnail patch body: %v", err)
	}
	fragmentThumbnailReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(string(fragmentThumbnailBody)))
	fragmentThumbnailReq.Header.Set("Authorization", "Bearer owner-token")
	fragmentThumbnailResp := httptest.NewRecorder()
	server.ServeHTTP(fragmentThumbnailResp, fragmentThumbnailReq)
	if fragmentThumbnailResp.Code != http.StatusBadRequest {
		t.Fatalf("fragment thumbnail patch status=%d body=%s", fragmentThumbnailResp.Code, fragmentThumbnailResp.Body.String())
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

func TestHTTPAIAgentClientDevelopmentAdminCreateUsesAuthorizedWorkspaceRuntime(t *testing.T) {
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
	ownedDevice, ok := findDevice(devices.Devices, "device-dev-macbook")
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
		RuntimeID:  "runtime-cursor-dev",
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
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.RuntimeKind != RuntimeKindCursor ||
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

func TestHTTPAIAgentClientDevelopmentSSEReplaysTypedCommentStatus(t *testing.T) {
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
		AIAgentClient: NewDevelopmentAIAgentClientStore(),
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

func TestHTTPAgentEventsUpdateAIAgentTaskThreadReadModel(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}

	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil || poll.Assignment.ID == "" {
		t.Fatalf("poll response = %+v", poll)
	}

	readyBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"ready","event_type":"assignment_ready","message":"runtime accepted assignment"}`
	readyReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(readyBody))
	readyReq.Header.Set("Authorization", "Bearer daemon-token")
	readyResp := httptest.NewRecorder()
	handler.ServeHTTP(readyResp, readyReq)
	if readyResp.Code != http.StatusOK {
		t.Fatalf("ready event status=%d body=%s", readyResp.Code, readyResp.Body.String())
	}

	readyThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	readyThreadsReq.Header.Set("Authorization", "Bearer user-token")
	readyThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(readyThreadsResp, readyThreadsReq)
	if readyThreadsResp.Code != http.StatusOK {
		t.Fatalf("ready threads status=%d body=%s", readyThreadsResp.Code, readyThreadsResp.Body.String())
	}
	var readyThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(readyThreadsResp.Body.Bytes(), &readyThreads); err != nil {
		t.Fatalf("ready threads json: %v", err)
	}
	if len(readyThreads.Threads) != 1 ||
		readyThreads.ActiveStream == nil ||
		readyThreads.ActiveStream.ThreadID != assigned.ThreadID ||
		readyThreads.Threads[0].WorkStatus != AgentWorkStatusRunning ||
		readyThreads.Threads[0].AssignmentState != AgentAssignmentStateRunning ||
		readyThreads.Threads[0].CommentKind != AgentTaskCommentAssignmentStarted {
		t.Fatalf("threads after ready event = %+v", readyThreads)
	}

	logBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"running","event_type":"riido_log","message":"팀 프로젝트 수집 중 - 진행 상태를 조회 중.","metadata":{"` + metadatakeys.ThreadProgressSeq.String() + `":"1"}}`
	logReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(logBody))
	logReq.Header.Set("Authorization", "Bearer daemon-token")
	logResp := httptest.NewRecorder()
	handler.ServeHTTP(logResp, logReq)
	if logResp.Code != http.StatusOK {
		t.Fatalf("log event status=%d body=%s", logResp.Code, logResp.Body.String())
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 ||
		threads.ActiveStream == nil ||
		threads.ActiveStream.ThreadID != assigned.ThreadID ||
		len(threads.Threads[0].Lines) != 1 ||
		threads.Threads[0].Lines[0].Message != "팀 프로젝트 수집 중 - 진행 상태를 조회 중." {
		t.Fatalf("threads after log event = %+v", threads)
	}

	progressBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","run_id":"` + poll.Assignment.ID + `","lines":[{"seq":2,"message":"파일 생성 중 - 산출물을 작성 중."}]}`
	progressReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/thread-progress", strings.NewReader(progressBody))
	progressReq.Header.Set("Authorization", "Bearer daemon-token")
	progressResp := httptest.NewRecorder()
	handler.ServeHTTP(progressResp, progressReq)
	if progressResp.Code != http.StatusAccepted {
		t.Fatalf("progress event status=%d body=%s", progressResp.Code, progressResp.Body.String())
	}
	var progress AgentThreadProgressBatchResponse
	if err := json.Unmarshal(progressResp.Body.Bytes(), &progress); err != nil {
		t.Fatalf("progress json: %v", err)
	}
	if progress.Event.ThreadID != assigned.ThreadID || progress.Event.RunID != poll.Assignment.ID {
		t.Fatalf("thread-progress should reconcile to assignment thread: assigned=%+v progress=%+v", assigned, progress.Event)
	}

	progressThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	progressThreadsReq.Header.Set("Authorization", "Bearer user-token")
	progressThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(progressThreadsResp, progressThreadsReq)
	if progressThreadsResp.Code != http.StatusOK {
		t.Fatalf("progress threads status=%d body=%s", progressThreadsResp.Code, progressThreadsResp.Body.String())
	}
	var progressThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(progressThreadsResp.Body.Bytes(), &progressThreads); err != nil {
		t.Fatalf("progress threads json: %v", err)
	}
	if len(progressThreads.Threads) != 1 ||
		progressThreads.ActiveStream == nil ||
		progressThreads.ActiveStream.ThreadID != assigned.ThreadID ||
		progressThreads.Threads[0].ThreadID != assigned.ThreadID ||
		progressThreads.Threads[0].RunID != poll.Assignment.ID ||
		len(progressThreads.Threads[0].Lines) != 2 {
		t.Fatalf("threads after thread-progress = %+v", progressThreads)
	}

	completedBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"completed","event_type":"assignment_completed","message":"작업 완료"}`
	completedReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(completedBody))
	completedReq.Header.Set("Authorization", "Bearer daemon-token")
	completedResp := httptest.NewRecorder()
	handler.ServeHTTP(completedResp, completedReq)
	if completedResp.Code != http.StatusOK {
		t.Fatalf("completed event status=%d body=%s", completedResp.Code, completedResp.Body.String())
	}

	completedThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	completedThreadsReq.Header.Set("Authorization", "Bearer user-token")
	completedThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(completedThreadsResp, completedThreadsReq)
	if completedThreadsResp.Code != http.StatusOK {
		t.Fatalf("completed threads status=%d body=%s", completedThreadsResp.Code, completedThreadsResp.Body.String())
	}
	var completedThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(completedThreadsResp.Body.Bytes(), &completedThreads); err != nil {
		t.Fatalf("completed threads json: %v", err)
	}
	if completedThreads.ActiveStream != nil ||
		len(completedThreads.Threads) != 1 ||
		completedThreads.Threads[0].ThreadID != assigned.ThreadID ||
		completedThreads.Threads[0].AssignmentState != AgentAssignmentStateCompleted ||
		completedThreads.Threads[0].CommentKind != AgentTaskCommentTaskCompleted ||
		completedThreads.Threads[0].Message != "작업 완료" ||
		completedThreads.Threads[0].CompletedAt.IsZero() ||
		len(completedThreads.Threads[0].Lines) != 2 {
		t.Fatalf("completed threads = %+v", completedThreads)
	}

	followupBody := `{"body":"README.md를 추가하고 실행 방법을 정리해 주세요.","source_message_id":"message-followup-1"}`
	followupReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/threads/"+assigned.ThreadID+"/messages", strings.NewReader(followupBody))
	followupReq.Header.Set("Authorization", "Bearer user-token")
	followupResp := httptest.NewRecorder()
	handler.ServeHTTP(followupResp, followupReq)
	if followupResp.Code != http.StatusAccepted {
		t.Fatalf("followup status=%d body=%s", followupResp.Code, followupResp.Body.String())
	}
	var followup AIAgentTaskActionResponse
	if err := json.Unmarshal(followupResp.Body.Bytes(), &followup); err != nil {
		t.Fatalf("followup json: %v", err)
	}
	if followup.ThreadID != assigned.ThreadID ||
		followup.AgentID != assigned.AgentID ||
		followup.WorkStatus != AgentWorkStatusRunning ||
		followup.AssignmentState != AgentAssignmentStateRunning ||
		followup.CommentKind != AgentTaskCommentRuntimeProgress {
		t.Fatalf("followup response = %+v", followup)
	}
	followupPoll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent followup: %v", err)
	}
	if followupPoll.Assignment == nil ||
		followupPoll.Assignment.AgentID != assigned.AgentID ||
		!strings.Contains(followupPoll.Assignment.Prompt, "README.md를 추가하고 실행 방법을 정리해 주세요.") ||
		!strings.Contains(followupPoll.Assignment.Prompt, "## Follow-up Thread Message") ||
		!strings.Contains(followupPoll.Assignment.Prompt, assigned.ThreadID) {
		t.Fatalf("followup poll assignment = %+v", followupPoll.Assignment)
	}
	followupThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	followupThreadsReq.Header.Set("Authorization", "Bearer user-token")
	followupThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(followupThreadsResp, followupThreadsReq)
	if followupThreadsResp.Code != http.StatusOK {
		t.Fatalf("followup threads status=%d body=%s", followupThreadsResp.Code, followupThreadsResp.Body.String())
	}
	var followupThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(followupThreadsResp.Body.Bytes(), &followupThreads); err != nil {
		t.Fatalf("followup threads json: %v", err)
	}
	if followupThreads.ActiveStream == nil ||
		followupThreads.ActiveStream.ThreadID != assigned.ThreadID ||
		len(followupThreads.Threads) != 1 ||
		followupThreads.Threads[0].ThreadID != assigned.ThreadID ||
		followupThreads.Threads[0].SourceMessageID != "message-followup-1" {
		t.Fatalf("followup threads = %+v", followupThreads)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	handler.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	eventsBody := eventsResp.Body.String()
	if !strings.Contains(eventsBody, "event: agent_thread_progress\n") ||
		!strings.Contains(eventsBody, "event: agent_work_status_changed\n") {
		t.Fatalf("events body = %q", eventsBody)
	}
}

func TestHTTPAgentHeartbeatStaleLeaseClosesAIAgentTaskThreadReadModel(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 2, 0, 0, 0, time.UTC)
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	operations := &runtimeFakeActiveLeaseOperationStore{}
	assignmentStore := NewStoreWithConfig(StoreConfig{
		AgentRegistry:       aiAgentStore,
		ActiveLeaseDuration: 20 * time.Second,
		Now:                 func() time.Time { return now },
		OperationStore:      operations,
	})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-stale:read", "task:task-stale:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:heartbeat"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stale/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}

	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil || poll.Assignment.ID == "" || poll.Assignment.LeaseToken == "" {
		t.Fatalf("poll response = %+v", poll)
	}
	operations.activeFound = true
	operations.activeLease = AssignmentActiveLease{
		AgentID:            poll.Assignment.AgentID,
		ActiveAssignmentID: poll.Assignment.ID,
		LeaseToken:         poll.Assignment.LeaseToken,
		HeartbeatAt:        now,
		LeaseExpiresAt:     now.Add(20 * time.Second),
		LeaseExpiresUnixMS: now.Add(20 * time.Second).UnixMilli(),
	}
	now = now.Add(21 * time.Second)

	heartbeatBody := `{"daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","active_assignment_ids":["` + poll.Assignment.ID + `"],"running_task_ids":["task-stale"]}`
	heartbeatReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/heartbeat", strings.NewReader(heartbeatBody))
	heartbeatReq.Header.Set("Authorization", "Bearer daemon-token")
	heartbeatResp := httptest.NewRecorder()
	handler.ServeHTTP(heartbeatResp, heartbeatReq)
	if heartbeatResp.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", heartbeatResp.Code, heartbeatResp.Body.String())
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-stale/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].ThreadID != assigned.ThreadID ||
		threads.Threads[0].WorkStatus != AgentWorkStatusFailed ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateFailed ||
		threads.Threads[0].CommentKind != AgentTaskCommentTaskFailed ||
		!strings.Contains(threads.Threads[0].Message, "active assignment lease expired") {
		t.Fatalf("threads after stale heartbeat = %+v", threads)
	}
}

func TestHTTPTaskThreadListRepairsStaleReadModelFromAssignmentProjection(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-repair:read", "task:task-repair:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assigned.AssignmentID == "" {
		t.Fatalf("assigned response must expose assignment_id: %+v", assigned)
	}
	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil || poll.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll response = %+v, assigned=%+v", poll, assigned)
	}

	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentFailed,
		EventType:    EventAssignmentFailed,
		Message:      "provider process exited before client read-model update",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	beforeRepair, err := aiAgentStore.ListAIAgentTaskThreads(ctx, AuthorizationResult{PrincipalID: "user-1"}, "task-repair")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads before repair: %v", err)
	}
	if beforeRepair.ActiveStream == nil || beforeRepair.Threads[0].AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup should leave stale active client read-model: %+v", beforeRepair)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-repair/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].ThreadID != assigned.ThreadID ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateFailed ||
		threads.Threads[0].CommentKind != AgentTaskCommentTaskFailed ||
		threads.Threads[0].CompletedAt.IsZero() {
		t.Fatalf("threads after projection repair = %+v", threads)
	}
}

func TestHTTPBootstrapRepairsStaleReadModelFromAssignmentProjection(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-bootstrap-repair:read", "task:task-bootstrap-repair:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-bootstrap-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-bootstrap-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running before terminal projection",
	}); err != nil {
		t.Fatalf("RecordAgentEvent running: %v", err)
	}
	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-bootstrap-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentCompleted,
		EventType:    EventAssignmentCompleted,
		Message:      "completed before bootstrap repaired the client read-model",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	staleBootstrap, err := aiAgentStore.BootstrapAIAgentClient(ctx, AuthorizationResult{PrincipalID: "user-1"}, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient before repair: %v", err)
	}
	if stale := agentByID(staleBootstrap.Agents, "agent-public-openclaw"); stale.AssignedTaskCount != 1 || stale.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("test setup should leave bootstrap stale before HTTP repair: %+v", stale)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer user-token")
	bootstrapResp := httptest.NewRecorder()
	handler.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	repaired := agentByID(bootstrap.Agents, "agent-public-openclaw")
	if repaired.AssignedTaskCount != 0 ||
		repaired.Editability != AgentEditabilityEditable ||
		repaired.WorkStatus != AgentWorkStatusCompleted {
		t.Fatalf("bootstrap should repair stale assignment projection: assigned=%+v repaired=%+v", assigned, repaired)
	}
}

func TestHTTPBootstrapRepairsQueuedProjectionFromEagerRunningReadModel(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-queued-repair:read", "task:task-queued-repair:assign"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-queued-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assigned.WorkStatus != AgentWorkStatusRunning || assigned.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup should create eager running read-model before daemon poll: %+v", assigned)
	}
	staleBootstrap, err := aiAgentStore.BootstrapAIAgentClient(ctx, AuthorizationResult{PrincipalID: "user-1"}, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient before repair: %v", err)
	}
	if stale := agentByID(staleBootstrap.Agents, "agent-public-openclaw"); stale.AssignedTaskCount != 1 || stale.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("test setup should leave bootstrap eager-running before HTTP repair: %+v", stale)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer user-token")
	bootstrapResp := httptest.NewRecorder()
	handler.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	repaired := agentByID(bootstrap.Agents, "agent-public-openclaw")
	if repaired.AssignedTaskCount != 1 ||
		repaired.Editability != AgentEditabilityBlockedAssignedTasks ||
		repaired.WorkStatus != AgentWorkStatusQueued {
		t.Fatalf("bootstrap should repair eager running read-model to queued projection: assigned=%+v repaired=%+v", assigned, repaired)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-queued-repair/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream == nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].ThreadID != assigned.ThreadID ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateQueued ||
		threads.Threads[0].WorkStatus != AgentWorkStatusQueued ||
		threads.Threads[0].CommentKind != AgentTaskCommentQueuedByBusyAgent ||
		!threads.Threads[0].CompletedAt.IsZero() {
		t.Fatalf("threads after queued projection repair = %+v", threads)
	}
}

func TestHTTPAIAgentClientDevelopmentRequiresConfigurationAndAuthorization(t *testing.T) {
	unconfigured := NewServer(ServerConfig{Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	unconfigured.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured status=%d body=%s", resp.Code, resp.Body.String())
	}

	configured := NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore(), Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:stream"}, "user-1")}).Handler()
	forbiddenReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(`{"name":"nope"}`))
	forbiddenReq.Header.Set("Authorization", "Bearer ai-agent-token")
	forbiddenResp := httptest.NewRecorder()
	configured.ServeHTTP(forbiddenResp, forbiddenReq)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("forbidden patch status=%d body=%s", forbiddenResp.Code, forbiddenResp.Body.String())
	}
}

func TestHTTPAIAgentClientExternalAuthorizerRequiresWorkspaceScopedRoute(t *testing.T) {
	var calls int
	var got externalAuthorizerRequest
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode authorizer request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
			SchemaVersion: ExternalAuthorizerResponseSchemaVersion,
			Allowed:       true,
			PrincipalID:   "user-1",
		})
	}))
	defer authorizerServer.Close()
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{
		Endpoint: authorizerServer.URL,
	})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	server := NewServer(ServerConfig{
		AIAgentClient: NewDevelopmentAIAgentClientStore(),
		Authorizer:    authorizer,
	}).Handler()

	v1Req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	v1Req.Header.Set(aiAgentTokenHeader, "external-token")
	v1Resp := httptest.NewRecorder()
	server.ServeHTTP(v1Resp, v1Req)
	if v1Resp.Code != http.StatusForbidden {
		t.Fatalf("v1 status=%d body=%s", v1Resp.Code, v1Resp.Body.String())
	}
	if calls != 0 {
		t.Fatalf("workspace-less v1 request reached authorizer %d times", calls)
	}

	v2Req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-1/ai-agent/bootstrap", nil)
	v2Req.Header.Set(aiAgentTokenHeader, "external-token")
	v2Resp := httptest.NewRecorder()
	server.ServeHTTP(v2Resp, v2Req)
	if v2Resp.Code != http.StatusOK {
		t.Fatalf("v2 status=%d body=%s", v2Resp.Code, v2Resp.Body.String())
	}
	if calls != 1 {
		t.Fatalf("authorizer calls=%d, want 1", calls)
	}
	if got.Request.WorkspaceID != "workspace-1" || got.Request.Resource != AuthorizationResourceAIAgentClient {
		t.Fatalf("authorizer request = %+v", got)
	}
}

func newAIAgentClientHTTPTestServer(t *testing.T, credentials []StaticTokenCredential) http.Handler {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer(credentials)
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() {
		assignmentStore.Close()
	})
	profileThumbnailCredentials, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	profileThumbnails, err := NewS3AIAgentProfileThumbnailUploadService(S3AIAgentProfileThumbnailUploadConfig{
		Region:              "ap-northeast-2",
		Bucket:              "profile-upload-test",
		CDNBaseURL:          "https://cdn.example.test",
		CredentialsProvider: profileThumbnailCredentials,
		Now:                 func() time.Time { return time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewS3AIAgentProfileThumbnailUploadService: %v", err)
	}
	return NewServer(ServerConfig{
		AIAgentClient:            aiAgentStore,
		AIAgentProfileThumbnails: profileThumbnails,
		Assignment:               assignmentStore,
		TaskContext:              &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:               authorizer,
	}).Handler()
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

func findRuntime(runtimes []RuntimeRecord, id string) (RuntimeRecord, bool) {
	for _, runtime := range runtimes {
		if runtime.RuntimeID == id {
			return runtime, true
		}
	}
	return RuntimeRecord{}, false
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
	return slices.Contains(values, want)
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
