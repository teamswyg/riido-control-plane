package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

// Stale runtime rows from older daemon builds (same kind, different/legacy
// runtime IDs) accumulate in the stored device. The client device list must
// collapse them to one runtime per kind, preferring the live one, so the UI
// does not render duplicate "Claude Code"/"Codex" entries.
func TestListAIAgentDevicesDedupesRuntimesByKind(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	now := time.Now().UTC()

	store.mu.Lock()
	store.devices = append(store.devices, DeviceRecord{
		DeviceID:         "dev_dedupe",
		OwnerPrincipalID: "user-dedupe",
		DaemonLastSeenAt: now, // fresh → liveness projection keeps online ones online
		Runtimes: []RuntimeRecord{
			// stale legacy id, offline
			{RuntimeID: "mlaude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOffline, DetectionState: RuntimeDetectionStateMissing, LastDetectedAt: now.Add(-time.Hour)},
			// current id, online — should win
			{RuntimeID: "m:claude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, LastDetectedAt: now},
			{RuntimeID: "m:codex", Kind: RuntimeKindCodex, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, LastDetectedAt: now},
		},
	})
	store.mu.Unlock()

	resp, err := store.ListAIAgentDevices(ctx, AuthorizationResult{PrincipalID: "user-dedupe"})
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	var dev *DeviceRecord
	for i := range resp.Devices {
		if resp.Devices[i].DeviceID == "dev_dedupe" {
			dev = &resp.Devices[i]
			break
		}
	}
	if dev == nil {
		t.Fatal("device dev_dedupe not visible to owner")
	}
	byKind := map[RuntimeKind][]RuntimeRecord{}
	for _, rt := range dev.Runtimes {
		byKind[rt.Kind] = append(byKind[rt.Kind], rt)
	}
	if len(byKind[RuntimeKindClaudeCode]) != 1 {
		t.Fatalf("claude_code not deduped: %+v", byKind[RuntimeKindClaudeCode])
	}
	if got := byKind[RuntimeKindClaudeCode][0]; got.RuntimeID != "m:claude" {
		t.Fatalf("dedup kept stale runtime %q, want live m:claude", got.RuntimeID)
	}
	if len(byKind[RuntimeKindCodex]) != 1 {
		t.Fatalf("codex count = %d, want 1", len(byKind[RuntimeKindCodex]))
	}
}
