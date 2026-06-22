package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func decodeCloudWatchEMFEnvelope(t *testing.T, body []byte) cloudWatchEMFEnvelope {
	t.Helper()
	var envelope cloudWatchEMFEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode emf: %v\n%s", err, string(body))
	}
	return envelope
}

func cloudWatchEMFMetricUnits(envelope cloudWatchEMFEnvelope) map[string]string {
	metricUnits := map[string]string{}
	for _, spec := range envelope.AWS.CloudWatchMetrics[0].Metrics {
		metricUnits[spec.Name] = spec.Unit
	}
	return metricUnits
}
