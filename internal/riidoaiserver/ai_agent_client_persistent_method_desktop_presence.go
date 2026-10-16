package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

// DesktopPresenceLastSeenAt reads the saved snapshot's connection grants
// without restoring it into memory. A full reload would replace in-memory
// daemon heartbeats that are saved only every few minutes, while the default
// reload interval is too slow for onboarding polls; a desktop connection saved
// by another control-plane instance is still visible on the next poll.
func (s *PersistentAIAgentClientStore) DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error) {
	lastSeenAt, err := s.DevelopmentAIAgentClientStore.DesktopPresenceLastSeenAt(ctx, principal)
	principalID := strings.TrimSpace(principal.PrincipalID)
	if err != nil || principalID == "" || s.snapshotStore == nil {
		return lastSeenAt, err
	}
	snapshot, ok, err := s.snapshotStore.LoadAIAgentClientSnapshot(ctx)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return lastSeenAt, nil
	}
	for _, grant := range snapshot.DeviceConnectionGrants {
		if strings.TrimSpace(grant.PrincipalID) == principalID && grant.LastSeenAt.After(lastSeenAt) {
			lastSeenAt = grant.LastSeenAt.UTC()
		}
	}
	return lastSeenAt, nil
}
