package riidoaiserver

import (
	"strings"
	"testing"
)

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
		"The New User Instruction is authoritative for this run, including over earlier task intent gates and requests to modify files.",
		"Use the supplied task document and recent conversation context. Fetch more only when needed context is missing or its stated freshness is insufficient for the current request.",
		"Do not ask the user to click local CLI, desktop, or provider approval dialogs",
		"Do not ask conversational permission questions for file, shell, or tool execution",
		"Ask a conversational question only when task intent, business requirements, or missing context are ambiguous.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("follow-up prompt missing %q:\n%s", want, prompt)
		}
	}
}
