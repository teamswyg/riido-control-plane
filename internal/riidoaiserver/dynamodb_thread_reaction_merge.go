package riidoaiserver

import "context"

// threadReactionMutationContextKey carries the one task thread whose reaction
// state the current save changes.
type threadReactionMutationContextKey struct{}

func withThreadReactionMutation(ctx context.Context, mutation threadReactionMutation) context.Context {
	return context.WithValue(ctx, threadReactionMutationContextKey{}, mutation)
}

// isThreadReactionMutation fences a reaction write like a scoped membership
// write: any part another replica committed after the reload rejects the save,
// so the stale lifecycle parts of this instance never overwrite it.
func isThreadReactionMutation(ctx context.Context) bool {
	_, ok := ctx.Value(threadReactionMutationContextKey{}).(threadReactionMutation)
	return ok
}

func mergeCommittedSharedSnapshotParts(ctx context.Context, items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	if err := mergeCommittedDeviceConnectionGrants(items, requested); err != nil {
		return err
	}
	return mergeCommittedThreadReactions(ctx, items, requested)
}

// mergeCommittedThreadReactions keeps the committed reactions so a writer with
// a stale cache never erases them. Only a reaction write changes its own
// thread, and it applies its own change on top of the committed state, so two
// replicas that wrote the same thread at once both keep their change.
func mergeCommittedThreadReactions(ctx context.Context, items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	var committed AIAgentClientSnapshot
	if err := decodeCommittedAIAgentClientPart(dynamoDBAIAgentClientSnapshotPartTaskThreadReactions, current, parts[dynamoDBAIAgentClientSnapshotPartTaskThreadReactions], &committed); err != nil {
		return err
	}
	merged := make(map[string]AIAgentTaskThreadReactionState, len(committed.TaskThreadReactions)+1)
	for threadID, state := range committed.TaskThreadReactions {
		merged[threadID] = state
	}
	if mutation, ok := ctx.Value(threadReactionMutationContextKey{}).(threadReactionMutation); ok && mutation.threadID != "" {
		state := mergedThreadReactionState(merged[mutation.threadID], mutation.before, requested.TaskThreadReactions[mutation.threadID])
		if state.isEmpty() {
			delete(merged, mutation.threadID)
		} else {
			merged[mutation.threadID] = state
		}
	}
	requested.TaskThreadReactions = retainedTaskThreadReactions(merged, requested.TaskThreads)
	return nil
}
