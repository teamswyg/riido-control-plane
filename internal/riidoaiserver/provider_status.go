package riidoaiserver

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

type ProviderStatusRecord struct {
	ProviderKind  capability.ProviderKind               `json:"provider_kind"`
	RoutingStatus hostintegration.ProviderRoutingStatus `json:"routing_status"`
}

type ProviderStatusSyncRequest struct {
	DaemonID            string                              `json:"daemon_id"`
	DeviceID            string                              `json:"device_id,omitempty"`
	RuntimeID           string                              `json:"runtime_id"`
	DistributionChannel hostintegration.DistributionChannel `json:"distribution_channel"`
	AppVersion          string                              `json:"app_version,omitempty"`
	Providers           []ProviderStatusRecord              `json:"providers"`
}

type ProviderStatusSyncResponse struct {
	SchemaVersion       string                              `json:"schema_version"`
	AgentID             string                              `json:"agent_id"`
	DaemonID            string                              `json:"daemon_id"`
	DeviceID            string                              `json:"device_id,omitempty"`
	RuntimeID           string                              `json:"runtime_id"`
	DistributionChannel hostintegration.DistributionChannel `json:"distribution_channel"`
	AppVersion          string                              `json:"app_version,omitempty"`
	Providers           []ProviderStatusRecord              `json:"providers"`
	SyncedAt            time.Time                           `json:"synced_at"`
}

func normalizeProviderStatusSync(agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncRequest, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return ProviderStatusSyncRequest{}, errors.New("agent_id is required")
	}
	req.DaemonID = strings.TrimSpace(req.DaemonID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.RuntimeID = strings.TrimSpace(req.RuntimeID)
	req.AppVersion = strings.TrimSpace(req.AppVersion)
	if req.DaemonID == "" {
		return ProviderStatusSyncRequest{}, errors.New("daemon_id is required")
	}
	if req.RuntimeID == "" {
		return ProviderStatusSyncRequest{}, errors.New("runtime_id is required")
	}
	if !req.DistributionChannel.Valid() {
		return ProviderStatusSyncRequest{}, fmt.Errorf("unknown distribution channel %q", req.DistributionChannel)
	}
	if len(req.Providers) == 0 {
		return ProviderStatusSyncRequest{}, errors.New("providers is required")
	}
	seenProviders := map[capability.ProviderKind]struct{}{}
	for i := range req.Providers {
		req.Providers[i].ProviderKind = capability.ProviderKind(strings.TrimSpace(string(req.Providers[i].ProviderKind)))
		if req.Providers[i].ProviderKind == "" {
			return ProviderStatusSyncRequest{}, fmt.Errorf("providers[%d].provider_kind is required", i)
		}
		if !req.Providers[i].RoutingStatus.Valid() {
			return ProviderStatusSyncRequest{}, fmt.Errorf("providers[%d].routing_status is invalid", i)
		}
		if _, ok := seenProviders[req.Providers[i].ProviderKind]; ok {
			return ProviderStatusSyncRequest{}, fmt.Errorf("providers[%d].provider_kind duplicates %s", i, req.Providers[i].ProviderKind)
		}
		seenProviders[req.Providers[i].ProviderKind] = struct{}{}
	}
	sort.Slice(req.Providers, func(i, j int) bool {
		return req.Providers[i].ProviderKind < req.Providers[j].ProviderKind
	})
	return req, nil
}
