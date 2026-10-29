package riidoaiserver

import (
	"context"
	"encoding/json"
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
	if ok && snapshot.Content != nil {
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			return AssignmentContextSnapshot{}, false, err
		}
		var cloned AssignmentContextSnapshot
		if err := json.Unmarshal(encoded, &cloned); err != nil {
			return AssignmentContextSnapshot{}, false, err
		}
		return cloned, true, nil
	}
	return snapshot, ok, nil
}
