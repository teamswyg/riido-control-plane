package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientGeneratedEndpointSmokeV1(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	token := "user-token"
	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/bootstrap", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/onboarding/fixtures", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/devices", token, "", http.StatusOK)

	createThumbnailURL := "https://cdn.riido.io/dev/ai-agents/v1-generated-smoke.png"
	createDescription := "v1 generated endpoint smoke"
	createInstruction := "v1 generated endpoint smoke instruction"
	createBody := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
		Name:                "v1 smoke direct agent",
		ProfileThumbnailURL: &createThumbnailURL,
		Description:         &createDescription,
		Instruction:         &createInstruction,
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-fast"),
	})
	createdBytes := aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/agents", token, createBody, http.StatusCreated)
	var created AgentClientRecordResponse
	aiAgentSmokeDecode(t, createdBytes, &created)
	if created.Agent.AgentID == "" ||
		created.Agent.ProfileThumbnailURL != createThumbnailURL ||
		created.Agent.Description != createDescription ||
		created.Agent.Instruction != createInstruction ||
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.ModelID != "cursor-fast" ||
		created.Agent.ModelLabel != "Cursor Fast" ||
		created.Agent.Visibility != AgentVisibilityPrivate {
		t.Fatalf("created v1 agent = %+v", created.Agent)
	}

	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/agents/"+created.Agent.AgentID+"/editability", token, "", http.StatusOK)
	patchThumbnailURL := "https://cdn.riido.io/dev/ai-agents/v1-generated-smoke-patched.png"
	patchDescription := "v1 generated endpoint smoke patched"
	patchInstruction := "v1 generated endpoint smoke patched instruction"
	patchBody := aiAgentSmokeJSON(t, UpdateAgentConfigurationRequest{
		Name:                "v1 smoke patched agent",
		ProfileThumbnailURL: &patchThumbnailURL,
		Description:         &patchDescription,
		Instruction:         &patchInstruction,
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-auto"),
	})
	patchedBytes := aiAgentSmokeRequest(t, server, http.MethodPatch, "/v1/client/ai-agent/agents/"+created.Agent.AgentID, token, patchBody, http.StatusOK)
	var patched AgentClientRecordResponse
	aiAgentSmokeDecode(t, patchedBytes, &patched)
	if patched.Agent.Name != "v1 smoke patched agent" ||
		patched.Agent.ProfileThumbnailURL != patchThumbnailURL ||
		patched.Agent.Description != patchDescription ||
		patched.Agent.Instruction != patchInstruction ||
		patched.Agent.RuntimeID != "runtime-cursor-dev" ||
		patched.Agent.ModelID != "cursor-auto" ||
		patched.Agent.Visibility != AgentVisibilityPublic {
		t.Fatalf("patched v1 agent = %+v", patched.Agent)
	}
	aiAgentSmokeRequest(t, server, http.MethodDelete, "/v1/client/ai-agent/agents/"+created.Agent.AgentID, token, "", http.StatusOK)

	fixtureCreateBody := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
		Name:       "v1 smoke fixture agent",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/riido_pm/agents", token, fixtureCreateBody, http.StatusCreated)

	assignmentTaskID := "task-v1-generated-smoke"
	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/tasks/"+assignmentTaskID+"/assignable-agents", token, "", http.StatusOK)
	assignBytes := aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/tasks/"+assignmentTaskID+"/assignment", token, `{"agent_id":"agent-public-openclaw"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, assignBytes, &assigned)
	if assigned.ThreadID == "" {
		t.Fatalf("assigned v1 thread id is empty: %+v", assigned)
	}
	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/tasks/"+assignmentTaskID+"/threads", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/tasks/"+assignmentTaskID+"/threads/"+assigned.ThreadID+"/messages", token, `{"body":"v1 next instruction","source_message_id":"smoke-v1-message"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodDelete, "/v1/client/ai-agent/tasks/"+assignmentTaskID+"/assignment", token, `{"agent_id":"agent-public-openclaw","reason":"smoke unassign"}`, http.StatusAccepted)

	commentTaskID := "task-v1-generated-comment"
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/tasks/"+commentTaskID+"/comments", token, `{"agent_id":"agent-owned-claude","body":"v1 compatibility comment","source_comment_id":"smoke-v1-comment"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/tasks/"+commentTaskID+"/stop", token, `{"agent_id":"agent-owned-claude","reason":"smoke stop"}`, http.StatusAccepted)
	eventsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/events?replay=1", token, "", http.StatusOK)
	if !strings.Contains(string(eventsBytes), "event:") {
		t.Fatalf("v1 events stream did not include SSE events: %s", string(eventsBytes))
	}

	aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/restart", token, `{"reason":"smoke restart"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/stop", token, `{"reason":"smoke stop"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/agents/agent-public-openclaw/daemon/start", token, `{"reason":"smoke start"}`, http.StatusAccepted)
}

func TestHTTPAIAgentClientGeneratedEndpointSmokeV2(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	token := "user-token"
	base := "/v2/client/workspaces/workspace-dev-riid/ai-agent"
	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/bootstrap", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/onboarding/fixtures", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/devices", token, "", http.StatusOK)

	fixtureCreateBody := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
		Name:       "v2 smoke fixture agent",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/onboarding/fixtures/hongdo_frontend/agents", token, fixtureCreateBody, http.StatusCreated)

	createThumbnailURL := "https://cdn.riido.io/dev/ai-agents/v2-generated-smoke.png"
	createDescription := "v2 generated endpoint smoke"
	createInstruction := "v2 generated endpoint smoke instruction"
	createBody := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
		Name:                "v2 smoke direct agent",
		ProfileThumbnailURL: &createThumbnailURL,
		Description:         &createDescription,
		Instruction:         &createInstruction,
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-fast"),
	})
	createdBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents", token, createBody, http.StatusCreated)
	var created AgentClientRecordResponse
	aiAgentSmokeDecode(t, createdBytes, &created)
	if created.Agent.AgentID == "" ||
		created.Agent.WorkspaceID != "workspace-dev-riid" ||
		created.Agent.ProfileThumbnailURL != createThumbnailURL ||
		created.Agent.Description != createDescription ||
		created.Agent.Instruction != createInstruction ||
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.ModelID != "cursor-fast" ||
		created.Agent.ModelLabel != "Cursor Fast" ||
		created.Agent.Visibility != AgentVisibilityPrivate {
		t.Fatalf("created v2 agent = %+v", created.Agent)
	}

	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/agents/"+created.Agent.AgentID+"/editability", token, "", http.StatusOK)
	patchThumbnailURL := "https://cdn.riido.io/dev/ai-agents/v2-generated-smoke-patched.png"
	patchDescription := "v2 generated endpoint smoke patched"
	patchInstruction := "v2 generated endpoint smoke patched instruction"
	patchBody := aiAgentSmokeJSON(t, UpdateAgentConfigurationRequest{
		Name:                "v2 smoke patched agent",
		ProfileThumbnailURL: &patchThumbnailURL,
		Description:         &patchDescription,
		Instruction:         &patchInstruction,
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-cursor-dev",
		ModelID:             stringPtr("cursor-auto"),
	})
	patchedBytes := aiAgentSmokeRequest(t, server, http.MethodPatch, base+"/agents/"+created.Agent.AgentID, token, patchBody, http.StatusOK)
	var patched AgentClientRecordResponse
	aiAgentSmokeDecode(t, patchedBytes, &patched)
	if patched.Agent.Name != "v2 smoke patched agent" ||
		patched.Agent.WorkspaceID != "workspace-dev-riid" ||
		patched.Agent.ProfileThumbnailURL != patchThumbnailURL ||
		patched.Agent.Description != patchDescription ||
		patched.Agent.Instruction != patchInstruction ||
		patched.Agent.RuntimeID != "runtime-cursor-dev" ||
		patched.Agent.ModelID != "cursor-auto" ||
		patched.Agent.Visibility != AgentVisibilityPublic {
		t.Fatalf("patched v2 agent = %+v", patched.Agent)
	}
	aiAgentSmokeRequest(t, server, http.MethodDelete, base+"/agents/"+created.Agent.AgentID, token, "", http.StatusOK)

	assignmentTaskID := "task-v2-generated-smoke"
	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+assignmentTaskID+"/assignable-agents", token, "", http.StatusOK)
	assignBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+assignmentTaskID+"/assignment", token, `{"agent_id":"agent-public-openclaw"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, assignBytes, &assigned)
	if assigned.ThreadID == "" {
		t.Fatalf("assigned v2 thread id is empty: %+v", assigned)
	}
	threadsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+assignmentTaskID+"/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, threadsBytes, &threads)
	if threads.ActiveStream == nil || threads.ActiveStream.Href != base+"/events" {
		t.Fatalf("v2 active stream = %+v", threads.ActiveStream)
	}
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+assignmentTaskID+"/threads/"+assigned.ThreadID+"/messages", token, `{"body":"v2 next instruction","source_message_id":"smoke-v2-message"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodDelete, base+"/tasks/"+assignmentTaskID+"/assignment", token, `{"agent_id":"agent-public-openclaw","reason":"smoke unassign"}`, http.StatusAccepted)

	commentTaskID := "task-v2-generated-comment"
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+commentTaskID+"/comments", token, `{"agent_id":"agent-owned-claude","body":"v2 compatibility comment","source_comment_id":"smoke-v2-comment"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+commentTaskID+"/stop", token, `{"agent_id":"agent-owned-claude","reason":"smoke stop"}`, http.StatusAccepted)
	eventsBytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/events?replay=1", token, "", http.StatusOK)
	if !strings.Contains(string(eventsBytes), "event:") {
		t.Fatalf("v2 events stream did not include SSE events: %s", string(eventsBytes))
	}

	aiAgentSmokeRequest(t, server, http.MethodGet, base+"/agents/agent-owned-codex/daemon", token, "", http.StatusOK)
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents/agent-owned-codex/daemon/restart", token, `{"reason":"smoke restart"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents/agent-owned-codex/daemon/stop", token, `{"reason":"smoke stop"}`, http.StatusAccepted)
	aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents/agent-owned-codex/daemon/start", token, `{"reason":"smoke start"}`, http.StatusAccepted)
}

