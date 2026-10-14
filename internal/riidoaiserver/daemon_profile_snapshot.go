package riidoaiserver

func (s *DevelopmentAIAgentClientStore) rejectUnexpectedDaemonProfileSnapshot(
	req DeviceRuntimeSnapshotSyncRequest,
) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	expected := s.expectedDaemonProfile
	if expected == "" || daemonProfileMatches(expected, req.Profile) {
		return expected, false, nil
	}
	pruned := s.pruneUnexpectedDaemonProfilesLocked(req.DeviceID, expected)
	logDaemonProfileMismatch(req.DeviceID, req.DaemonID, expected, req.Profile, pruned)
	return expected, pruned.changed(), daemonProfileMismatchError(expected, req.Profile)
}

func (s *DevelopmentAIAgentClientStore) pruneExpectedDaemonProfilesForSnapshotLocked(
	deviceID, expected string,
) daemonProfilePruneResult {
	if expected == "" {
		return daemonProfilePruneResult{}
	}
	pruned := s.pruneUnexpectedDaemonProfilesLocked(deviceID, expected)
	logDaemonProfilePruned(deviceID, expected, "runtime_snapshot", pruned)
	return pruned
}
