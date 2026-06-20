package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
