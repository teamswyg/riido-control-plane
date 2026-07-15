package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// DeviceConnectionGrant records which authenticated account connected a
// physical device to a workspace. Device ownership and credentials remain
// unchanged; this is an append-only access/recovery fact.
type DeviceConnectionGrant struct {
	DeviceID    string    `json:"device_id"`
	PrincipalID string    `json:"principal_id"`
	WorkspaceID string    `json:"workspace_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

func deviceConnectionGrantKey(deviceID, principalID, workspaceID string) string {
	return strings.TrimSpace(deviceID) + "\x00" + strings.TrimSpace(principalID) + "\x00" + strings.TrimSpace(workspaceID)
}

func (s *DevelopmentAIAgentClientStore) upsertDeviceConnectionGrantLocked(deviceID string, principal AuthorizationResult, now time.Time) DeviceConnectionGrant {
	if s.deviceConnectionGrants == nil {
		s.deviceConnectionGrants = map[string]map[string]DeviceConnectionGrant{}
	}
	deviceID = strings.TrimSpace(deviceID)
	principalID := strings.TrimSpace(principal.PrincipalID)
	workspaceID := strings.TrimSpace(principal.WorkspaceID)
	deviceGrants := s.deviceConnectionGrants[deviceID]
	if deviceGrants == nil {
		deviceGrants = map[string]DeviceConnectionGrant{}
		s.deviceConnectionGrants[deviceID] = deviceGrants
	}
	key := deviceConnectionGrantKey(deviceID, principalID, workspaceID)
	grant, ok := deviceGrants[key]
	if !ok {
		grant = DeviceConnectionGrant{
			DeviceID:    deviceID,
			PrincipalID: principalID,
			WorkspaceID: workspaceID,
			ConnectedAt: now.UTC(),
		}
	}
	grant.LastSeenAt = now.UTC()
	deviceGrants[key] = grant
	return grant
}

func (s *DevelopmentAIAgentClientStore) deviceConnectionSummaryLocked(deviceID string) (string, int) {
	keys := make([]string, 0)
	principals := map[string]struct{}{}
	for _, grant := range s.deviceConnectionGrants[deviceID] {
		keys = append(keys, deviceConnectionGrantKey(grant.DeviceID, grant.PrincipalID, grant.WorkspaceID))
		principals[grant.PrincipalID] = struct{}{}
	}
	if len(keys) == 0 {
		return "", 0
	}
	sort.Strings(keys)
	hash := sha256.Sum256([]byte(strings.Join(keys, "\n")))
	return hex.EncodeToString(hash[:8]), len(principals)
}
