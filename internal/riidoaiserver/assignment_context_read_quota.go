package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

const assignmentContextReadLimit = 100

// ConsumeAssignmentContextRead charges an authenticated active-lease attempt,
// including retries and refusals. The caller validates the lease first. The
// durable assignment ID survives lease renewal and runtime run-label changes;
// a new user follow-up execution gets a new assignment and its own budget.
func (m *MemoryAssignmentContextStore) ConsumeAssignmentContextRead(ctx context.Context, assignmentID string) (bool, error) {
	id := strings.TrimSpace(assignmentID)
	if id == "" {
		return false, errors.New("assignment_id is required")
	}
	if expired, err := m.expirePayloads(ctx, id); err != nil {
		return false, err
	} else if expired {
		return false, errAssignmentContextExpired
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if m.readAttempts == nil {
		m.readAttempts = map[string]int{}
	}
	if m.readAttempts[id] >= assignmentContextReadLimit || m.expiry[id] > 0 {
		return false, nil
	}
	m.readAttempts[id]++
	return true, nil
}
