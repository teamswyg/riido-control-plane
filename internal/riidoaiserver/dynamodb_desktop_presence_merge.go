package riidoaiserver

// Connection grants are append-only historical facts, so unrelated stale
// snapshot writers preserve every committed key and the latest saved timestamp.
// CURRENT's existing CAS fences a concurrent connection after the atomic read.
func mergeCommittedDeviceConnectionGrants(items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	var committed AIAgentClientSnapshot
	if err := decodeCommittedAIAgentClientPart(dynamoDBAIAgentClientSnapshotPartDeviceConnections, current, parts[dynamoDBAIAgentClientSnapshotPartDeviceConnections], items, &committed); err != nil {
		return err
	}
	byDevice := map[string]map[string]DeviceConnectionGrant{}
	for _, grants := range [][]DeviceConnectionGrant{committed.DeviceConnectionGrants, requested.DeviceConnectionGrants} {
		for _, grant := range grants {
			grant, err := normalizedDeviceConnectionGrant(grant)
			if err != nil {
				return err
			}
			if byDevice[grant.DeviceID] == nil {
				byDevice[grant.DeviceID] = map[string]DeviceConnectionGrant{}
			}
			key := deviceConnectionGrantKey(grant.DeviceID, grant.PrincipalID, grant.WorkspaceID)
			if old, ok := byDevice[grant.DeviceID][key]; ok {
				if old.LastSeenAt.After(grant.LastSeenAt) {
					grant.LastSeenAt = old.LastSeenAt
				}
				if grant.ConnectedAt.IsZero() || (!old.ConnectedAt.IsZero() && old.ConnectedAt.Before(grant.ConnectedAt)) {
					grant.ConnectedAt = old.ConnectedAt
				}
			}
			byDevice[grant.DeviceID][key] = grant
		}
	}
	requested.DeviceConnectionGrants = copyDeviceConnectionGrants(byDevice)
	return nil
}
