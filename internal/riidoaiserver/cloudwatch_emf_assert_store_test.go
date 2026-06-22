package riidoaiserver

import "testing"

func assertCloudWatchEMFStoreOperationCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.StoreOperationCallsTotal != 29 || envelope.StoreOperationErrorsTotal != 3 ||
		envelope.StoreOperationLatencySamplesTotal != 29 || envelope.StoreOperationLatencyMaxMilliseconds != 78 {
		t.Fatalf("emf store operation aggregate = %+v", envelope)
	}
	if len(envelope.StoreOperations) != 1 ||
		envelope.StoreOperations[0].Operation != StoreOperationPollAssignment.String() {
		t.Fatalf("emf store operation breakdown = %+v", envelope.StoreOperations)
	}
}

func assertCloudWatchEMFSnapshotCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.AIAgentClientSnapshotLoadCallsTotal != 31 ||
		envelope.AIAgentClientSnapshotLoadBytesLast != 40960 ||
		envelope.AIAgentClientSnapshotLoadLatencySamplesTotal != 31 ||
		envelope.AIAgentClientSnapshotLoadLatencyMaxMilliseconds != 81 {
		t.Fatalf("emf AI Agent client snapshot load metrics = %+v", envelope)
	}
	if envelope.AIAgentClientSnapshotSaveCallsTotal != 37 ||
		envelope.AIAgentClientSnapshotSaveBytesLast != 20480 ||
		envelope.AIAgentClientSnapshotSaveLatencySamplesTotal != 37 ||
		envelope.AIAgentClientSnapshotSaveLatencyMaxMilliseconds != 82 {
		t.Fatalf("emf AI Agent client snapshot save metrics = %+v", envelope)
	}
}
