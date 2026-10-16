package riidoaiserver

import (
	"context"
	"time"
)

// desktopPresenceSnapshotMaxAge bounds how stale a presence answer can be when
// the desktop app connected through another control-plane instance. The
// default snapshot reload interval is a minute, too slow for onboarding polls,
// while reloading on every poll would load the whole snapshot per request.
const desktopPresenceSnapshotMaxAge = 2 * time.Second

// DesktopPresenceLastSeenAt reloads a snapshot older than
// desktopPresenceSnapshotMaxAge first, so at most one reload per instance runs
// in that window however many web clients are polling.
func (s *PersistentAIAgentClientStore) DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error) {
	if err := s.lockOperationContext(ctx); err != nil {
		return time.Time{}, err
	}
	defer s.operationMu.Unlock()
	s.snapshotReloadMu.Lock()
	if !s.snapshotNow().Before(s.snapshotLastReloadAt.Add(desktopPresenceSnapshotMaxAge)) {
		s.snapshotLastReloadAt = time.Time{}
	}
	s.snapshotReloadMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return time.Time{}, err
	}
	return s.DevelopmentAIAgentClientStore.DesktopPresenceLastSeenAt(ctx, principal)
}
