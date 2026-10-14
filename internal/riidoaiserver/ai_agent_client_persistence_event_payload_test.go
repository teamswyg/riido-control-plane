package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRestoreSnapshotEventPayloadTypes(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		wantType  string
	}{
		{
			name:      "device runtime snapshot",
			eventType: AgentClientEventDeviceRuntimeSnapshot,
			wantType:  "riidoaiserver.DeviceRuntimeSnapshotEvent",
		},
		{
			name:      "device daemon status",
			eventType: AgentClientEventDeviceDaemonStatus,
			wantType:  "riidoaiserver.DeviceDaemonStatusEvent",
		},
		{
			name:      "editability changed",
			eventType: AgentClientEventEditabilityChanged,
			wantType:  "riidoaiserver.AgentEditabilityChangedEvent",
		},
		{
			name:      "work status changed",
			eventType: AgentClientEventWorkStatusChanged,
			wantType:  "riidoaiserver.AgentWorkStatusChangedEvent",
		},
		{
			name:      "thread progress",
			eventType: AgentClientEventThreadProgress,
			wantType:  "riidoaiserver.AgentThreadProgressEvent",
		},
		{
			name:      "unknown event",
			eventType: "custom.event",
			wantType:  "map[string]interface {}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := restoreSnapshotEventPayload(tc.eventType, json.RawMessage(`{}`))
			if err != nil {
				t.Fatalf("restoreSnapshotEventPayload: %v", err)
			}
			if got := typeName(payload); got != tc.wantType {
				t.Fatalf("payload type = %s, want %s", got, tc.wantType)
			}
		})
	}
}

func TestRestoreSnapshotEventPayloadRejectsInvalidJSON(t *testing.T) {
	_, err := restoreSnapshotEventPayload(AgentClientEventThreadProgress, json.RawMessage(`{`))
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func typeName(v any) string {
	return fmt.Sprintf("%T", v)
}
