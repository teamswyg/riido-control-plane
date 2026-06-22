package riidoaiserver

import (
	"bytes"
	"testing"
)

func TestWriteCloudWatchEMF(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCloudWatchEMF(&buf, CloudWatchEMFConfig{}, sampleCloudWatchEMFSnapshot()); err != nil {
		t.Fatalf("WriteCloudWatchEMF: %v", err)
	}
	envelope := decodeCloudWatchEMFEnvelope(t, buf.Bytes())
	assertCloudWatchEMFIdentity(t, envelope)
	assertCloudWatchEMFScopes(t, envelope)
	assertCloudWatchEMFCounters(t, envelope)
	assertCloudWatchEMFMetricUnits(t, envelope)
}
