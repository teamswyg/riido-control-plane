package riidoaiserver

import (
	"context"
	"time"
)

// Inventory preflight and snapshot HTTP calls must not outlive their caller
// while queued behind an unrelated persistent snapshot operation.
func (s *PersistentAIAgentClientStore) lockOperationContext(ctx context.Context) error {
	timer := time.NewTicker(5 * time.Millisecond)
	defer timer.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if s.operationMu.TryLock() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
}
