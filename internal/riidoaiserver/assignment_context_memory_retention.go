package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (m *MemoryAssignmentContextStore) bindRetention(reader AssignmentStore) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retentionReader, _ = reader.(AssignmentProjectionReader)
}

// The development store prunes payloads on access using the same terminal
// event clock. Durable physical removal is provided by DynamoDB TTL.
func (m *MemoryAssignmentContextStore) expirePayloads(ctx context.Context, id string) (bool, error) {
	id = strings.TrimSpace(id)
	m.mu.Lock()
	reader := m.retentionReader
	m.mu.Unlock()
	if reader == nil {
		return false, nil
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, id)
	if err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.expiry == nil {
		m.expiry = map[string]int64{}
	}
	if found && isTerminal(projection.Assignment.State) {
		terminalAt := projection.ContextTerminalAt
		if terminalAt.IsZero() {
			terminalAt = projection.Assignment.UpdatedAt
		}
		if terminalAt.IsZero() {
			return false, errors.New("terminal assignment timestamp missing")
		}
		at := terminalAt.Add(assignmentContextRetention).Unix()
		if m.expiry[id] == 0 || at < m.expiry[id] {
			m.expiry[id] = at
		}
	}
	if m.expiry[id] > 0 && m.now().Unix() >= m.expiry[id] {
		delete(m.snapshots, id)
		delete(m.comments, id)
		return true, nil
	}
	return false, nil
}
