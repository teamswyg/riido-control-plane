package riidoaiserver

import (
	"strings"
	"testing"
)

func assertRootIntentQuestion(t *testing.T, thread AIAgentTaskThreadHistoryRecord) {
	t.Helper()
	if thread.WorkStatus != AgentWorkStatusWaitingForUser || !historyAgentMessageContains(thread, "어떤 작업부터") {
		t.Fatalf("root thread did not ask for intent: %+v", thread)
	}
	if !historyMessagesContainProgressBody(thread.Messages, "본문을 먼저 읽겠습니다.") ||
		!historyMessagesContainProgressBody(thread.Messages, "파악 완료.") {
		t.Fatalf("root progress missing: %+v", thread.Messages)
	}
}

func assertFollowupResearchLimit(t *testing.T, thread AIAgentTaskThreadHistoryRecord) {
	t.Helper()
	userText := "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해."
	if !historyMessagesContainUserBody(thread.Messages, userText) {
		t.Fatalf("followup user message missing: %+v", thread.Messages)
	}
	if !historyMessagesContainProgressBody(thread.Messages, "Surfing Google...") {
		t.Fatalf("research progress missing: %+v", thread.Messages)
	}
	if !historyAgentResultContains(thread, clientMessageCloudCreditInsufficient) ||
		historyAgentResultContains(thread, "토큰 이용 한도 초과") {
		t.Fatalf("provider limit was not normalized: %+v", thread.Messages)
	}
	if !historyMessageOrder(thread.Messages, AIAgentTaskThreadMessageRoleUser, AIAgentTaskThreadMessageRoleProgress, AIAgentTaskThreadMessageRoleAgent) {
		t.Fatalf("followup timeline order is unstable: %+v", thread.Messages)
	}
}

func historyMessagesContainProgressBody(messages []AIAgentTaskThreadHistoryMessage, body string) bool {
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleProgress && strings.Contains(message.Body, body) {
			return true
		}
	}
	return false
}

func historyMessageOrder(messages []AIAgentTaskThreadHistoryMessage, roles ...AIAgentTaskThreadMessageRole) bool {
	index := -1
	for _, role := range roles {
		next := indexOfRoleAfter(messages, role, index)
		if next <= index {
			return false
		}
		index = next
	}
	return true
}

func indexOfRoleAfter(messages []AIAgentTaskThreadHistoryMessage, role AIAgentTaskThreadMessageRole, after int) int {
	for i := after + 1; i < len(messages); i++ {
		if messages[i].Role == role {
			return i
		}
	}
	return -1
}
