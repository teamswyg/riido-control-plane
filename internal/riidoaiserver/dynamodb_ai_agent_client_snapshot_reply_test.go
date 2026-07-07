package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDynamoDBSnapshotWaitLoadReplyBranches(t *testing.T) {
	store := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{})}
	reply := make(chan dynamoDBAIAgentClientSnapshotLoadResult, 1)
	reply <- dynamoDBAIAgentClientSnapshotLoadResult{
		snapshot: AIAgentClientSnapshot{SchemaVersion: "schema"},
		ok:       true,
		err:      errors.New("load"),
	}
	snapshot, ok, err := store.waitLoadReply(context.Background(), reply)
	if snapshot.SchemaVersion != "schema" || !ok || err == nil {
		t.Fatalf("load reply = snapshot %+v ok %v err %v", snapshot, ok, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, ok, err = store.waitLoadReply(ctx, make(chan dynamoDBAIAgentClientSnapshotLoadResult))
	if ok || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled load reply = ok %v err %v", ok, err)
	}

	closed := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{})}
	close(closed.done)
	_, ok, err = closed.waitLoadReply(context.Background(), make(chan dynamoDBAIAgentClientSnapshotLoadResult))
	if ok || err.Error() != errDynamoDBAIAgentClientSnapshotClosed().Error() {
		t.Fatalf("closed load reply = ok %v err %v", ok, err)
	}
}

func TestDynamoDBSnapshotWaitErrorReplyBranches(t *testing.T) {
	store := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{})}
	reply := make(chan error, 1)
	reply <- errors.New("save")
	if err := store.waitErrorReply(context.Background(), reply); err == nil {
		t.Fatal("save reply error missing")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.waitErrorReply(ctx, make(chan error)); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled save reply err = %v", err)
	}

	closed := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{})}
	close(closed.done)
	if err := closed.waitErrorReply(context.Background(), make(chan error)); err.Error() != errDynamoDBAIAgentClientSnapshotClosed().Error() {
		t.Fatalf("closed save reply err = %v", err)
	}
}
