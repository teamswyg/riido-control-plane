package riidoaiserver

import (
	"errors"
	"strings"
)

func (s *Store) handleProviderStatusSync(state *storeState, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error) {
	agentID = strings.TrimSpace(agentID)
	req, err := normalizeProviderStatusSync(agentID, req)
	if err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	state.providerStatusSyncTotal++
	response := ProviderStatusSyncResponse{
		SchemaVersion:       SchemaVersion,
		AgentID:             agentID,
		DaemonID:            req.DaemonID,
		DeviceID:            req.DeviceID,
		RuntimeID:           req.RuntimeID,
		DistributionChannel: req.DistributionChannel,
		AppVersion:          req.AppVersion,
		Providers:           cloneProviderStatusRecords(req.Providers),
		SyncedAt:            s.now(),
	}
	state.providerStatuses[agentID] = response
	return cloneProviderStatusResponse(response), nil
}

func handleGetProviderStatus(state *storeState, agentID string) (ProviderStatusSyncResponse, bool, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return ProviderStatusSyncResponse{}, false, errors.New("agent_id is required")
	}
	response, ok := state.providerStatuses[agentID]
	if !ok {
		return ProviderStatusSyncResponse{}, false, nil
	}
	return cloneProviderStatusResponse(response), true, nil
}
