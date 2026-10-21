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
		AIAgentOnboardingLocaleKo: "단계 4개의 기간과 날짜, 산출물을 이 작업의 문서로 정리해 뒀어요. 팀원을 초대하면 단계마다 담당자를 정해 보세요.",
		AIAgentOnboardingLocaleEn: "Laid out the 4 phases' durations, dates, and deliverables in this task's document. Once you invite teammates, pick an owner for each phase.",
	},
	AIAgentOnboardingTemplateInterview: {
		AIAgentOnboardingLocaleKo: "질문지를 이 작업의 문서로 남겼어요. \"~라면 쓰시겠어요?\"처럼 답을 유도하는 질문은 뺐어요.",
		AIAgentOnboardingLocaleEn: "Saved the questions as this task's document. Leading questions like \"Would you use it if…?\" were left out.",
	},
	AIAgentOnboardingTemplateLaunch: {
		AIAgentOnboardingLocaleKo: "출시 전·당일·후에 확인할 항목 9개를 이 작업의 문서로 정리해 뒀어요. 출시일은 오늘부터 7일 뒤로 잡았어요.",
		AIAgentOnboardingLocaleEn: "Laid out 9 items to check before, on, and after launch in this task's document. The launch date is set 7 days from today.",
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
