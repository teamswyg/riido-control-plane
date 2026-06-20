package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestStartMetricsPublisherWritesCloudWatchEMF(t *testing.T) {
	store := riidoaiserver.NewStoreWithClock(func() time.Time { return time.Unix(2000, 0).UTC() })
	defer store.Close()
	_, err := store.AssignTask(context.Background(), "task-a", riidoaiserver.AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "hello",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	writer := metricsCaptureWriter{ch: make(chan string, 1)}
	cancel, errCh := startMetricsPublisher(store, time.Hour, writer)
	defer stopMetricsPublisher(cancel, errCh)

	select {
	case body := <-writer.ch:
		if !strings.Contains(body, "\"_aws\"") || !strings.Contains(body, "\"assignments_total\":1") {
			t.Fatalf("metrics body = %s", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for metrics output")
	}
}
