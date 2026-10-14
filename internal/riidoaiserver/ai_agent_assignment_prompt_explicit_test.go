package riidoaiserver

import "testing"

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
		"- intent_gate_required: false",
		"- first_response_policy: execute_the_explicit_instruction",
	})
}
