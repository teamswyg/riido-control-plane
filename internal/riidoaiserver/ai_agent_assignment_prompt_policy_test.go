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
		"- title: [1.23 신기능 마케팅] 카피라이트 3개안 준비",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}

func TestFollowupPromptKeepsLatestUserInstructionAuthoritative(t *testing.T) {
	t.Parallel()
	prompt := appendAIAgentTaskThreadMessagePrompt("base prompt", AIAgentTaskThreadRecord{
		ThreadID:        "thread-a",
		RunID:           "run-a",
		WorkStatus:      AgentWorkStatusFailed,
		AssignmentState: AgentAssignmentStateFailed,
		Message:         "보유하신 크레딧이 부족합니다.",
	}, CreateAIAgentTaskThreadMessageRequest{
		Body: "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.",
	})
	for _, want := range []string{
		"## Follow-up Thread Message",
		"### Previous Thread Message",
		"보유하신 크레딧이 부족합니다.",
		"### New User Instruction",
		"팩트 기반으로 작성하자",
		"### Follow-up Execution Policy",
		"The New User Instruction is authoritative for this run.",
		"Re-read the latest Task Document before answering because it may have changed after the previous run.",
		"Do not ask the user to click local CLI, desktop, or provider approval dialogs",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("follow-up prompt missing %q:\n%s", want, prompt)
		}
	}
}
