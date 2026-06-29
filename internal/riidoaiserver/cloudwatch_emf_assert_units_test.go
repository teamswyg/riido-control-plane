package riidoaiserver

import "testing"

func assertCloudWatchEMFMetricUnits(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	metricUnits := cloudWatchEMFMetricUnits(envelope)
	if metricUnits["event_append_latency_samples_total"] != "Count" ||
		metricUnits["event_append_latency_max_ms"] != "Milliseconds" {
		t.Fatalf("emf latency metric units = %+v", metricUnits)
	}
	if metricUnits["http_requests_total"] != "Count" ||
		metricUnits["http_request_latency_max_ms"] != "Milliseconds" {
		t.Fatalf("emf http metric units = %+v", metricUnits)
	}
	if metricUnits["http_requests_daemon_total"] != "Count" ||
		metricUnits["http_requests_client_app_total"] != "Count" {
		t.Fatalf("emf http surface metric units = %+v", metricUnits)
	}
	if metricUnits["store_operation_calls_total"] != "Count" ||
		metricUnits["store_operation_latency_max_ms"] != "Milliseconds" {
		t.Fatalf("emf store metric units = %+v", metricUnits)
	}
	if metricUnits["ai_agent_client_snapshot_load_bytes_last"] != "Bytes" ||
		metricUnits["ai_agent_client_snapshot_save_latency_max_ms"] != "Milliseconds" {
		t.Fatalf("emf AI Agent client snapshot metric units = %+v", metricUnits)
	}
}
