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
}

func NewMemoryAssignmentContextStore() *MemoryAssignmentContextStore {
	return &MemoryAssignmentContextStore{
		now:       time.Now,
		snapshots: map[string]AssignmentContextSnapshot{},
		reads:     map[string][]AssignmentContextRead{},
	}
}

func (m *MemoryAssignmentContextStore) SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error {
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
	m.snapshots[id] = snapshot
	return nil
}

func (m *MemoryAssignmentContextStore) AppendAssignmentContextReads(ctx context.Context, reads []AssignmentContextRead) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, read := range reads {
		if strings.TrimSpace(read.AssignmentID) == "" {
			return errors.New("assignment_id is required")
		}
		m.reads[read.AssignmentID] = append(m.reads[read.AssignmentID], read)
	}
	return nil
}

func (m *MemoryAssignmentContextStore) ListAssignmentContextReads(ctx context.Context, assignmentID string) ([]AssignmentContextRead, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]AssignmentContextRead(nil), m.reads[strings.TrimSpace(assignmentID)]...), nil
}
