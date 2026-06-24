package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptInfersFromMetadataWhenDocumentMissing(t *testing.T) {
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
		"- intent_class: metadata_only",
		"- first_response_policy: infer_from_metadata_then_ask_when_unsure",
		"- title: 신기능 분석 방향 정리",
		"not provided",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("metadata-only prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
