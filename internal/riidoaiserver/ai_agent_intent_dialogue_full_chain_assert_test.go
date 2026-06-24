package riidoaiserver

import "testing"

func assertNestedConversationFollowup(t *testing.T, history AIAgentTaskThreadHistoryCollectionResponse, rootID, parentID, childID string) {
	t.Helper()
	root := historyThreadByID(t, history, rootID)
	parent := historyThreadByID(t, history, parentID)
	child := historyThreadByID(t, history, childID)
	if parent.ConversationID != root.ConversationID || child.ConversationID != root.ConversationID {
		t.Fatalf("nested conversation collapsed: root=%+v parent=%+v child=%+v", root, parent, child)
	}
	if child.ParentThreadID != parent.ThreadID {
		t.Fatalf("nested parent_thread_id mismatch: parent=%+v child=%+v", parent, child)
	}
}

func assertDraftCopywritingComplete(t *testing.T, thread AIAgentTaskThreadHistoryRecord) {
	t.Helper()
	userText := "우리 신기능 셀링 포인트 세 가지 반영해서 훅이 강한 카피라이팅 초안 3개만 짜줘."
	if !historyMessagesContainUserBody(thread.Messages, userText) {
		t.Fatalf("draft user request missing: %+v", thread.Messages)
	}
	if thread.AssignmentState != AgentAssignmentStateCompleted ||
		!historyAgentResultContains(thread, "Riido 카피라이팅 3개 대안 비교") {
		t.Fatalf("draft result missing: %+v", thread)
	}
}
