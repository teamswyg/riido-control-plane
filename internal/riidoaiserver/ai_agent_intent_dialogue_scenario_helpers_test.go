package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func assertMarketingPromptAsksForIntent(t *testing.T) {
	t.Helper()
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-copywriting", Context: AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task-copywriting", Title: "[1.23 신기능 마케팅] 카피라이트 3개안 준비"},
		Document:  AIAgentTaskContextDocument{Content: "신기능 셀링 포인트 세 가지를 정리하고 분석한다.", ContentFormat: "html"},
	}})
	if err != nil {
		t.Fatalf("ComposeAIAgentAssignmentPrompt: %v", err)
	}
	if !strings.Contains(prompt.Prompt, "intent_class: intent_oriented") || !strings.Contains(prompt.Prompt, "원하는 결과물이나 방향") {
		t.Fatalf("marketing prompt did not ask for intent:\n%s", prompt.Prompt)
	}
}

func recordNeedsInput(t *testing.T, store *DevelopmentAIAgentClientStore, root AIAgentTaskActionResponse, message string) {
	t.Helper()
	event := TaskEvent{TaskID: root.TaskID, AssignmentID: root.AssignmentID, AgentID: root.AgentID, Type: EventAssignmentStateUpdated, State: AssignmentRunning, Message: message, Metadata: map[string]string{metadatakeys.AssignmentResultStatus.String(): "needs_input"}, At: time.Now().UTC()}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), root.AgentID, AgentEventRequest{}, event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent needs_input: %v", err)
	}
}

func recordProviderLimit(t *testing.T, store *DevelopmentAIAgentClientStore, followup AIAgentTaskActionResponse) {
	t.Helper()
	event := TaskEvent{TaskID: followup.TaskID, AssignmentID: followup.AssignmentID, AgentID: followup.AgentID, Type: EventAssignmentFailed, State: AssignmentFailed, Message: "Token quota exceeded while researching", At: time.Now().UTC()}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), followup.AgentID, AgentEventRequest{}, event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent provider limit: %v", err)
	}
}

func historyAgentMessageContains(thread AIAgentTaskThreadHistoryRecord, text string) bool {
	return historyAgentFieldContains(thread, text, func(m AIAgentTaskThreadHistoryMessage) string { return m.Body })
}

func historyAgentResultContains(thread AIAgentTaskThreadHistoryRecord, text string) bool {
	return historyAgentFieldContains(thread, text, func(m AIAgentTaskThreadHistoryMessage) string { return m.ResultMessage })
}

func historyAgentFieldContains(thread AIAgentTaskThreadHistoryRecord, text string, field func(AIAgentTaskThreadHistoryMessage) string) bool {
	for _, message := range thread.Messages {
		if message.Role == AIAgentTaskThreadMessageRoleAgent && strings.Contains(field(message), text) {
			return true
		}
	}
	return false
}
