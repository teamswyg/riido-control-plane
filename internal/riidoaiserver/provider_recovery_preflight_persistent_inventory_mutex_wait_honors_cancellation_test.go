package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPersistentInventoryMutexWaitHonorsCancellation(t *testing.T) {
	store := &PersistentAIAgentClientStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore()}
	store.operationMu.Lock()
	defer store.operationMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := store.ListAIAgentTaskAssignableAgents(ctx, AuthorizationResult{}, "task"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("mutex wait=%v", err)
	}
}
