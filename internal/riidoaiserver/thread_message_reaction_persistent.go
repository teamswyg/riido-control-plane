package riidoaiserver

import (
	"context"
	"errors"
)

func (s *PersistentAIAgentClientStore) MutateTaskMessageReaction(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, remove bool) (AIAgentTaskThreadMessageReactionsResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	var response AIAgentTaskThreadMessageReactionsResponse
	var err error
	for attempt := 0; attempt < threadSnapshotSaveAttempts; attempt++ {
		response, err = s.mutateTaskMessageReactionAttempt(ctx, principal, taskID, threadID, messageID, emoji, remove)
		if !errors.Is(err, errTaskAgentSnapshotChanged) && !errors.Is(err, errAIAgentSnapshotConflict) {
			return response, err
		}
	}
	return response, err
}

// Bound reply and reaction retries when another replica commits after reload.
const threadSnapshotSaveAttempts = 3

// A reaction is applied to freshly reloaded state and saved fenced against
// every part committed since the reload, so it never overwrites another
// replica's messages. A rejected save reloads and applies the reaction again.
func (s *PersistentAIAgentClientStore) mutateTaskMessageReactionAttempt(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, remove bool) (AIAgentTaskThreadMessageReactionsResponse, error) {
	s.requireSnapshotReload()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadMessageReactionsResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.MutateTaskMessageReaction(ctx, principal, taskID, threadID, messageID, emoji, remove)
	if err != nil {
		return response, err
	}
	if err := s.saveSnapshot(withThreadReactionMutation(ctx)); err != nil {
		s.requireSnapshotReload()
		return response, err
	}
	return response, nil
}

func (s *PersistentAIAgentClientStore) requireSnapshotReload() {
	if s == nil || s.snapshotStore == nil {
		return
	}
	s.snapshotReloadMu.Lock()
	s.snapshotReloadRequired = true
	s.snapshotReloadMu.Unlock()
}

// threadReactionMutationContextKey marks a save that only adds a reaction.
type threadReactionMutationContextKey struct{}

func withThreadReactionMutation(ctx context.Context) context.Context {
	return context.WithValue(ctx, threadReactionMutationContextKey{}, true)
}

// isThreadReactionMutation fences a reaction write like a scoped membership
// write: any part another replica committed after the reload rejects the save,
// so the stale parts of this instance never overwrite it.
func isThreadReactionMutation(ctx context.Context) bool {
	marked, _ := ctx.Value(threadReactionMutationContextKey{}).(bool)
	return marked
}
