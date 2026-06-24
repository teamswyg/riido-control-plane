package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptClassifiesIntentComponentTypes(t *testing.T) {
	t.Parallel()
	for _, componentType := range []string{"project", "milestone", "task", "subtask"} {
		prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
			TaskID: "component-" + componentType,
			Context: AIAgentTaskContext{
				Component: AIAgentTaskContextComponent{
					ID:            "component-" + componentType,
					ComponentType: componentType,
					Title:         "[1.23 신기능 마케팅] 카피라이트 3개안 준비",
				},
				Document: AIAgentTaskContextDocument{
					Content:       "신기능 셀링 포인트 세 가지를 분석하고 마케팅 방향을 정리한다.",
					ContentFormat: "html",
				},
			},
		})
		if err != nil {
			t.Fatalf("compose %s prompt: %v", componentType, err)
		}
		assertPromptHasAll(t, prompt.Prompt, []string{
			"- component_type: " + componentType,
			"- intent_class: intent_oriented",
			"- first_response_policy: ask_for_intent_before_deliverables_when_first_action_is_ambiguous",
			"- clarification_question_example: 어떤 작업부터 진행할까요?",
		})
	}
}

func TestAssignmentPromptClassifiesIntentFromMetadataWhenDocumentMissing(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-marketing-copy-empty",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-marketing-copy-empty",
				ComponentType: "task",
				Title:         "[1.23 신기능 마케팅] 카피라이트 3개안 준비",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose metadata intent prompt: %v", err)
	}
	assertPromptHasAll(t, prompt.Prompt, []string{
		"- intent_class: intent_oriented",
		"- first_response_policy: ask_for_intent_before_deliverables_when_first_action_is_ambiguous",
		"- clarification_question_example: 어떤 작업부터 진행할까요?",
	})
}

func TestAssignmentPromptKeepsMetadataOnlyWhenDocumentAndIntentMarkersMissing(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-empty",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-empty",
				ComponentType: "task",
				Title:         "일반 작업",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose metadata-only prompt: %v", err)
	}
	assertPromptHasAll(t, prompt.Prompt, []string{
		"- intent_class: metadata_only",
		"- first_response_policy: infer_from_metadata_then_ask_when_unsure",
	})
}

func TestAssignmentPromptKeepsExplicitInstructionExecutable(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-go-hello",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-go-hello",
				ComponentType: "task",
				Title:         "Go 헬로월드 구현",
			},
			Document: AIAgentTaskContextDocument{
				Content: "go.mod, main.go, main_test.go를 만들고 go test ./...를 통과시켜라.",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose explicit prompt: %v", err)
	}
	assertPromptHasAll(t, prompt.Prompt, []string{
		"- intent_class: explicit_instruction",
		"- first_response_policy: execute_the_explicit_instruction",
	})
}

func assertPromptHasAll(t *testing.T, prompt string, values []string) {
	t.Helper()
	for _, want := range values {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}
