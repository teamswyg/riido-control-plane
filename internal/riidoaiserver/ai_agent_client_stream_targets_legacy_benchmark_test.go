package riidoaiserver

import "testing"

func BenchmarkLegacyAIAgentTaskThreadStreamSubscriptionTargets(b *testing.B) {
	principal := AuthorizationResult{
		PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	store := inactiveHeavyStreamTargetFixtureStore(50, 2, 200)
	for range b.N {
		store.mu.Lock()
		_ = legacyActiveTaskThreadStreamTargetsLocked(store, principal, "task-load-read")
		store.mu.Unlock()
	}
}

func legacyActiveTaskThreadStreamTargetsLocked(
	s *DevelopmentAIAgentClientStore,
	principal AuthorizationResult,
	taskID string,
) []AIAgentTaskThreadStreamTarget {
	source := s.taskThreads[taskID]
	targets := make([]AIAgentTaskThreadStreamTarget, 0, len(source))
	for i := range source {
		s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
		if !taskThreadHasActiveStream(source[i]) || !s.taskThreadVisibleTo(principal, source[i]) {
			continue
		}
		targets = append(targets, AIAgentTaskThreadStreamTarget{
			AgentID: source[i].AgentID, ThreadID: source[i].ThreadID, RunID: source[i].RunID,
		})
	}
	s.taskThreads[taskID] = source
	return targets
}
