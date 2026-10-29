package riidoaiserver

import (
	"context"
	"errors"
)

func (s Server) assignmentContentView(ctx context.Context, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest) (any, []AssignmentContextSource, bool, error) {
	index := 0
	for _, part := range snapshot.Content {
		for i, cursor := range part.Cursors {
			if req.Cursor != cursor || req.Target != part.SourceID {
				index++
				continue
			}
			store, ok := s.config.AssignmentContext.(AssignmentContentStore)
			if !ok {
				return nil, nil, false, errors.New("assignment content store unavailable")
			}
			page, found, err := store.LoadAssignmentContentPage(ctx, snapshot.AssignmentID, index)
			if err != nil {
				return nil, nil, false, err
			}
			if !found || page.Cursor != cursor || page.ResourceKind != part.ResourceKind || page.SourceID != part.SourceID || page.Part != part.Part || page.Version != part.Version {
				return nil, nil, false, errors.New("assignment content page unavailable")
			}
			next := ""
			if i+1 < len(part.Cursors) {
				next = part.Cursors[i+1]
			}
			return struct {
				AssignmentContentPage
				NextCursor string `json:"next_cursor,omitempty"`
			}{page, next}, page.Sources, true, nil
		}
	}
	return nil, nil, false, nil
}
