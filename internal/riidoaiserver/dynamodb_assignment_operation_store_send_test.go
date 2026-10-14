package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDynamoDBAssignmentOperationContext(t *testing.T) {
	var nilContext context.Context
	if got := dynamoDBAssignmentOperationContext(nilContext); got == nil {
		t.Fatal("nil context should fall back to background context")
	}
	ctx := context.Background()
	if got := dynamoDBAssignmentOperationContext(ctx); got != ctx {
		t.Fatal("non-nil context should be preserved")
	}
}

func TestDynamoDBAssignmentOperationStoreSend(t *testing.T) {
	store := &DynamoDBAssignmentOperationStore{
		commands: make(chan dynamoDBAssignmentOperationCommand, 1),
		done:     make(chan struct{}),
	}
	cmd := dynamoDBAssignmentOperationCommand{load: true}
	if err := store.send(context.Background(), cmd); err != nil {
		t.Fatalf("send active store: %v", err)
	}
	if got := <-store.commands; !got.load {
		t.Fatalf("send command load=%v, want true", got.load)
	}

	closed := &DynamoDBAssignmentOperationStore{done: make(chan struct{})}
	close(closed.done)
	if err := closed.send(context.Background(), cmd); err == nil {
		t.Fatal("send closed store should fail")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	blocked := &DynamoDBAssignmentOperationStore{done: make(chan struct{})}
	if err := blocked.send(ctx, cmd); !errors.Is(err, context.Canceled) {
		t.Fatalf("send canceled context error=%v, want context.Canceled", err)
	}
}

func TestDynamoDBAssignmentOperationStoreReceiveError(t *testing.T) {
	store := &DynamoDBAssignmentOperationStore{done: make(chan struct{})}
	reply := make(chan error, 1)
	want := errors.New("boom")
	reply <- want
	if err := store.receiveError(context.Background(), reply); !errors.Is(err, want) {
		t.Fatalf("receive reply error=%v, want %v", err, want)
	}

	closed := &DynamoDBAssignmentOperationStore{done: make(chan struct{})}
	close(closed.done)
	if err := closed.receiveError(context.Background(), nil); err == nil {
		t.Fatal("receive closed store should fail")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.receiveError(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("receive canceled context error=%v, want context.Canceled", err)
	}
}
