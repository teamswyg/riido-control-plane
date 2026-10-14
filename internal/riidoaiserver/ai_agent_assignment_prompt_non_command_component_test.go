package riidoaiserver

import "testing"

func TestAssignmentPromptAsksForIntentOnNonCommandComponentDocument(t *testing.T) {
	t.Parallel()
	for _, componentType := range []string{"project", "milestone", "task", "subtask"} {
		prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
			TaskID: "non-command-" + componentType,
			Context: AIAgentTaskContext{
				Component: AIAgentTaskContextComponent{
					ID:            "non-command-" + componentType,
					ComponentType: componentType,
					Title:         "1.23 신기능 출시 자료",
				},
				Document: AIAgentTaskContextDocument{
					Content:       "신기능 포지셔닝과 사용자 반응을 담은 배경 문서입니다.",
					ContentFormat: "html",
				},
			},
		})
		if err != nil {
			t.Fatalf("compose %s prompt: %v", componentType, err)
		}
		assertPromptHasAll(t, prompt.Prompt, []string{
			"- intent_class: intent_oriented",
			"- intent_gate_required: true",
			"- first_response_policy: ask_for_intent_before_deliverables_do_not_create_deliverables_until_user_replies",
		})
	}
}

func TestAssignmentPromptKeepsExplicitTitleWithPreferenceDocumentExecutable(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-bye-world",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-bye-world",
				ComponentType: "task",
				Title:         "가벼운 golang 테스트 코드 작성",
			},
			Document: AIAgentTaskContextDocument{
				Content:       "헬로 월드 말고, 바이 월드가 좋겠어",
				ContentFormat: "html",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	assertPromptHasAll(t, prompt.Prompt, []string{
		"- intent_class: explicit_instruction",
		"- intent_gate_required: false",
	})
}
