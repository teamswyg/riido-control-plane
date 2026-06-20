package riidoaiserver

import (
	"time"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func developmentSharedDevice(now time.Time) DeviceRecord {
	return DeviceRecord{
		DeviceID:         "device-shared-studio",
		OwnerPrincipalID: "user-2",
		DisplayName:      "Shared Studio Mac",
		DaemonLastSeenAt: now,
		Runtimes: []RuntimeRecord{
			{
				RuntimeID:        "runtime-openclaw-shared",
				DeviceID:         "device-shared-studio",
				Kind:             RuntimeKindOpenClaw,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				ProviderVersion:  "openclaw 0.1.0",
				OwnerPrincipalID: "user-2",
				LastDetectedAt:   now,
				HasAssignedAgent: true,
				Models: []RuntimeModelRecord{
					{ModelID: providercatalog.DefaultOpenClawModelID, Label: "OpenClaw 기본 모델", IsDefault: true},
				},
			},
			developmentPrivateCursorRuntime(now),
		},
	}
}

func developmentPrivateCursorRuntime(now time.Time) RuntimeRecord {
	return RuntimeRecord{
		RuntimeID:        "runtime-cursor-private",
		DeviceID:         "device-shared-studio",
		Kind:             RuntimeKindCursor,
		Availability:     RuntimeAvailabilityOnline,
		DetectionState:   RuntimeDetectionStateDetected,
		ProviderVersion:  "cursor-agent 1.0.0",
		OwnerPrincipalID: "user-2",
		LastDetectedAt:   now,
		HasAssignedAgent: true,
		Models: []RuntimeModelRecord{
			{ModelID: providercatalog.DefaultCursorModelID, Label: "Cursor Auto", IsDefault: true},
			{ModelID: "cursor-fast", Label: "Cursor Fast", IsDefault: false},
		},
	}
}
