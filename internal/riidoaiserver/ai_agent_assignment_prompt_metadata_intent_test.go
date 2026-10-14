package riidoaiserver

import "testing"

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
		"- intent_gate_required: true",
		"- first_response_policy: ask_for_intent_before_deliverables_do_not_create_deliverables_until_user_replies",
		"- clarification_question_example: 작업 내용을 확인했어요. 원하는 결과물이나 방향을 댓글로 알려주세요.",
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
		"- intent_gate_required: true",
		"- first_response_policy: infer_from_metadata_then_ask_when_unsure",
	})
}
