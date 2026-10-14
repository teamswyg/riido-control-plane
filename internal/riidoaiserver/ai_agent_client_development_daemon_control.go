package riidoaiserver

import (
	"errors"
	"strconv"
	"time"
)

func (s *DevelopmentAIAgentClientStore) applyDaemonControlLocked(daemon DeviceDaemonRecord, action DaemonControlAction) (developmentDaemonControlResult, error) {
	now := time.Now().UTC()
	result := developmentDaemonControlResult{
		CommandID:  "daemon-command-" + strconv.Itoa(s.nextDaemonCommand),
		Daemon:     daemon,
		AcceptedAt: now,
	}
	s.nextDaemonCommand++
	result.Daemon.LastCommandID = result.CommandID
	result.Daemon.LastCommandAction = action
	result.Daemon.LastCommandRequestedAt = now
	result.Daemon.LastSeenAt = now
	message, err := s.applyDaemonActionLocked(&result.Daemon, action, now)
	if err != nil {
		return developmentDaemonControlResult{}, err
	}
	result.Message = message
	s.publishDaemonControlLocked(result.Daemon, action)
	return result, nil
}

func (s *DevelopmentAIAgentClientStore) applyDaemonActionLocked(daemon *DeviceDaemonRecord, action DaemonControlAction, now time.Time) (string, error) {
	switch action {
	case DaemonControlActionStart:
		daemon.Availability = DaemonAvailabilityOnline
		daemon.ControlState = DaemonControlStateStarting
		daemon.StartedAt = now
		daemon.PID = 5111
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop}
		return "daemon start command accepted", nil
	case DaemonControlActionRestart:
		daemon.Availability = DaemonAvailabilityOnline
		daemon.ControlState = DaemonControlStateRestarting
		daemon.StartedAt = now
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionStop}
		return "daemon restart command accepted", nil
	case DaemonControlActionStop:
		daemon.Availability = DaemonAvailabilityOffline
		daemon.ControlState = DaemonControlStateStopping
		daemon.PID = 0
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionStart}
		s.markDeviceRuntimesOfflineLocked(daemon.DeviceID, now)
		return "daemon stop command accepted", nil
	default:
		return "", errors.New("unsupported daemon action")
	}
}
