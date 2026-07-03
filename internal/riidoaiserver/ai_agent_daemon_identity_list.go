package riidoaiserver

import (
	"sort"
	"time"
)

func (s *DevelopmentAIAgentClientStore) visibleDaemonsForDeviceLocked(principal AuthorizationResult, deviceID string) ([]DeviceDaemonRecord, bool) {
	device, ok := s.deviceByIDLocked(deviceID)
	if !ok || !s.deviceVisibleToPrincipalLocked(principal, device) {
		return nil, false
	}
	now := time.Now().UTC()
	out := make([]DeviceDaemonRecord, 0, len(s.daemons))
	for _, daemon := range s.daemons {
		if !daemonBelongsToDevice(daemon, deviceID) || !s.daemonVisibleToPrincipalLocked(principal, daemon) {
			continue
		}
		out = append(out, projectDeviceDaemonLiveness(daemon, now))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].LastSeenAt.Equal(out[j].LastSeenAt) {
			return out[i].LastSeenAt.After(out[j].LastSeenAt)
		}
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out, true
}
