package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptAsksForIntentOnAmbiguousMarketingDocument(t *testing.T) {
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
		"- intent_gate_required: true",
		"- first_response_policy: ask_for_intent_before_deliverables_do_not_create_deliverables_until_user_replies",
		"- clarification_question_example: 어떤 작업부터 진행할까요?",
		"Classify the task context as either an explicit instruction or background/intent before doing work.",
		"ask a concise clarification question in the existing AI Agent thread before producing deliverables.",
		"Use the user's apparent language and product tone when asking clarification questions or reporting provider limits.",
		"When a follow-up thread message supplies a concrete instruction, treat that latest user message as the current directive",
		"Do not ask the user to click local CLI, desktop, or provider approval dialogs",
		"Do not ask conversational permission questions for file, shell, or tool execution",
		"Ask a conversational question only when task intent, business requirements, or missing context are ambiguous.",
		"- title: [1.23 신기능 마케팅] 카피라이트 3개안 준비",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
