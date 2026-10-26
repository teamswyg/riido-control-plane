package riidoaiserver

import (
	"context"
	"strings"
)

func (m *MemoryAssignmentContextStore) LoadAssignmentContextSnapshot(ctx context.Context, assignmentID string) (AssignmentContextSnapshot, bool, error) {
	if expired, err := m.expirePayloads(ctx, assignmentID); expired || err != nil {
		return AssignmentContextSnapshot{}, false, err
	}
	if err := ctx.Err(); err != nil {
		return AssignmentContextSnapshot{}, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot, ok := m.snapshots[strings.TrimSpace(assignmentID)]
	return snapshot, ok, nil
}
