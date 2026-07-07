package riidoaiserver

import (
	"strings"
	"testing"
)

func TestSnapshotEventsReportsEventTypeOnMarshalError(t *testing.T) {
	_, err := snapshotEvents([]ClientStreamEvent{{
		EventType: AgentClientEventWorkStatusChanged,
		Payload:   func() {},
	}})
	if err == nil {
		t.Fatal("expected marshal error")
	}
	if !strings.Contains(err.Error(), AgentClientEventWorkStatusChanged) {
		t.Fatalf("error = %v, want event type", err)
	}
}

func TestRestoreSnapshotEventsReportsEventTypeOnPayloadError(t *testing.T) {
	_, err := restoreSnapshotEvents([]AIAgentClientEventSnapshot{{
		Seq:       7,
		EventType: AgentClientEventThreadProgress,
		Payload:   []byte(`{`),
	}})
	if err == nil {
		t.Fatal("expected restore error")
	}
	if !strings.Contains(err.Error(), AgentClientEventThreadProgress) {
		t.Fatalf("error = %v, want event type", err)
	}
}

func TestSnapshotEventsRoundTripsKnownPayload(t *testing.T) {
	snapshots, err := snapshotEvents([]ClientStreamEvent{{
		Seq:       3,
		EventType: AgentClientEventWorkStatusChanged,
		Payload: AgentWorkStatusChangedEvent{
			EventType:     AgentClientEventWorkStatusChanged,
			SchemaVersion: SchemaVersion,
			AgentID:       "agent-codex",
			WorkStatus:    AgentWorkStatusRunning,
		},
	}})
	if err != nil {
		t.Fatalf("snapshotEvents: %v", err)
	}
	events, err := restoreSnapshotEvents(snapshots)
	if err != nil {
		t.Fatalf("restoreSnapshotEvents: %v", err)
	}
	if events[0].Seq != 3 {
		t.Fatalf("seq = %d, want 3", events[0].Seq)
	}
	if _, ok := events[0].Payload.(AgentWorkStatusChangedEvent); !ok {
		t.Fatalf("payload type = %T", events[0].Payload)
	}
}
