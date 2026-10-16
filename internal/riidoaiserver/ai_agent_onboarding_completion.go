package riidoaiserver

import (
	"errors"
	"strings"
)

// Onboarding completion records the first onboarding task result as a completed,
// agent-authored task comment. It never creates a queued or leasable assignment.
const onboardingSeedAssignmentIDPrefix = "onboarding-seed-"

type AIAgentOnboardingTaskTemplateKey string

const (
	AIAgentOnboardingTemplateSchedule  AIAgentOnboardingTaskTemplateKey = "schedule"
	AIAgentOnboardingTemplateInterview AIAgentOnboardingTaskTemplateKey = "interview"
	AIAgentOnboardingTemplateLaunch    AIAgentOnboardingTaskTemplateKey = "launch"
)

type AIAgentOnboardingLocale string

const (
	AIAgentOnboardingLocaleKo AIAgentOnboardingLocale = "ko"
	AIAgentOnboardingLocaleEn AIAgentOnboardingLocale = "en"
)

type CompleteAIAgentOnboardingTaskRequest struct {
	AgentID     string                           `json:"agent_id"`
	TemplateKey AIAgentOnboardingTaskTemplateKey `json:"template_key"`
	RequestID   string                           `json:"request_id,omitempty"`
	Locale      AIAgentOnboardingLocale          `json:"locale,omitempty"`
}

// Approved product copy. These strings must match ai_agent_onboarding_completion.v1
// in riido-contracts verbatim.
var onboardingCompletionCopy = map[AIAgentOnboardingTaskTemplateKey]map[AIAgentOnboardingLocale]string{
	AIAgentOnboardingTemplateSchedule: {
		AIAgentOnboardingLocaleKo: "단계 4개를 이 작업의 하위 작업으로 추가하고 마감일을 넣어 뒀어요. 참여자는 비워 뒀으니 팀원을 초대하면 바로 넣을 수 있어요.",
		AIAgentOnboardingLocaleEn: "Added the 4 phases as subtasks of this task with due dates. Assignees are empty, so you can add them as soon as you invite teammates.",
	},
	AIAgentOnboardingTemplateInterview: {
		AIAgentOnboardingLocaleKo: "질문지를 이 작업의 문서로 남겼어요. \"~라면 쓰시겠어요?\"처럼 답을 유도하는 질문은 뺐어요.",
		AIAgentOnboardingLocaleEn: "Saved the questions as this task's document. Leading questions like \"Would you use it if…?\" were left out.",
	},
	AIAgentOnboardingTemplateLaunch: {
		AIAgentOnboardingLocaleKo: "항목 9개를 이 작업의 하위 작업으로 추가하고 D-3, 당일, D+7 마감일을 넣어 뒀어요.",
		AIAgentOnboardingLocaleEn: "Added the 9 items as subtasks of this task with D-3, launch-day, and D+7 due dates.",
	},
}

var errOnboardingCompletionInvalid = errors.New("onboarding completion request is invalid")

// onboardingCompletionText resolves the approved copy. An omitted locale means ko.
func onboardingCompletionText(template AIAgentOnboardingTaskTemplateKey, locale AIAgentOnboardingLocale) (string, error) {
	byLocale, ok := onboardingCompletionCopy[template]
	if !ok {
		return "", errOnboardingCompletionInvalid
	}
	if locale == "" {
		locale = AIAgentOnboardingLocaleKo
	}
	text, ok := byLocale[locale]
	if !ok {
		return "", errOnboardingCompletionInvalid
	}
	return text, nil
}

// onboardingSeedAssignmentID is deterministic per (task, agent), which makes the
// completion idempotent without a request id.
func onboardingSeedAssignmentID(taskID, agentID string) string {
	return onboardingSeedAssignmentIDPrefix + strings.TrimSpace(taskID) + "-" + strings.TrimSpace(agentID)
}

func isOnboardingSeedAssignmentID(assignmentID string) bool {
	return strings.HasPrefix(strings.TrimSpace(assignmentID), onboardingSeedAssignmentIDPrefix)
}
