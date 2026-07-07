package riidoaiserver

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestPublishCloudWatchEMFUsesStoreSnapshot(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()
	ctx := context.Background()
	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	var buf bytes.Buffer
	config := CloudWatchEMFConfig{Writer: &buf, Namespace: "Riido/Test", ServiceName: "test-service"}
	if err := PublishCloudWatchEMF(ctx, store, config); err != nil {
		t.Fatalf("PublishCloudWatchEMF: %v", err)
	}
	envelope := decodeCloudWatchEMFEnvelope(t, buf.Bytes())
	if envelope.AWS.CloudWatchMetrics[0].Namespace != "Riido/Test" || envelope.Service != "test-service" {
		t.Fatalf("emf identity = %+v", envelope)
	}
	if envelope.AssignmentsTotal != 1 || envelope.TaskEventsTotal != 1 {
		t.Fatalf("store counters = %+v", envelope)
	}
}

func TestRunCloudWatchEMFPublisherRejectsNonPositiveInterval(t *testing.T) {
	err := RunCloudWatchEMFPublisher(context.Background(), nil, 0, CloudWatchEMFConfig{})
	if err == nil || err.Error() != "riidoaiserver: metrics interval must be positive" {
		t.Fatalf("RunCloudWatchEMFPublisher() error = %v", err)
	}
}

func TestRunCloudWatchEMFPublisherStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	metrics := metricsReaderFunc(func(context.Context) (MetricsSnapshot, error) {
		calls++
		return MetricsSnapshot{}, nil
	})
	var buf bytes.Buffer
	err := RunCloudWatchEMFPublisher(ctx, metrics, time.Hour, CloudWatchEMFConfig{Writer: &buf})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunCloudWatchEMFPublisher() error = %v, want context.Canceled", err)
	}
	if calls != 1 || buf.Len() == 0 {
		t.Fatalf("publisher calls=%d output=%q", calls, buf.String())
	}
}
