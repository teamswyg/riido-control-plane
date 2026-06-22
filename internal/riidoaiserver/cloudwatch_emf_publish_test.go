package riidoaiserver

import (
	"bytes"
	"context"
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
