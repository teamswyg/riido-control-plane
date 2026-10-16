package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptExecutesMarketingOutcomeInSameTurn(t *testing.T) {
	t.Parallel()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-marketing-copy",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-marketing-copy",
				ComponentType: "task",
				Title:         "[1.23 신기능 마케팅] 카피라이트 3개안 준비",
			},
			Document: AIAgentTaskContextDocument{
				Content:       "신기능의 셀링 포인트를 정리하고 마케팅 방향을 분석한다.",
				ContentFormat: "html",
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	for _, want := range []string{
		"## Interaction Policy",
		"## Task Interpretation",
		"- intent_class: intent_oriented",
		"- intent_gate_required: false",
		"- first_response_policy: act_on_requested_outcome_ask_only_for_specific_blocking_information",
		"- clarification_policy: identify_the_specific_blocker_and_ask_one_focused_question_only_if_needed",
		"Intent labels are heuristic metadata only; do not let missing keywords or task type block execution.",
		"start the supported work in this same turn instead of asking the user to repeat an already specified outcome.",
		"Use the user's apparent language and product tone when asking clarification questions or reporting provider limits.",
		"When a follow-up thread message supplies a concrete instruction, treat that latest user message as the current directive",
		"Do not ask the user to click local CLI, desktop, or provider approval dialogs",
		"Do not ask conversational permission questions for file, shell, or tool execution",
		"Ask a conversational question only when a specific missing requirement blocks progress;",
		"- title: [1.23 신기능 마케팅] 카피라이트 3개안 준비",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
