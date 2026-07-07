package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForAssignmentWithOperationStoreTimesOut(t *testing.T) {
	store := NewStoreWithConfig(StoreConfig{
		OperationStore: &runtimeFakeAssignmentOperationStore{},
	})
	defer store.Close()

	start := time.Now()
	resp, err := store.WaitForAssignment(context.Background(), "agent-a",
		daemonPollRequest(), 80*time.Millisecond, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForAssignment: %v", err)
	}
	if resp.Action != PollNone {
		t.Fatalf("action = %s, want none", resp.Action)
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("operation-store wait returned before the hold budget")
	}
}

func TestWaitForAssignmentWithOperationStoreContextCancel(t *testing.T) {
	store := NewStoreWithConfig(StoreConfig{
		OperationStore: &runtimeFakeAssignmentOperationStore{},
	})
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		_, err := store.WaitForAssignment(ctx, "agent-a", daemonPollRequest(), time.Second, time.Second)
		errs <- err
	}()
	time.Sleep(30 * time.Millisecond)
	cancel()
	if err := <-errs; !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitForAssignment err = %v, want context.Canceled", err)
	}
}
