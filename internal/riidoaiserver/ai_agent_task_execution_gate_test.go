package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestTaskAgentExecutionGateScopesAndReleasesCancelledWaiters(t *testing.T) {
	gate := newTaskAgentExecutionGates()
	release, err := gate.beginPublication(context.Background(), "task-a")
	if err != nil {
		t.Fatal(err)
	}
	other, err := gate.beginRemoval(context.Background(), "task-b")
	if err != nil {
		t.Fatal(err)
	}
	other()
	ctx, cancel := context.WithCancel(context.Background())
	waiting := make(chan error, 1)
	go func() { _, err := gate.beginRemoval(ctx, "task-a"); waiting <- err }()
	cancel()
	select {
	case err := <-waiting:
		if err != context.Canceled {
			t.Fatalf("cancel err=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled waiter blocked")
	}
	release()
	gate.mu.Lock()
	defer gate.mu.Unlock()
	if len(gate.tasks) != 0 {
		t.Fatalf("released gate retained tasks: %d", len(gate.tasks))
	}
}
