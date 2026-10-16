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
	if err := ctx.Err(); err != nil {
		return time.Time{}, err
	}
	principalID := strings.TrimSpace(principal.PrincipalID)
	if principalID == "" {
		return time.Time{}, ErrAuthorizationForbidden
	}
	reader, ok := s.snapshotStore.(deviceConnectionGrantsReader)
	if !ok {
		return s.DevelopmentAIAgentClientStore.DesktopPresenceLastSeenAt(ctx, principal)
	}
	grants, err := reader.loadDeviceConnectionGrants(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return desktopPresenceFromGrants(grants, principalID)
}
