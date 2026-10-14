package riidoaiserver

import "testing"

func assertSnapshotMetrics(t *testing.T, metrics *AIAgentClientPersistenceMetrics) {
	t.Helper()
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if snapshot.AIAgentClientSnapshotSaveCallsTotal != 1 ||
		snapshot.AIAgentClientSnapshotLoadCallsTotal != 1 {
		t.Fatalf("snapshot persistence call metrics = %+v", snapshot)
	}
	if snapshot.AIAgentClientSnapshotSaveBytesLast <= 0 ||
		snapshot.AIAgentClientSnapshotLoadBytesLast <= 0 {
		t.Fatalf("snapshot persistence byte metrics = %+v", snapshot)
	}
	if snapshot.AIAgentClientSnapshotSaveLatencySamplesTotal != 1 ||
		snapshot.AIAgentClientSnapshotLoadLatencySamplesTotal != 1 {
		t.Fatalf("snapshot persistence latency metrics = %+v", snapshot)
	}
}
