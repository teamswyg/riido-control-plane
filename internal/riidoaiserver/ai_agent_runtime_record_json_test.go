package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestRuntimeRecordJSONIncludesEmptyProviderVersion(t *testing.T) {
	body, err := json.Marshal(RuntimeRecord{
		RuntimeID:                 "runtime-empty-version",
		DeviceID:                  "device-1",
		Kind:                      RuntimeKindCodex,
		Availability:              RuntimeAvailabilityOffline,
		DetectionState:            RuntimeDetectionStateMissing,
		HasAssignedAgent:          false,
		RequiresExperimentalOptIn: false,
		Models:                    []RuntimeModelRecord{},
	})
	if err != nil {
		t.Fatalf("marshal runtime record: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode runtime record: %v", err)
	}
	if value, ok := decoded["provider_version"]; !ok || value != "" {
		t.Fatalf("provider_version = %#v, present=%v, body=%s", value, ok, body)
	}
}
