package riidoaiserver

import (
	"strings"
	"testing"
)

func TestComposeAssignRequestWithTaskContextResultPreservesIntentGate(t *testing.T) {
	t.Parallel()
	composed, err := composeAssignRequestWithTaskContextResult("task-1", "component-fallback", AssignRequest{
		AgentID:         "agent-a",
		RuntimeProvider: "claude",
		ComponentID:     " ",
	}, AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:            "component-context",
			ComponentType: "task",
			Title:         "1.23 신기능 마케팅 카피 준비",
		},
		Document: AIAgentTaskContextDocument{
			Content:       "신기능 셀링 포인트와 포지셔닝을 정리한 배경 문서입니다.",
			ContentFormat: "html",
		},
	})
	if err != nil {
		t.Fatalf("compose assignment request: %v", err)
	}
	if !composed.IntentGateRequired {
		t.Fatal("intent gate flag should survive request composition")
	}
	if got := composed.Request.ComponentID; got != "component-context" {
		t.Fatalf("component id = %q, want context component id", got)
	}
	assertPromptHasAll(t, composed.Request.Prompt, []string{
		"- intent_class: intent_oriented",
		"- intent_gate_required: true",
	})
}

func TestComposeAssignRequestWithTaskContextResultFallsBackComponentID(t *testing.T) {
	t.Parallel()
	composed, err := composeAssignRequestWithTaskContextResult("task-1", "component-route", AssignRequest{}, AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{Title: "Execute a tiny task"},
		Document:  AIAgentTaskContextDocument{Content: "Create a tiny result."},
	})
	if err != nil {
		t.Fatalf("compose assignment request: %v", err)
	}
	if got := composed.Request.ComponentID; got != "component-route" {
		t.Fatalf("component id = %q, want route component id", got)
	}
	if strings.TrimSpace(composed.Request.Prompt) == "" {
		t.Fatal("prompt should be composed")
	}
}

func TestComposeAssignRequestWithTaskContextResultRejectsEmptyContext(t *testing.T) {
	t.Parallel()
	_, err := composeAssignRequestWithTaskContextResult("", "", AssignRequest{}, AIAgentTaskContext{})
	if err == nil {
		t.Fatal("expected empty task context to fail closed")
	}
}
