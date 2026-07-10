package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
	if got, want := aiAgentIDs(assignable.Agents), []string{"agent-owned-codex", "agent-public-openclaw"}; !sameStrings(got, want) {
		t.Fatalf("assignable agents = %v, want %v", got, want)
	}
	if !assignable.Agents[0].IsOwnedByViewer || assignable.Agents[1].IsOwnedByViewer {
		t.Fatalf("assignable owned-first flags = %+v", assignable.Agents)
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
