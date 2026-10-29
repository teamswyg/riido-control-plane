package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// MemoryAssignmentContextStore keeps snapshots and read records in process.
// It backs local development and tests; production uses DynamoDB.
type MemoryAssignmentContextStore struct {
	retentionReader AssignmentProjectionReader
	expiry          map[string]int64
	now             func() time.Time
	mu              sync.Mutex
	snapshots       map[string]AssignmentContextSnapshot
	reads           map[string][]AssignmentContextRead
	comments        map[string][][]AIAgentTaskContextComment
	content         map[string][]AssignmentContentPage
	readAttempts    map[string]int
}

func NewMemoryAssignmentContextStore() *MemoryAssignmentContextStore {
	return &MemoryAssignmentContextStore{
		now:       time.Now,
		snapshots: map[string]AssignmentContextSnapshot{},
		reads:     map[string][]AssignmentContextRead{},
	}
}

func (m *MemoryAssignmentContextStore) SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error {
	if snapshot.Content != nil || snapshot.ContentBlocks != nil {
		return errAssignmentContentConflict
	}
	if expired, err := m.expirePayloads(ctx, snapshot.AssignmentID); err != nil {
		return err
	} else if expired {
		return errAssignmentContextExpired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	id := strings.TrimSpace(snapshot.AssignmentID)
	if id == "" {
		return errors.New("assignment_id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.snapshots[id].Content != nil {
		return errAssignmentContentConflict
	}
	m.snapshots[id] = snapshot
	return nil
}
