package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

// DesktopPresenceLastSeenAt answers onboarding polls without reloading the
// snapshot. A reload restores every part and replaces in-memory daemon
// heartbeats that are saved only every few minutes, and the default reload
// interval is too slow for polling. Instead it reads only the saved connection
// grants, so a desktop connection saved by another instance shows up on the
// next poll; stores without a grants reader answer from memory.
func (s *PersistentAIAgentClientStore) DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error) {
	lastSeenAt, err := s.DevelopmentAIAgentClientStore.DesktopPresenceLastSeenAt(ctx, principal)
	principalID := strings.TrimSpace(principal.PrincipalID)
	if err != nil || principalID == "" {
		return lastSeenAt, err
	}
	reader, ok := s.snapshotStore.(deviceConnectionGrantsReader)
	if !ok {
		return lastSeenAt, nil
	}
	grants, err := reader.loadDeviceConnectionGrants(ctx)
	if err != nil {
		return time.Time{}, err
	}
	for _, grant := range grants {
		if strings.TrimSpace(grant.PrincipalID) == principalID && grant.LastSeenAt.After(lastSeenAt) {
			lastSeenAt = grant.LastSeenAt.UTC()
		}
	}
	return lastSeenAt, nil
}
