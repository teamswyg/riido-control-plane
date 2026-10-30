package riidoaiserver

import "context"

func (m *MemoryAssignmentContextStore) VisitAssignmentContentPages(ctx context.Context, id string, first, count int, visit func(AssignmentContentPage) bool) error {
	for i := range count {
		page, found, err := m.LoadAssignmentContentPage(ctx, id, first+i)
		if err != nil {
			return err
		}
		if !found || !visit(page) {
			return nil
		}
	}
	return nil
}
