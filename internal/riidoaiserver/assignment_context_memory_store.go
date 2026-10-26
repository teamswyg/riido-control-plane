package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// MemoryAssignmentContextStore keeps snapshots and read records in process.
// It backs local development and tests; production uses DynamoDB.
type MemoryAssignmentContextStore struct {
	mu        sync.Mutex
	snapshots map[string]AssignmentContextSnapshot
	reads     map[string][]AssignmentContextRead
}

func NewMemoryAssignmentContextStore() *MemoryAssignmentContextStore {
	return &MemoryAssignmentContextStore{
		snapshots: map[string]AssignmentContextSnapshot{},
		reads:     map[string][]AssignmentContextRead{},
	}
}

func (m *MemoryAssignmentContextStore) SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error {
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

func (m *MemoryAssignmentContextStore) LoadAssignmentContextSnapshot(ctx context.Context, assignmentID string) (AssignmentContextSnapshot, bool, error) {
	if err := ctx.Err(); err != nil {
		return AssignmentContextSnapshot{}, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot, ok := m.snapshots[strings.TrimSpace(assignmentID)]
	return snapshot, ok, nil
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
