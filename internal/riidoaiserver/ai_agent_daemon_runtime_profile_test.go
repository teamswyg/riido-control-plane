package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentDaemonRuntimeSnapshotKeepsProfilesSeparate(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-profiled", WorkspaceID: "workspace-profiled"}
	deviceID := "device-profiled"
	startedAt := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)

	syncProfileRuntime(t, store, ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:  "daemon-staging",
		DeviceID:  deviceID,
		Profile:   "staging",
		PID:       200,
		StartedAt: startedAt,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      "device-profiled:staging:codex",
			Kind:           RuntimeKindCodex,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	})
	syncProfileRuntime(t, store, ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:  "daemon-production",
		DeviceID:  deviceID,
		Profile:   "production",
		PID:       100,
		StartedAt: startedAt.Add(-time.Minute),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      "device-profiled:production:claude",
			Kind:           RuntimeKindClaudeCode,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	})

	staging := createAgentForSharedRuntime(t, store, ctx, principal, "device-profiled:staging:codex", "Staging")
	production := createAgentForSharedRuntime(t, store, ctx, principal, "device-profiled:production:claude", "Production")
	requireAgentDaemonBinding(t, store, staging.Agent.AgentID, "daemon-staging")
	requireAgentDaemonBinding(t, store, production.Agent.AgentID, "daemon-production")
	if got := countDaemonsForDevice(store, deviceID); got != 2 {
		t.Fatalf("daemon records for %s = %d, want 2", deviceID, got)
	}
}
