package riidoaiserver

import (
	"context"
	"testing"
)

func BenchmarkAIAgentTaskThreadHistoryProjection(b *testing.B) {
	principal := AuthorizationResult{
		PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	b.Run("running_threads", func(b *testing.B) {
		store := streamTargetFixtureStore(24, 40)
		for b.Loop() {
			_, err := store.ListAIAgentTaskThreadHistory(
				context.Background(), principal, "task-load-read",
			)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("queued_and_running_conversation", func(b *testing.B) {
		store := queuedHistoryFixtureStore()
		for b.Loop() {
			_, err := store.ListAIAgentTaskThreadHistory(
				context.Background(), principal, "task-load-read",
			)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func queuedHistoryFixtureStore() *DevelopmentAIAgentClientStore {
	store := streamTargetFixtureStore(24, 40)
	threads := store.taskThreads["task-load-read"]
	threads[0].WorkStatus = AgentWorkStatusQueued
	threads[0].AssignmentState = AgentAssignmentStateQueued
	threads[0].CommentKind = AgentTaskCommentQueuedByBusyAgent
	threads[1].ConversationID = taskThreadConversationID(threads[0])
	store.taskThreads["task-load-read"] = threads
	return store
}
