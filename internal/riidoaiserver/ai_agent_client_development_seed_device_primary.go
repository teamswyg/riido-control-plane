package riidoaiserver

import (
	"time"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func developmentPrimaryDevice(now time.Time) DeviceRecord {
	return DeviceRecord{
		DeviceID:         "device-dev-macbook",
		OwnerPrincipalID: "user-1",
		DisplayName:      "Development MacBook Pro",
		DaemonLastSeenAt: now,
		Runtimes: []RuntimeRecord{
			{
				RuntimeID:        "runtime-codex-dev",
				DeviceID:         "device-dev-macbook",
				Kind:             RuntimeKindCodex,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				ProviderVersion:  "codex-cli 0.133.0",
				OwnerPrincipalID: "user-1",
				LastDetectedAt:   now,
				HasAssignedAgent: true,
				Models: []RuntimeModelRecord{
					{ModelID: providercatalog.DefaultCodexModelID, Label: "Codex 기본 모델", IsDefault: true},
				},
			},
			{
				RuntimeID:        "runtime-claude-code-dev",
				DeviceID:         "device-dev-macbook",
				Kind:             RuntimeKindClaudeCode,
				Availability:     RuntimeAvailabilityOffline,
				DetectionState:   RuntimeDetectionStateMissing,
				ProviderVersion:  "2.1.142 (Claude Code)",
				OwnerPrincipalID: "user-1",
				LastDetectedAt:   now.Add(-30 * time.Second),
				HasAssignedAgent: true,
				Models: []RuntimeModelRecord{
					{ModelID: "claude-sonnect-4-6", Label: "Sonnect 4.6 (기본값)", IsDefault: true},
					{ModelID: "claude-opus-4-7", Label: "Opus 4.7", IsDefault: false},
					{ModelID: "claude-haiku-4-5", Label: "Haiku 4.5", IsDefault: false},
					{ModelID: "claude-opus-4-6", Label: "Opus 4.6", IsDefault: false},
					{ModelID: "claude-opus-3", Label: "Opus 3", IsDefault: false},
				},
			},
			developmentCursorRuntime(now),
		},
	}
}

func developmentCursorRuntime(now time.Time) RuntimeRecord {
	return RuntimeRecord{
		RuntimeID:        "runtime-cursor-dev",
		DeviceID:         "device-dev-macbook",
		Kind:             RuntimeKindCursor,
		Availability:     RuntimeAvailabilityOnline,
		DetectionState:   RuntimeDetectionStateDetected,
		ProviderVersion:  "cursor-agent 1.0.0",
		OwnerPrincipalID: "user-1",
		LastDetectedAt:   now,
		HasAssignedAgent: false,
		Models: []RuntimeModelRecord{
			{ModelID: providercatalog.DefaultCursorModelID, Label: "Cursor Auto", IsDefault: true},
			{ModelID: "cursor-fast", Label: "Cursor Fast", IsDefault: false},
		},
	}
}
