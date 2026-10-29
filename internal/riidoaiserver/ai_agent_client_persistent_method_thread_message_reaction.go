package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ReactToAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, add bool) (threadMessageReactionChange, error) {
	var change threadMessageReactionChange
	err := s.saveThreadReactionUpdate(ctx, threadID, func() (bool, error) {
		var err error
		change, err = s.DevelopmentAIAgentClientStore.ReactToAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, messageID, emoji, add)
		return change.Changed, err
	})
	return change, err
}

// saveThreadReactionUpdate changes one thread's reaction state on freshly
// reloaded state and saves only that thread's part of the reactions. A failed
// save forces the next operation to reload, so the unsaved change is dropped
// instead of being reported as an already applied no-op.
func (s *PersistentAIAgentClientStore) saveThreadReactionUpdate(ctx context.Context, threadID string, update func() (bool, error)) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	s.requireSnapshotReload()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	changed, err := update()
	if err != nil || !changed {
		return err
	}
	if err := s.saveSnapshot(withThreadReactionMutation(ctx, threadID)); err != nil {
		s.requireSnapshotReload()
		return err
	}
	return nil
}

func (s *PersistentAIAgentClientStore) requireSnapshotReload() {
	if s == nil || s.snapshotStore == nil {
		return
	}
	s.snapshotReloadMu.Lock()
	s.snapshotReloadRequired = true
	s.snapshotReloadMu.Unlock()
}
