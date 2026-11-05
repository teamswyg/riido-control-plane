package riidoaiserver

import "testing"

func TestLegacyCommentConversationAdoptionPersistsWithoutChangingExecution(t *testing.T) {
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	first, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "legacy-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "legacy-run", SourceCommentID: "old-reply"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, first.TaskID, AssignAIAgentTaskRequest{AgentID: first.AgentID, AssignmentID: "next-run", conversationID: "comment:root", parentThreadID: first.ThreadID, legacyCommentConversations: []string{first.ThreadID}})
	if err != nil {
		t.Fatal(err)
	}
	reopened := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	result, err := reopened.ListAIAgentTaskThreads(t.Context(), principal, first.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Threads) != 2 {
		t.Fatal("lost original execution")
	}
	for _, thread := range result.Threads {
		if thread.ConversationID != "comment:root" {
			t.Fatal("left a separate legacy card")
		}
		if thread.ThreadID == first.ThreadID && (thread.AssignmentID != first.AssignmentID || thread.SourceCommentID != "old-reply") {
			t.Fatal("changed original execution identity")
		}
	}
}
