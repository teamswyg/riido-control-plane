package riidoaiserver

import "testing"

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
			"- intent_gate_required: false",
			"- first_response_policy: act_on_requested_outcome_ask_only_for_specific_blocking_information",
			"- clarification_policy: identify_the_specific_blocker_and_ask_one_focused_question_only_if_needed",
		})
	}
}
