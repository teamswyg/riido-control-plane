package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

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
