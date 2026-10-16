package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

// AIAgentDesktopPresenceResponse tells web onboarding when the same account
// last connected a desktop app. Clients compare LastSeenAt only against
// server time derived from CheckedAt, never against the browser clock.
type AIAgentDesktopPresenceResponse struct {
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CheckedAt  time.Time  `json:"checked_at"`
}

// AIAgentDesktopPresenceReader returns the latest desktop connection time of
// the principal across its device connection grants. A zero time means the
// principal has never connected a desktop app.
type AIAgentDesktopPresenceReader interface {
	DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error)
}

func (s *DevelopmentAIAgentClientStore) DesktopPresenceLastSeenAt(ctx context.Context, principal AuthorizationResult) (time.Time, error) {
	if err := ctx.Err(); err != nil {
		return time.Time{}, err
	}
	principalID := strings.TrimSpace(principal.PrincipalID)
	if principalID == "" {
		return time.Time{}, ErrAuthorizationForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var lastSeenAt time.Time
	for _, deviceGrants := range s.deviceConnectionGrants {
		for _, grant := range deviceGrants {
			if grant.PrincipalID == principalID && grant.LastSeenAt.After(lastSeenAt) {
				lastSeenAt = grant.LastSeenAt
			}
		}
	}
	return lastSeenAt.UTC(), nil
}

func newAIAgentDesktopPresenceResponse(lastSeenAt, checkedAt time.Time) AIAgentDesktopPresenceResponse {
	response := AIAgentDesktopPresenceResponse{CheckedAt: checkedAt.UTC()}
	if !lastSeenAt.IsZero() {
		seen := lastSeenAt.UTC()
		response.LastSeenAt = &seen
	}
	return response
}
