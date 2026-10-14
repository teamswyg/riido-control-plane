package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestWaitForAssignmentWithOperationStoreWakesOnSignal(t *testing.T) {
	store := NewStoreWithConfig(StoreConfig{
		OperationStore: &runtimeFakeAssignmentOperationStore{},
	})
	defer store.Close()

	done := make(chan PollResponse, 1)
	errs := make(chan error, 1)
	go func() {
		resp, err := store.WaitForAssignment(context.Background(), "agent-a",
			daemonPollRequest(), time.Second, time.Second)
		if err != nil {
			errs <- err
			return
		}
		done <- resp
	}()
	time.Sleep(30 * time.Millisecond)
	if _, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-a", RuntimeProvider: "codex", Prompt: "go",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	select {
	case resp := <-done:
		if resp.Action != PollStart {
			t.Fatalf("action = %s, want start", resp.Action)
		}
	case err := <-errs:
		t.Fatalf("WaitForAssignment: %v", err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("operation-store wait did not wake on assignment signal")
	}
}
