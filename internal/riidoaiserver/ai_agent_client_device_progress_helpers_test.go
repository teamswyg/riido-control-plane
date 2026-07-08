package riidoaiserver

import (
	"testing"
	"time"
)

func TestDeviceRuntimeSnapshotStale(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	if deviceRuntimeSnapshotStale(time.Time{}, now) {
		t.Fatalf("zero last-seen timestamp should not be stale")
	}
	fresh := now.Add(-AIAgentDeviceRuntimeSnapshotStaleAfter)
	if deviceRuntimeSnapshotStale(fresh, now) {
		t.Fatalf("snapshot at stale threshold should remain fresh")
	}
	stale := now.Add(-AIAgentDeviceRuntimeSnapshotStaleAfter - time.Second)
	if !deviceRuntimeSnapshotStale(stale, now) {
		t.Fatalf("snapshot older than stale threshold should be stale")
	}
}

func TestEventDeviceRecord(t *testing.T) {
	want := DeviceRecord{DeviceID: "dev-1"}
	got, ok := eventDeviceRecord(DeviceRuntimeSnapshotEvent{Device: want})
	if !ok || got.DeviceID != want.DeviceID {
		t.Fatalf("eventDeviceRecord(snapshot) = %#v, %v", got, ok)
	}
	if got, ok := eventDeviceRecord("not-a-device-event"); ok || got.DeviceID != "" {
		t.Fatalf("eventDeviceRecord(other) = %#v, %v", got, ok)
	}
}

func TestAppendRetainedProgressLinesInPlaceDropsAllExisting(t *testing.T) {
	existing := []AgentThreadProgressLine{{Seq: 1}, {Seq: 2}}
	incoming := make([]AgentThreadProgressLine, aiAgentClientThreadProgressLineLimit+2)
	for i := range incoming {
		incoming[i] = AgentThreadProgressLine{Seq: 1000 + i}
	}
	got := appendRetainedProgressLinesInPlace(existing, incoming)
	if len(got) != aiAgentClientThreadProgressLineLimit {
		t.Fatalf("line count = %d, want %d", len(got), aiAgentClientThreadProgressLineLimit)
	}
	if got[0].Seq != 1002 || got[len(got)-1].Seq != 1001+aiAgentClientThreadProgressLineLimit {
		t.Fatalf("retained seq range = %d..%d", got[0].Seq, got[len(got)-1].Seq)
	}
}
