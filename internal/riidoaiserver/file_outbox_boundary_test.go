package riidoaiserver

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestFileOutboxNilAndClosedBoundaries(t *testing.T) {
	var nilOutbox *FileOutbox
	var nilContext context.Context
	if err := nilOutbox.AppendTaskEvent(nilContext, TaskEvent{}); err != nil {
		t.Fatalf("nil AppendTaskEvent() error = %v, want nil", err)
	}
	if err := nilOutbox.Close(); err != nil {
		t.Fatalf("nil Close() error = %v, want nil", err)
	}

	outbox := newFileOutboxForBoundary(t)
	if err := outbox.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := outbox.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	err := outbox.AppendTaskEvent(context.Background(), TaskEvent{})
	if err == nil || err.Error() != "riidoaiserver: outbox closed" {
		t.Fatalf("AppendTaskEvent closed error = %v", err)
	}
}

func TestFileOutboxAppendRespectsCanceledContextBeforeSend(t *testing.T) {
	outbox := &FileOutbox{
		commands: make(chan outboxCommand),
		done:     make(chan struct{}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := outbox.AppendTaskEvent(ctx, TaskEvent{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("AppendTaskEvent canceled error = %v", err)
	}
}

func TestFileOutboxLoopRejectsNilEventCommand(t *testing.T) {
	outbox := newFileOutboxForBoundary(t)
	defer outbox.Close()
	reply := make(chan error, 1)
	outbox.commands <- outboxCommand{reply: reply}
	if err := <-reply; err == nil || err.Error() != "riidoaiserver: nil outbox event" {
		t.Fatalf("nil outbox command error = %v", err)
	}
}

func TestFileOutboxConstructorReportsFilesystemError(t *testing.T) {
	if _, err := NewFileOutbox(t.TempDir()); err == nil {
		t.Fatal("expected directory open error")
	}
}

func newFileOutboxForBoundary(t *testing.T) *FileOutbox {
	t.Helper()
	outbox, err := NewFileOutbox(filepath.Join(t.TempDir(), "outbox.jsonl"))
	if err != nil {
		t.Fatalf("NewFileOutbox: %v", err)
	}
	return outbox
}
