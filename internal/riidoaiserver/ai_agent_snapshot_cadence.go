package riidoaiserver

import (
	"errors"
	"time"
)

func (s *PersistentAIAgentClientStore) ConfigureSnapshotCadence(reloadInterval, heartbeatSaveInterval time.Duration) error {
	if s == nil {
		return nil
	}
	if reloadInterval < 0 {
		return errors.New("riidoaiserver: AI Agent client snapshot reload interval must not be negative")
	}
	if heartbeatSaveInterval < 0 {
		return errors.New("riidoaiserver: AI Agent client heartbeat snapshot save interval must not be negative")
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	if reloadInterval == 0 {
		reloadInterval = s.snapshotReloadInterval
	}
	if heartbeatSaveInterval == 0 {
		heartbeatSaveInterval = s.snapshotHeartbeatSaveInterval
	}
	if reloadInterval >= AIAgentDeviceRuntimeSnapshotStaleAfter ||
		heartbeatSaveInterval >= AIAgentDeviceRuntimeSnapshotStaleAfter-reloadInterval {
		return errors.New("riidoaiserver: snapshot reload and heartbeat save intervals must total less than the runtime liveness window")
	}
	s.snapshotReloadInterval = reloadInterval
	s.snapshotHeartbeatSaveInterval = heartbeatSaveInterval
	return nil
}

func (s *PersistentAIAgentClientStore) shouldSaveHeartbeatSnapshot(daemon DeviceDaemonRecord) bool {
	if s == nil {
		return true
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	lastSeen := s.snapshotDaemonLastSeen[daemonStorageKey(daemon)]
	if s.snapshotHeartbeatSaveInterval <= 0 || lastSeen.IsZero() {
		return true
	}
	return !daemon.LastSeenAt.Before(lastSeen.Add(s.snapshotHeartbeatSaveInterval))
}

func (s *PersistentAIAgentClientStore) shouldReloadSnapshotLocked() bool {
	if s.snapshotReloadInterval <= 0 || s.snapshotLastReloadAt.IsZero() {
		return true
	}
	return !s.snapshotNow().Before(s.snapshotLastReloadAt.Add(s.snapshotReloadInterval))
}

func (s *PersistentAIAgentClientStore) markSnapshotObserved(snapshot AIAgentClientSnapshot) {
	if s == nil {
		return
	}
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	s.snapshotLastReloadAt = s.snapshotNow()
	s.recordSnapshotDaemonLastSeenLocked(snapshot)
}

func (s *PersistentAIAgentClientStore) recordSnapshotDaemonLastSeenLocked(snapshot AIAgentClientSnapshot) {
	// A save by another device does not prove this daemon's heartbeat was saved.
	s.snapshotDaemonLastSeen = make(map[string]time.Time, len(snapshot.Daemons))
	for _, daemon := range snapshot.Daemons {
		s.snapshotDaemonLastSeen[daemonStorageKey(daemon)] = daemon.LastSeenAt
	}
}

func (s *PersistentAIAgentClientStore) snapshotNow() time.Time {
	if s != nil && s.snapshotClock != nil {
		return s.snapshotClock()
	}
	return time.Now()
}

func (s *PersistentAIAgentClientStore) markSnapshotSaved(snapshot AIAgentClientSnapshot) {
	s.snapshotReloadMu.Lock()
	defer s.snapshotReloadMu.Unlock()
	// Local writes do not refresh data committed by other server instances.
	s.recordSnapshotDaemonLastSeenLocked(snapshot)
}
