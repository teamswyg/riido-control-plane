package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("generated client missing %q", required)
		}
	}
}
