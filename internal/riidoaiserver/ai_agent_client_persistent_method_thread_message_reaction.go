package riidoaiserver

import (
	"context"
	"errors"
)

func (s *PersistentAIAgentClientStore) ReactToAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, add bool) (threadMessageReactionChange, error) {
	var change threadMessageReactionChange
	err := s.saveThreadReactionUpdate(ctx, threadID, func() (bool, error) {
		var err error
		change, err = s.DevelopmentAIAgentClientStore.ReactToAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, messageID, emoji, add)
		return change.Changed, err
	})
	return change, err
}

// threadReactionSaveAttempts bounds the reload and reapply loop when another
// replica commits between the reload and the save.
const threadReactionSaveAttempts = 3

// saveThreadReactionUpdate changes one thread's reaction state on freshly
// reloaded state and saves it fenced against every part committed since the
// reload. A rejected save reloads and applies the update again; the updates
// are idempotent. Any failed save forces the next operation to reload, so the
// unsaved change is never reported as an already applied no-op.
func (s *PersistentAIAgentClientStore) saveThreadReactionUpdate(ctx context.Context, threadID string, update func() (bool, error)) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	var err error
	for attempt := 0; attempt < threadReactionSaveAttempts; attempt++ {
		if err = s.saveThreadReactionAttempt(ctx, threadID, update); !errors.Is(err, errTaskAgentSnapshotChanged) {
			return err
		}
	}
	return err
}

func (s *PersistentAIAgentClientStore) saveThreadReactionAttempt(ctx context.Context, threadID string, update func() (bool, error)) error {
	s.requireSnapshotReload()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	mutation := threadReactionMutation{threadID: threadID, before: s.taskThreadReactionState(threadID)}
	changed, err := update()
	if err != nil || !changed {
		return err
	}
	if err := s.saveSnapshot(withThreadReactionMutation(ctx, mutation)); err != nil {
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

func (s *DevelopmentAIAgentClientStore) taskThreadReactionState(threadID string) AIAgentTaskThreadReactionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.taskThreadReactions[threadID]
}
