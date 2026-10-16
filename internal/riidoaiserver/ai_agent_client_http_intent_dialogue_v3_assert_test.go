package riidoaiserver

import (
	"testing"
)

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
