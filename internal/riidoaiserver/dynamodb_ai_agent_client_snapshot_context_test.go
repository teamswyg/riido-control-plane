package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotRepliesCanceledContextBeforeCommandSend(t *testing.T) {
	store := &DynamoDBAIAgentClientSnapshot{
		commands: make(chan dynamoDBAIAgentClientSnapshotCommand),
		done:     make(chan struct{}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := store.LoadAIAgentClientSnapshot(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("LoadAIAgentClientSnapshot canceled error = %v", err)
	}
	if err := store.SaveAIAgentClientSnapshot(ctx, AIAgentClientSnapshot{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveAIAgentClientSnapshot canceled error = %v", err)
	}
}
