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
		"- intent_gate_required: false",
		"- first_response_policy: act_on_requested_outcome_ask_only_for_specific_blocking_information",
		"- clarification_policy: identify_the_specific_blocker_and_ask_one_focused_question_only_if_needed",
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
		"- intent_gate_required: false",
		"- first_response_policy: act_on_requested_outcome_ask_only_for_specific_blocking_information",
	})
}
