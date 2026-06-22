package riidoaiserver

import "testing"

func assertCloudWatchEMFCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	assertCloudWatchEMFStoreStateCounters(t, envelope)
	assertCloudWatchEMFEventCounters(t, envelope)
	assertCloudWatchEMFHTTPCounters(t, envelope)
	assertCloudWatchEMFStoreOperationCounters(t, envelope)
	assertCloudWatchEMFSnapshotCounters(t, envelope)
}

func assertCloudWatchEMFStoreStateCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.TasksTotal != 2 || envelope.AssignmentsQueued != 1 || envelope.AssignmentsRunning != 2 {
		t.Fatalf("emf store state counters = %+v", envelope)
	}
}

func assertCloudWatchEMFEventCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.PollStartTotal != 2 || envelope.SSESubscribers != 13 || envelope.OutboxErrorsTotal != 17 {
		t.Fatalf("emf event counters = %+v", envelope)
	}
	if envelope.EventAppendLatencySamplesTotal != 19 || envelope.EventAppendLatencyMaxMilliseconds != 89 {
		t.Fatalf("emf event latency counters = %+v", envelope)
	}
}

func assertCloudWatchEMFHTTPCounters(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.HTTPRequestsTotal != 23 || envelope.HTTPResponse2xxTotal != 17 ||
		envelope.HTTPResponse4xxTotal != 4 || envelope.HTTPResponse5xxTotal != 2 {
		t.Fatalf("emf http counters = %+v", envelope)
	}
	if envelope.HTTPRequestLatencySamplesTotal != 23 || envelope.HTTPRequestLatencyMaxMilliseconds != 67 ||
		len(envelope.HTTPTransactions) != 1 || envelope.HTTPTransactions[0].Route != "/healthz" {
		t.Fatalf("emf http latency/routes = %+v", envelope)
	}
}
