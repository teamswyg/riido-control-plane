package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestNormalizeProviderStatusSyncMatchesPrivateContract(t *testing.T) {
	req := ProviderStatusSyncRequest{
		DaemonID:            " daemon-a ",
		DeviceID:            " device-a ",
		RuntimeID:           " runtime-a ",
		DistributionChannel: hostintegration.DistributionChannelMSIXStore,
		AppVersion:          " 1.2.3 ",
		Providers: []ProviderStatusRecord{
			{ProviderKind: " cursor ", RoutingStatus: hostintegration.ProviderRoutingLoginRequired},
			{ProviderKind: " codex ", RoutingStatus: hostintegration.ProviderRoutingAvailable},
		},
	}
	got, err := normalizeProviderStatusSync(" agent-a ", req)
	if err != nil {
		t.Fatalf("normalizeProviderStatusSync: %v", err)
	}
	if got.DaemonID != "daemon-a" || got.DeviceID != "device-a" || got.RuntimeID != "runtime-a" || got.AppVersion != "1.2.3" {
		t.Fatalf("trimmed request = %+v", got)
	}
	if got.DistributionChannel != hostintegration.DistributionChannelMSIXStore {
		t.Fatalf("distribution channel = %q", got.DistributionChannel)
	}
	if got.Providers[0].ProviderKind != "codex" || got.Providers[1].ProviderKind != "cursor" {
		t.Fatalf("providers not normalized/sorted: %+v", got.Providers)
	}
}

func TestNormalizeProviderStatusSyncRejectsInvalidRequests(t *testing.T) {
	base := ProviderStatusSyncRequest{
		DaemonID:            "daemon-a",
		RuntimeID:           "runtime-a",
		DistributionChannel: hostintegration.DistributionChannelDeveloperID,
		Providers: []ProviderStatusRecord{{
			ProviderKind:  capability.ProviderKind("codex"),
			RoutingStatus: hostintegration.ProviderRoutingAvailable,
		}},
	}
	cases := []struct {
		name string
		edit func(*ProviderStatusSyncRequest)
		want string
	}{
		{name: "missing daemon id", edit: func(req *ProviderStatusSyncRequest) { req.DaemonID = " " }, want: "daemon_id is required"},
		{name: "missing runtime id", edit: func(req *ProviderStatusSyncRequest) { req.RuntimeID = " " }, want: "runtime_id is required"},
		{name: "unknown channel", edit: func(req *ProviderStatusSyncRequest) { req.DistributionChannel = "snap-store" }, want: "unknown distribution channel"},
		{name: "missing providers", edit: func(req *ProviderStatusSyncRequest) { req.Providers = nil }, want: "providers is required"},
		{name: "missing provider kind", edit: func(req *ProviderStatusSyncRequest) { req.Providers[0].ProviderKind = " " }, want: "provider_kind is required"},
		{name: "unknown routing status", edit: func(req *ProviderStatusSyncRequest) { req.Providers[0].RoutingStatus = "maybe" }, want: "routing_status is invalid"},
		{name: "duplicate provider", edit: func(req *ProviderStatusSyncRequest) {
			req.Providers = append(req.Providers, ProviderStatusRecord{
				ProviderKind:  " codex ",
				RoutingStatus: hostintegration.ProviderRoutingLoginRequired,
			})
		}, want: "provider_kind duplicates codex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			req.Providers = append([]ProviderStatusRecord(nil), base.Providers...)
			tc.edit(&req)
			_, err := normalizeProviderStatusSync("agent-a", req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("normalizeProviderStatusSync() error = %v, want %q", err, tc.want)
			}
		})
	}
	if _, err := normalizeProviderStatusSync(" ", base); err == nil || !strings.Contains(err.Error(), "agent_id is required") {
		t.Fatalf("missing agent id error = %v", err)
	}
}

func TestProviderStatusJSONShape(t *testing.T) {
	syncedAt := time.Date(2026, 5, 27, 15, 0, 0, 0, time.UTC)
	response := ProviderStatusSyncResponse{
		SchemaVersion:       SchemaVersion,
		AgentID:             "agent-a",
		DaemonID:            "daemon-a",
		DeviceID:            "device-a",
		RuntimeID:           "runtime-a",
		DistributionChannel: hostintegration.DistributionChannelMacAppStore,
		AppVersion:          "1.2.3",
		Providers: []ProviderStatusRecord{{
			ProviderKind:  capability.ProviderKind("codex"),
			RoutingStatus: hostintegration.ProviderRoutingStoreBlocked,
		}},
		SyncedAt: syncedAt,
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal provider status response: %v", err)
	}
	want := `{"schema_version":"riido-ai-server.v1","agent_id":"agent-a","daemon_id":"daemon-a","device_id":"device-a","runtime_id":"runtime-a","distribution_channel":"mac-app-store","app_version":"1.2.3","providers":[{"provider_kind":"codex","routing_status":"store-blocked"}],"synced_at":"2026-05-27T15:00:00Z"}`
	if got := string(data); got != want {
		t.Fatalf("provider status response JSON = %s, want %s", got, want)
	}
}

func TestProviderStatusPortsCompile(t *testing.T) {
	store := newProviderStatusTestStore(time.Now().UTC())
	var writer ProviderStatusStore = store
	var reader ProviderStatusReader = store
	req := ProviderStatusSyncRequest{
		DaemonID:            "daemon-a",
		RuntimeID:           "runtime-a",
		DistributionChannel: hostintegration.DistributionChannelDevLocal,
		Providers: []ProviderStatusRecord{{
			ProviderKind:  capability.ProviderKind("claude"),
			RoutingStatus: hostintegration.ProviderRoutingLoginRequired,
		}},
	}
	if _, err := writer.SyncProviderStatus(context.Background(), "agent-a", req); err != nil {
		t.Fatalf("SyncProviderStatus: %v", err)
	}
	if _, ok, err := reader.GetProviderStatus(context.Background(), "agent-a"); err != nil || !ok {
		t.Fatalf("GetProviderStatus ok=%v err=%v", ok, err)
	}
}

type providerStatusTestStore struct {
	syncedAt time.Time
	records  map[string]ProviderStatusSyncResponse
}

func newProviderStatusTestStore(syncedAt time.Time) *providerStatusTestStore {
	return &providerStatusTestStore{syncedAt: syncedAt, records: map[string]ProviderStatusSyncResponse{}}
}

func (s *providerStatusTestStore) SyncProviderStatus(_ context.Context, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error) {
	normalized, err := normalizeProviderStatusSync(agentID, req)
	if err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	response := ProviderStatusSyncResponse{
		SchemaVersion:       SchemaVersion,
		AgentID:             strings.TrimSpace(agentID),
		DaemonID:            normalized.DaemonID,
		DeviceID:            normalized.DeviceID,
		RuntimeID:           normalized.RuntimeID,
		DistributionChannel: normalized.DistributionChannel,
		AppVersion:          normalized.AppVersion,
		Providers:           append([]ProviderStatusRecord(nil), normalized.Providers...),
		SyncedAt:            s.syncedAt,
	}
	s.records[response.AgentID] = response
	return response, nil
}

func (s *providerStatusTestStore) GetProviderStatus(_ context.Context, agentID string) (ProviderStatusSyncResponse, bool, error) {
	response, ok := s.records[strings.TrimSpace(agentID)]
	if !ok {
		return ProviderStatusSyncResponse{}, false, nil
	}
	response.Providers = append([]ProviderStatusRecord(nil), response.Providers...)
	return response, true, nil
}
