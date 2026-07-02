package riidoaiserver

import (
	"net/http"
	"testing"
)

func getIntentDialogueV3Thread(
	t *testing.T,
	handler http.Handler,
	action AIAgentTaskActionResponse,
	threadID string,
) AIAgentTaskThreadHistoryRecord {
	t.Helper()
	return historyThreadByID(t, getIntentDialogueV3History(t, handler, action.TaskID), threadID)
}

func assertHTTPRootIntentQuestion(t *testing.T, thread AIAgentTaskThreadHistoryRecord) {
	t.Helper()
	if thread.WorkStatus != AgentWorkStatusWaitingForUser ||
		thread.AssignmentState != AgentAssignmentStateWaiting ||
		thread.ActiveStream != nil ||
		!historyAgentMessageContains(thread, "원하는 결과물이나 방향") {
		t.Fatalf("root thread did not ask for intent: %+v", thread)
	}
	if historyMessagesContainProgressBody(thread.Messages, "본문을 먼저 읽겠습니다.") {
		t.Fatalf("intent-gated root thread reached runtime before user intent: %+v", thread.Messages)
	}
}

func assertHTTPDraftReadsBodyBeforeCompletion(t *testing.T, thread AIAgentTaskThreadHistoryRecord) {
	t.Helper()
	if !historyMessagesContainProgressBody(thread.Messages, "본문을 먼저 읽겠습니다.") ||
		!historyMessagesContainProgressBody(thread.Messages, "파악 완료.") {
		t.Fatalf("draft thread did not preserve body-reading progress: %+v", thread.Messages)
	}
}

func assertPostLimitHandoffKeepsPriorConversation(
	t *testing.T,
	history AIAgentTaskThreadHistoryCollectionResponse,
	limitedThreadID string,
	handoffThreadID string,
) {
	t.Helper()
	limited := historyThreadByID(t, history, limitedThreadID)
	handoff := historyThreadByID(t, history, handoffThreadID)
	if !historyAgentResultContains(limited, clientMessageCloudCreditInsufficient) {
		t.Fatalf("provider limit result disappeared after handoff: %+v", limited)
	}
	if handoff.ConversationID == "" || handoff.ConversationID == limited.ConversationID {
		t.Fatalf("post-limit handoff reused limited conversation: limited=%+v handoff=%+v", limited, handoff)
	}
}
