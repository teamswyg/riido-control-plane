package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ControlAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return DeviceDaemonCommandResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, daemon, ok := s.deviceDaemonForAgentAccessLocked(principal, agentID)
	if !ok {
		return DeviceDaemonCommandResponse{}, ErrAIAgentNotFound
	}
	if !daemonSupportsAction(daemon, action) {
		return DeviceDaemonCommandResponse{}, errors.New("daemon action is not supported in the current state")
	}
	result, err := s.applyDaemonControlLocked(daemon, action)
	if err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	return DeviceDaemonCommandResponse{
		SchemaVersion: SchemaVersion,
		CommandID:     result.CommandID,
		DeviceID:      result.Daemon.DeviceID,
		Action:        action,
		Availability:  result.Daemon.Availability,
		ControlState:  result.Daemon.ControlState,
		AcceptedAt:    result.AcceptedAt,
		Message:       result.Message + " for agent " + agent.AgentID,
	}, nil
}
