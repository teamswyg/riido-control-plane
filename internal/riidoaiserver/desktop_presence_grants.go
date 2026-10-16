package riidoaiserver

import (
	"errors"
	"strings"
	"time"
)

// Match snapshot restore's identity validation and legacy timestamp fallback.
func normalizedDeviceConnectionGrant(grant DeviceConnectionGrant) (DeviceConnectionGrant, error) {
	grant.DeviceID = strings.TrimSpace(grant.DeviceID)
	grant.PrincipalID = strings.TrimSpace(grant.PrincipalID)
	grant.WorkspaceID = strings.TrimSpace(grant.WorkspaceID)
	if grant.DeviceID == "" || grant.PrincipalID == "" || grant.WorkspaceID == "" {
		return DeviceConnectionGrant{}, errors.New("device connection grant identity is required")
	}
	if grant.ConnectedAt.IsZero() {
		grant.ConnectedAt = grant.LastSeenAt
	}
	if grant.LastSeenAt.IsZero() {
		grant.LastSeenAt = grant.ConnectedAt
	}
	grant.ConnectedAt, grant.LastSeenAt = grant.ConnectedAt.UTC(), grant.LastSeenAt.UTC()
	return grant, nil
}

func desktopPresenceFromGrants(grants []DeviceConnectionGrant, principalID string) (time.Time, error) {
	var lastSeenAt time.Time
	for _, grant := range grants {
		grant, err := normalizedDeviceConnectionGrant(grant)
		if err != nil {
			return time.Time{}, err
		}
		if grant.PrincipalID == principalID && grant.LastSeenAt.After(lastSeenAt) {
			lastSeenAt = grant.LastSeenAt
		}
	}
	return lastSeenAt, nil
}
