package riidoaiserver

import (
	"context"
	"time"
)

// DesktopPresenceLastSeenAt reloads the snapshot first so a desktop connection
// saved by another control-plane instance is visible to the polling web client.
func (s *PersistentAIAgentClientStore) DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error) {
	if err := s.lockOperationContext(ctx); err != nil {
		return time.Time{}, err
	}
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return time.Time{}, err
	}
	return s.DevelopmentAIAgentClientStore.DesktopPresenceLastSeenAt(ctx, principal)
}
