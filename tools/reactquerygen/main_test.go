package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testContractOperation struct {
	OperationID string         `json:"operation_id"`
	Client      clientMetadata `json:"client"`
}

func TestGenerateReactQueryClientDoesNotDrift(t *testing.T) {
	openAPIPath := filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json")
	wantPath := filepath.Join("..", "..", "web", "generated", "aiAgentClient.ts")
	spec, err := loadOpenAPI(openAPIPath)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	got, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated React Query client drifted; run go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts")
	}
}

func TestGenerateReactQueryClientIncludesAIAgentSurface(t *testing.T) {
	openAPIPath := filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json")
	spec, err := loadOpenAPI(openAPIPath)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	got, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	body := string(got)
	for _, required := range []string{
		"export type AgentTaskCommentKind",
		"export function useListAIAgentTaskAssignableAgents",
		"/v1/client/ai-agent/tasks/${params.task_id}/assignable-agents",
		"export function useSubmitAIAgentTaskComment",
		"export function useStopAIAgentTask",
		"Promise<AIAgentTaskActionResponse>",
		"export async function streamAIAgentClientEvents",
		"Promise<Response>",
		"이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.",
		"AI Agent 화면 진입 시 필요한 agent와 device runtime 초기 데이터입니다.",
		"web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다",
		"React Query query hook입니다.",
		"export function getAIAgentClientBootstrapQueryOptions",
		"export interface StopAIAgentTaskMutationVariables",
		"export function stopAIAgentTaskMutationOptions",
		"export function createRiidoControlPlaneClient",
		"aiAgent: {",
		"query: (",
		"mutation: (",
		"tasks: {",
		"assignableAgents: {",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("generated client missing %q", required)
		}
	}
}

func TestAIAgentClientMetadataFlowsThroughContractProjections(t *testing.T) {
	base := filepath.Join("..", "..", "contracts", "ai-agent-client")
	dsl := loadContractClientMetadata(t, filepath.Join(base, "control-plane-ai-agent-client.dsl.riido.json"))
	ir := loadContractClientMetadata(t, filepath.Join(base, "control-plane-ai-agent-client.ir.riido.json"))
	openapi := loadOpenAPIClientMetadata(t, filepath.Join(base, "control-plane-ai-agent-client.openapi.json"))

	for operationID, want := range dsl {
		if got, ok := ir[operationID]; !ok || got != want {
			t.Fatalf("IR client metadata for %s = %q, want %q", operationID, got, want)
		}
		if got, ok := openapi[operationID]; !ok || got != want {
			t.Fatalf("OpenAPI client metadata for %s = %q, want %q", operationID, got, want)
		}
	}
	if len(ir) != len(dsl) {
		t.Fatalf("IR metadata count = %d, want %d", len(ir), len(dsl))
	}
	if len(openapi) != len(dsl) {
		t.Fatalf("OpenAPI metadata count = %d, want %d", len(openapi), len(dsl))
	}
}

func loadContractClientMetadata(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc struct {
		Operations []testContractOperation `json:"operations"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return clientMetadataByOperation(t, path, doc.Operations)
}

func loadOpenAPIClientMetadata(t *testing.T, path string) map[string]string {
	t.Helper()
	spec, err := loadOpenAPI(path)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	var operations []testContractOperation
	for _, byMethod := range spec.Paths {
		for _, operation := range byMethod {
			operations = append(operations, testContractOperation{
				OperationID: operation.OperationID,
				Client:      operation.Client,
			})
		}
	}
	return clientMetadataByOperation(t, path, operations)
}

func clientMetadataByOperation(t *testing.T, path string, operations []testContractOperation) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, operation := range operations {
		operationID := strings.TrimSpace(operation.OperationID)
		if operationID == "" {
			t.Fatalf("%s has operation without id", path)
		}
		if strings.TrimSpace(operation.Client.Module) == "" {
			t.Fatalf("%s operation %s missing client.module", path, operationID)
		}
		if len(operation.Client.FacadePath) == 0 {
			t.Fatalf("%s operation %s missing client.facade_path", path, operationID)
		}
		out[operationID] = operation.Client.Module + "." + strings.Join(operation.Client.FacadePath, ".")
	}
	return out
}
