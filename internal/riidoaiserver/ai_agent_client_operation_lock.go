package riidoaiserver

import (
	"context"
	"sync"
)

// Inventory preflight and snapshot HTTP calls must not outlive their caller
// while queued behind an unrelated persistent snapshot operation.
func (s *PersistentAIAgentClientStore) lockOperationContext(ctx context.Context) error {
	return s.operationMu.lockContext(ctx)
}

// All callers wait on the same channel. A TryLock polling caller cannot join
// sync.Mutex's waiter queue and can starve while ordinary readers keep arriving.
// The zero value is ready for use, like sync.Mutex.
type snapshotOperationMutex struct {
	once  sync.Once
	ready chan struct{}
}

func (m *snapshotOperationMutex) initialize() {
	m.once.Do(func() {
		m.ready = make(chan struct{}, 1)
		m.ready <- struct{}{}
	})
}

func (m *snapshotOperationMutex) Lock() {
	m.initialize()
	<-m.ready
}

func (m *snapshotOperationMutex) Unlock() {
	m.ready <- struct{}{}
}

func (m *snapshotOperationMutex) lockContext(ctx context.Context) error {
	m.initialize()
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ready:
		if err := ctx.Err(); err != nil {
			m.Unlock()
			return err
		}
		return nil
	}
}
