package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptActsFromMetadataWhenDocumentMissing(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-analysis",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-analysis",
				ComponentType: "task",
				Title:         "신기능 분석 방향 정리",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	for _, want := range []string{
		"- intent_class: intent_oriented",
		"- intent_gate_required: false",
		"- first_response_policy: act_on_requested_outcome_ask_only_for_specific_blocking_information",
		"- clarification_policy: identify_the_specific_blocker_and_ask_one_focused_question_only_if_needed",
		"- title: 신기능 분석 방향 정리",
		"not provided",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("metadata-only prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}

func TestAssignmentPromptExecutesExplicitMetadataTitleWhenDocumentMissing(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-fill-ai-content",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-fill-ai-content",
				ComponentType: "task",
				Title:         "AI 내용 채우기",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	for _, want := range []string{
		"- intent_class: explicit_instruction",
		"- intent_gate_required: false",
		"- first_response_policy: execute_the_explicit_instruction",
		"- title: AI 내용 채우기",
		"not provided",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("explicit metadata prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
