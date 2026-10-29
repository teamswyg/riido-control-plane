package riidoaiserver

import "context"

// threadReactionMutationContextKey names the one task thread whose reaction
// state the current save changes.
type threadReactionMutationContextKey struct{}

func withThreadReactionMutation(ctx context.Context, threadID string) context.Context {
	return context.WithValue(ctx, threadReactionMutationContextKey{}, threadID)
}

func mergeCommittedSharedSnapshotParts(ctx context.Context, items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	if err := mergeCommittedDeviceConnectionGrants(items, requested); err != nil {
		return err
	}
	return mergeCommittedThreadReactions(ctx, items, requested)
}

// mergeCommittedThreadReactions keeps the committed reactions so a writer with
// a stale cache never erases them. Only a reaction write replaces the state of
// its own thread; it reloads fresh state right before it changes it.
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
	if threadID, ok := ctx.Value(threadReactionMutationContextKey{}).(string); ok && threadID != "" {
		if state, ok := requested.TaskThreadReactions[threadID]; ok {
			merged[threadID] = state
		} else {
			delete(merged, threadID)
		}
	}
	requested.TaskThreadReactions = retainedTaskThreadReactions(merged, requested.TaskThreads)
	return nil
}
