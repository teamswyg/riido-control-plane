package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (m *MemoryAssignmentContextStore) SaveAssignmentContentBundle(ctx context.Context, snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) error {
	if err := validateAssignmentContentBundle(snapshot, pages); err != nil {
		return err
	}
	if expired, err := m.expirePayloads(ctx, snapshot.AssignmentID); err != nil {
		return err
	} else if expired {
		return errAssignmentContextExpired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	cloned, err := cloneAssignmentContentBundle(snapshot, pages)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if previous, exists := m.snapshots[snapshot.AssignmentID]; exists {
		return equalAssignmentContentBundle(previous, m.content[snapshot.AssignmentID], snapshot, pages)
	}
	if len(m.comments[snapshot.AssignmentID]) > 0 || len(m.content[snapshot.AssignmentID]) > 0 {
		return errAssignmentContentConflict
	}
	if m.content == nil {
		m.content = map[string][]AssignmentContentPage{}
	}
	m.snapshots[snapshot.AssignmentID], m.content[snapshot.AssignmentID] = cloned.Snapshot, cloned.Pages
	return nil
}

func (m *MemoryAssignmentContextStore) LoadAssignmentContentPage(ctx context.Context, id string, index int) (AssignmentContentPage, bool, error) {
	if expired, err := m.expirePayloads(ctx, id); expired || err != nil {
		return AssignmentContentPage{}, false, err
	}
	if err := ctx.Err(); err != nil {
		return AssignmentContentPage{}, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	pages := m.content[id]
	if index < 0 || index >= len(pages) {
		return AssignmentContentPage{}, false, nil
	}
	encoded, err := json.Marshal(pages[index])
	if err != nil {
		return AssignmentContentPage{}, false, err
	}
	var page AssignmentContentPage
	err = json.Unmarshal(encoded, &page)
	return page, err == nil, err
}
