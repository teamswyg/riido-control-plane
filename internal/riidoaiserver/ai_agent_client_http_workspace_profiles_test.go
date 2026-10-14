package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

	assertWorkspaceAssignedProfileColor(t, server, "23958923859", "#B87EAD")
}
