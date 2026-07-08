package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFileOutboxAppendRespectsCanceledContextAfterSend(t *testing.T) {
	outbox := &FileOutbox{
		commands: make(chan outboxCommand, 1),
		done:     make(chan struct{}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- outbox.AppendTaskEvent(ctx, TaskEvent{})
	}()
	waitForOutboxCommand(t, outbox.commands)
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("AppendTaskEvent receive canceled error = %v", err)
	}
}

func TestFileOutboxAppendReportsClosedAfterSend(t *testing.T) {
	outbox := &FileOutbox{
		commands: make(chan outboxCommand, 1),
		done:     make(chan struct{}),
	}
	result := make(chan error, 1)
	go func() {
		result <- outbox.AppendTaskEvent(context.Background(), TaskEvent{})
	}()
	waitForOutboxCommand(t, outbox.commands)
	close(outbox.done)
	if err := <-result; err == nil || err.Error() != "riidoaiserver: outbox closed" {
		t.Fatalf("AppendTaskEvent receive closed error = %v", err)
	}
}

func waitForOutboxCommand(t *testing.T, commands <-chan outboxCommand) {
	t.Helper()
	select {
	case <-commands:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for outbox command")
	}
}
