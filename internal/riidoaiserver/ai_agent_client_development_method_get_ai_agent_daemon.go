package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return DeviceDaemonDetailResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, daemon, ok := s.deviceDaemonForAgentAccessLocked(principal, agentID)
	if !ok {
		return DeviceDaemonDetailResponse{}, ErrAIAgentNotFound
	}
	response := DeviceDaemonDetailResponse{SchemaVersion: SchemaVersion, Daemon: copyDeviceDaemon(daemon)}
	if _, runtime, ok := s.deviceRuntimeByRuntimeIDLocked(agent.RuntimeID); ok {
		response.Runtime = &runtime
	}
	return response, nil
}
