package riidoaiserver

import "testing"

func assertConversationFollowup(t *testing.T, history AIAgentTaskThreadHistoryCollectionResponse, rootThreadID, followupThreadID string) {
	t.Helper()
	root := historyThreadByID(t, history, rootThreadID)
	followup := historyThreadByID(t, history, followupThreadID)
	if root.ConversationID == "" {
		t.Fatalf("root conversation_id is empty: %+v", root)
	}
	if followup.ConversationID != root.ConversationID {
		t.Fatalf("followup conversation mismatch: root=%+v followup=%+v", root, followup)
	}
	if followup.ParentThreadID != root.ThreadID {
		t.Fatalf("followup parent_thread_id mismatch: root=%+v followup=%+v", root, followup)
	}
}

func assertConversationReassignment(t *testing.T, history AIAgentTaskThreadHistoryCollectionResponse, rootThreadID, reassignedThreadID string) {
	t.Helper()
	root := historyThreadByID(t, history, rootThreadID)
	reassigned := historyThreadByID(t, history, reassignedThreadID)
	if reassigned.ConversationID == "" || reassigned.ConversationID == root.ConversationID {
		t.Fatalf("reassignment conversation collapsed: root=%+v reassigned=%+v", root, reassigned)
	}
	if reassigned.ParentThreadID != "" {
		t.Fatalf("reassignment must not have parent_thread_id: %+v", reassigned)
	}
}