func TestAIAgentGeneratedEndpointSmokeMatrixMatchesOpenAPI(t *testing.T) {
	openAPI := readAIAgentGeneratedOpenAPIOperations(t)
	matrix := readAIAgentGeneratedSmokeMatrix(t)
	knownEvidenceTests := map[string]struct{}{
		"TestHTTPAIAgentClientGeneratedEndpointSmokeV1": {},
		"TestHTTPAIAgentClientGeneratedEndpointSmokeV2": {},
	}

	if matrix.SchemaVersion != "riido-ai-agent-generated-endpoint-smoke-matrix.v1" {
		t.Fatalf("unexpected matrix schema version: %s", matrix.SchemaVersion)
	}
	if len(matrix.Entries) == 0 {
		t.Fatal("smoke matrix has no entries")
	}

	matrixByGeneratedPath := map[string]aiAgentGeneratedEndpointSmokeEntry{}
	for _, entry := range matrix.Entries {
		if strings.TrimSpace(entry.GeneratedPath) == "" {
			t.Fatalf("matrix entry has empty generated_path: %+v", entry)
		}
		if _, exists := matrixByGeneratedPath[entry.GeneratedPath]; exists {
			t.Fatalf("duplicate matrix generated_path: %s", entry.GeneratedPath)
		}
		if strings.TrimSpace(entry.Method) == "" || strings.TrimSpace(entry.Path) == "" {
			t.Fatalf("matrix entry must include method/path: %+v", entry)
		}
		if len(entry.EvidenceTests) == 0 {
			t.Fatalf("matrix entry must include evidence_tests: %+v", entry)
		}
		for _, evidence := range entry.EvidenceTests {
			if _, ok := knownEvidenceTests[evidence]; !ok {
				t.Fatalf("unknown matrix evidence test %q for %s", evidence, entry.GeneratedPath)
			}
		}
		if strings.HasPrefix(entry.GeneratedPath, "v2.") && !containsString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV2") {
			t.Fatalf("v2 matrix entry must cite v2 smoke evidence: %+v", entry)
		}
		if !strings.HasPrefix(entry.GeneratedPath, "v2.") && !containsString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV1") {
			t.Fatalf("v1 matrix entry must cite v1 smoke evidence: %+v", entry)
		}
		matrixByGeneratedPath[entry.GeneratedPath] = entry
	}

	var missing []string
	for generatedPath, operation := range openAPI {
		entry, ok := matrixByGeneratedPath[generatedPath]
		if !ok {
			missing = append(missing, generatedPath)
			continue
		}
		if entry.Method != operation.Method || entry.Path != operation.Path {
			t.Fatalf("matrix drift for %s: matrix %s %s, openapi %s %s", generatedPath, entry.Method, entry.Path, operation.Method, operation.Path)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("smoke matrix missing OpenAPI generated paths: %v", missing)
	}

	var extra []string
	for generatedPath := range matrixByGeneratedPath {
		if _, ok := openAPI[generatedPath]; !ok {
			extra = append(extra, generatedPath)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		t.Fatalf("smoke matrix references unknown generated paths: %v", extra)
	}
}

func aiAgentSmokeRequest(t *testing.T, server http.Handler, method, path, token, body string, wantStatus int) []byte {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, resp.Code, wantStatus, resp.Body.String())
	}
	return resp.Body.Bytes()
}

func aiAgentSmokeJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal smoke body: %v", err)
	}
	return string(body)
}

func aiAgentSmokeDecode(t *testing.T, body []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode smoke json: %v body=%s", err, string(body))
	}
}

type aiAgentGeneratedEndpointSmokeMatrix struct {
	SchemaVersion string                               `json:"schema_version"`
	Entries       []aiAgentGeneratedEndpointSmokeEntry `json:"entries"`
}

type aiAgentGeneratedEndpointSmokeEntry struct {
	GeneratedPath string   `json:"generated_path"`
	Method        string   `json:"method"`
	Path          string   `json:"path"`
	EvidenceTests []string `json:"evidence_tests"`
}

type aiAgentGeneratedOpenAPIOperation struct {
	Method string
	Path   string
}

func readAIAgentGeneratedSmokeMatrix(t *testing.T) aiAgentGeneratedEndpointSmokeMatrix {
	t.Helper()
	matrixPath := filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.smoke-matrix.riido.json")
	data, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read smoke matrix: %v", err)
	}
	var matrix aiAgentGeneratedEndpointSmokeMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		t.Fatalf("decode smoke matrix: %v", err)
	}
	return matrix
}

func readAIAgentGeneratedOpenAPIOperations(t *testing.T) map[string]aiAgentGeneratedOpenAPIOperation {
	t.Helper()
	openAPIPath := filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json")
	data, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}
	out := map[string]aiAgentGeneratedOpenAPIOperation{}
	for path, pathItem := range doc.Paths {
		for method, raw := range pathItem {
			if !aiAgentGeneratedOpenAPIHTTPMethod(method) {
				continue
			}
			var operation struct {
				XRiidoClient struct {
					GeneratedPath string `json:"generated_path"`
				} `json:"x-riido-client"`
			}
			if err := json.Unmarshal(raw, &operation); err != nil {
				t.Fatalf("decode openapi operation %s %s: %v", method, path, err)
			}
			generatedPath := strings.TrimSpace(operation.XRiidoClient.GeneratedPath)
			if generatedPath == "" {
				continue
			}
			if _, exists := out[generatedPath]; exists {
				t.Fatalf("duplicate OpenAPI generated_path: %s", generatedPath)
			}
			out[generatedPath] = aiAgentGeneratedOpenAPIOperation{Method: strings.ToUpper(method), Path: path}
		}
	}
	if len(out) == 0 {
		t.Fatal("OpenAPI has no x-riido-client.generated_path operations")
	}
	return out
}

func aiAgentGeneratedOpenAPIHTTPMethod(value string) bool {
	switch strings.ToLower(value) {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}

func TestAIAgentGeneratedEndpointSmokeMatrixEntriesStaySorted(t *testing.T) {
	matrix := readAIAgentGeneratedSmokeMatrix(t)
	previous := ""
	for i, entry := range matrix.Entries {
		sortKey := fmt.Sprintf("%s %s", entry.GeneratedPath, entry.Method)
		if i > 0 && sortKey < previous {
			t.Fatalf("smoke matrix entries must stay sorted by generated_path: %q before %q", sortKey, previous)
		}
		previous = sortKey
	}
}
