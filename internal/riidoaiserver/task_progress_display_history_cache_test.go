package riidoaiserver

import "testing"

func TestContextualProgressStoreHistoryCacheInvalidatesChangedSummary(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	thread := AIAgentTaskThreadRecord{
		ThreadID: "thread",
		Lines: []AgentThreadProgressLine{{
			Seq: 1, Message: "same rendered body", MessageCode: 1103,
			MessageArgs: map[string]string{"description": "댓글 구조를 살펴보고 있어요."},
		}},
	}
	first := store.cachedTaskThreadHistoryMessagesLocked(thread)
	if len(first) == 0 || first[0].DisplayMessage != "댓글 구조를 살펴보고 있어요." {
		t.Fatalf("initial history summary: %+v", first)
	}
	thread.Lines[0].MessageArgs = map[string]string{"description": "답글 전송 권한을 점검하고 있어요."}
	updated := store.cachedTaskThreadHistoryMessagesLocked(thread)
	if len(updated) == 0 || updated[0].DisplayMessage != "답글 전송 권한을 점검하고 있어요." {
		t.Fatalf("store history returned stale summary: %+v", updated)
	}
}
