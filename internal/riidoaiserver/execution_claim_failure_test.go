package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type failingClaimSnapshot struct {
	SnapshotStore
	fail bool
}

func (s *failingClaimSnapshot) SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error {
	if s.fail {
		return errors.New("fixture persist failure")
	}
	return s.SnapshotStore.SaveStoreSnapshot(ctx, snapshot)
}

func TestExecutionClaimPersistenceFailureCannotDispatchOrPoisonRetry(t *testing.T) {
	snapshot, err := NewFileStoreSnapshot(t.TempDir() + "/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	failing := &failingClaimSnapshot{SnapshotStore: snapshot}
	actor := NewStoreWithConfig(StoreConfig{SnapshotStore: failing})
	defer actor.Close()
	if _, err := actor.AssignTask(context.Background(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"}); err != nil {
		t.Fatal(err)
	}
	_, events, cancel, err := actor.SubscribeTask(context.Background(), "task")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"device", "owner"})
	failing.fail = true
	poll, err := actor.PollAgent(ctx, "agent", PollRequest{})
	if err == nil || poll.Action == PollStart || poll.Assignment != nil {
		t.Fatal("failed claim dispatched")
	}
	select {
	case <-events:
		t.Fatal("failed claim published event")
	default:
	}
	observed, err := actor.observeDeviceExecutions(context.Background(), "device")
	if err != nil || observed != deviceExecutionUnknown {
		t.Fatal("uncommitted claim entered actor state")
	}
	failing.fail = false
	poll, err = actor.PollAgent(ctx, "agent", PollRequest{})
	if err != nil || poll.Action != PollStart {
		t.Fatalf("retry did not claim queued assignment: %v", err)
	}
	select {
	case <-events:
	default:
		t.Fatal("successful retry did not publish")
	}
}
