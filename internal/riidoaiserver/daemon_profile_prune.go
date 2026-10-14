package riidoaiserver

import "strings"

type daemonProfilePruneResult struct {
	daemons  int
	runtimes int
}

func (s *DevelopmentAIAgentClientStore) pruneUnexpectedDaemonProfiles(reason string) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	expected := s.expectedDaemonProfile
	if expected == "" {
		return false
	}
	pruned := s.pruneUnexpectedDaemonProfilesLocked("", expected)
	logDaemonProfilePruned("", expected, reason, pruned)
	return pruned.changed()
}

func (s *DevelopmentAIAgentClientStore) pruneUnexpectedDaemonProfilesLocked(deviceID, expected string) daemonProfilePruneResult {
	result := daemonProfilePruneResult{}
	if expected == "" {
		return result
	}
	for key, daemon := range s.daemons {
		if deviceID != "" && !daemonBelongsToDevice(daemon, deviceID) {
			continue
		}
		if daemonProfileMatches(expected, daemon.Profile) {
			continue
		}
		delete(s.daemons, key)
		result.daemons++
	}
	for deviceIndex := range s.devices {
		if deviceID != "" && strings.TrimSpace(s.devices[deviceIndex].DeviceID) != deviceID {
			continue
		}
		kept := s.devices[deviceIndex].Runtimes[:0]
		for _, runtime := range s.devices[deviceIndex].Runtimes {
			if daemonProfileMatches(expected, runtime.DaemonProfile) {
				kept = append(kept, runtime)
				continue
			}
			result.runtimes++
		}
		s.devices[deviceIndex].Runtimes = kept
	}
	return result
}

func (r daemonProfilePruneResult) changed() bool {
	return r.daemons > 0 || r.runtimes > 0
}
