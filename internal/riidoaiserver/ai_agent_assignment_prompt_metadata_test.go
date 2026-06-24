package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptAsksFromIntentMetadataWhenDocumentMissing(t *testing.T) {
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
		"- intent_gate_required: true",
		"- first_response_policy: ask_for_intent_before_deliverables_do_not_create_deliverables_until_user_replies",
		"- clarification_question_example: 어떤 작업부터 진행할까요?",
		"- title: 신기능 분석 방향 정리",
		"not provided",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("metadata-only prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
