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
	wantReactPath := filepath.Join("..", "..", "web", "generated", "aiAgentClient.react.ts")
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
	gotReact, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	wantReact, err := os.ReadFile(wantReactPath)
	if err != nil {
		t.Fatalf("read generated react file: %v", err)
	}
	if !bytes.Equal(gotReact, wantReact) {
		t.Fatalf("generated React Query hook wrapper drifted; run go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts")
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
		"export type RiidoQueryOptions",
		"export type RiidoMutationOptions",
		"/v1/client/ai-agent/tasks/${params.task_id}/assignable-agents",
		"export function listAIAgentTaskThreadsQueryOptions",
		"/v1/client/ai-agent/tasks/${params.task_id}/threads",
		"export interface AIAgentTaskThreadCollectionResponse",
		"export function submitAIAgentTaskCommentMutationOptions",
		"export function stopAIAgentTaskMutationOptions",
		"Promise<AIAgentTaskActionResponse>",
		"export async function streamAIAgentClientEvents",
		"Promise<Response>",
		"이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.",
		"AI Agent 화면 진입 시 필요한 agent와 device runtime 초기 데이터입니다.",
		"web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다",
		"cache tag: `aiAgent.tasks.assignableAgents`",
		"export function getAIAgentClientBootstrapQueryOptions",
		"export function listAIAgentTaskAssignableAgentsQueryKeyRoot",
		"export interface StopAIAgentTaskMutationVariables",
		"export function stopAIAgentTaskMutationOptions",
		"export interface RiidoControlPlaneClient",
		"export interface RiidoAIAgentDevicesNamespace",
		"export function createRiidoControlPlaneClient",
		"readonly queryKeyRoot",
		"readonly invalidateAll",
		"readonly prefetch",
		"invalidates: {",
		"import type { QueryClient, UseMutationOptions, UseQueryOptions } from '@/lib/react-query';",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("generated client missing %q", required)
		}
	}
	if strings.Contains(body, "@tanstack/react-query") {
		t.Fatalf("core generated client must import React Query types through '@/lib/react-query'")
	}

	gotReact, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	reactBody := string(gotReact)
	for _, required := range []string{
		"'use client';",
		"export function useRiidoControlPlaneClient",
		"readonly useQuery",
		"readonly useMutation",
		"readonly threads",
		"from '@/lib/react-query'",
		"core.RiidoQueryOptions<Response>",
		"hook은 반드시 `@/lib/react-query`를 통과하므로",
	} {
		if !strings.Contains(reactBody, required) {
			t.Fatalf("generated react wrapper missing %q", required)
		}
	}
	if strings.Contains(reactBody, "core.Response") {
		t.Fatalf("generated react wrapper must use global Response, not core.Response")
	}
	if strings.Contains(reactBody, "@tanstack/react-query") {
		t.Fatalf("react generated client must import hooks through '@/lib/react-query'")
	}
}

func TestAIAgentClientMetadataFlowsThroughContractProjections(t *testing.T) {
	base := filepath.Join("..", "..", "contracts", "ai-agent-client")
	dsl := loadContractClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.dsl.riido.json"))
	ir := loadContractClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.ir.riido.json"))
	openapi := loadOpenAPIClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.openapi.json"))

	if ir.modules != dsl.modules {
		t.Fatalf("IR client modules = %s, want %s", ir.modules, dsl.modules)
	}
	if openapi.modules != dsl.modules {
		t.Fatalf("OpenAPI client modules = %s, want %s", openapi.modules, dsl.modules)
	}

	for operationID, want := range dsl.operations {
		if got, ok := ir.operations[operationID]; !ok || got != want {
			t.Fatalf("IR client metadata for %s = %q, want %q", operationID, got, want)
		}
		if got, ok := openapi.operations[operationID]; !ok || got != want {
			t.Fatalf("OpenAPI client metadata for %s = %q, want %q", operationID, got, want)
		}
	}
	if len(ir.operations) != len(dsl.operations) {
		t.Fatalf("IR metadata count = %d, want %d", len(ir.operations), len(dsl.operations))
	}
	if len(openapi.operations) != len(dsl.operations) {
		t.Fatalf("OpenAPI metadata count = %d, want %d", len(openapi.operations), len(dsl.operations))
	}
	for operationID, required := range map[string][]string{
		"getAIAgentClientBootstrap":       {"cache_tag"},
		"listAIAgentTaskAssignableAgents": {"cache_tag"},
		"deleteAIAgent":                   {"invalidates"},
		"submitAIAgentTaskComment":        {"invalidates"},
		"stopAIAgentTask":                 {"invalidates"},
	} {
		metadata := dsl.operations[operationID]
		for _, field := range required {
			if !strings.Contains(metadata, field) {
				t.Fatalf("%s metadata missing %s in %s", operationID, field, metadata)
			}
		}
	}
}

type clientProjection struct {
	modules    string
	operations map[string]string
}

func loadContractClientProjection(t *testing.T, path string) clientProjection {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc struct {
		ClientModules []clientModuleMetadata  `json:"client_modules"`
		Operations    []testContractOperation `json:"operations"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return clientProjection{
		modules:    canonicalJSON(t, doc.ClientModules),
		operations: clientMetadataByOperation(t, path, doc.Operations),
	}
}

func loadOpenAPIClientProjection(t *testing.T, path string) clientProjection {
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
	return clientProjection{
		modules:    canonicalJSON(t, spec.ClientModules),
		operations: clientMetadataByOperation(t, path, operations),
	}
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
		out[operationID] = canonicalJSON(t, operation.Client)
	}
	return out
}

func canonicalJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return string(data)
}
