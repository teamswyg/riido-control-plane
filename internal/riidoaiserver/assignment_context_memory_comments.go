package riidoaiserver

import (
	"context"
	"strings"
)

func (m *MemoryAssignmentContextStore) SaveAssignmentContextComments(ctx context.Context, assignmentID string, pages [][]AIAgentTaskContextComment) error {
	if expired, err := m.expirePayloads(ctx, assignmentID); err != nil {
		return err
	} else if expired {
		return errAssignmentContextExpired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.comments == nil {
		m.comments = map[string][][]AIAgentTaskContextComment{}
	}
	m.comments[strings.TrimSpace(assignmentID)] = pages
	return nil
}

func (m *MemoryAssignmentContextStore) LoadAssignmentContextCommentPage(ctx context.Context, assignmentID string, page int) ([]AIAgentTaskContextComment, bool, error) {
	if expired, err := m.expirePayloads(ctx, assignmentID); expired || err != nil {
		return nil, false, err
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	pages := m.comments[strings.TrimSpace(assignmentID)]
	if page < 0 || page >= len(pages) {
		return nil, false, nil
	}
	return pages[page], true, nil
}
