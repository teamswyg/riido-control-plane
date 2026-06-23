package riidoaiserver

import (
	"context"
	"testing"
)

func TestActiveTaskThreadStreamTargetsMatchVisibleActiveThreads(t *testing.T) {
	store := streamTargetFixtureStore(6, 12)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-load-read")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	subscription, err := store.GetAIAgentTaskThreadStreamSubscription(context.Background(), principal, "task-load-read")
	if err != nil {
		t.Fatalf("GetAIAgentTaskThreadStreamSubscription: %v", err)
	}
	active := 0
	for _, thread := range threads.Threads {
		if taskThreadHasActiveStream(thread) {
			active++
		}
	}
	if len(subscription.ActiveThreadFilters) != active {
		t.Fatalf("filters=%+v active=%d", subscription.ActiveThreadFilters, active)
	}
}

func BenchmarkAIAgentTaskThreadStreamSubscriptionTargets(b *testing.B) {
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	b.Run("visible_threads_copy_lines", func(b *testing.B) {
		store := streamTargetFixtureStore(50, 200)
		for range b.N {
			store.mu.Lock()
			_ = store.visibleTaskThreadsLocked(principal, "task-load-read")
			store.mu.Unlock()
		}
	})
	b.Run("subscription_targets_only", func(b *testing.B) {
		store := streamTargetFixtureStore(50, 200)
		for range b.N {
			store.mu.Lock()
			_ = store.activeTaskThreadStreamTargetsLocked(principal, "task-load-read")
			store.mu.Unlock()
		}
	})
}
