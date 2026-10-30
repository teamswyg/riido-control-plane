package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

type assignmentContentBatch struct {
	ResourceKind AssignmentContentKind   `json:"resource_kind"`
	SourceID     string                  `json:"source_id"`
	Part         AssignmentContentPart   `json:"part"`
	Version      string                  `json:"version"`
	Pages        []AssignmentContentPage `json:"pages"`
	Complete     bool                    `json:"complete"`
	NextCursor   string                  `json:"next_cursor,omitempty"`
}

func (s Server) assignmentContentBatchView(ctx context.Context, store AssignmentContentStore, snapshot AssignmentContextSnapshot, part AssignmentContentManifest, position, index, maxBytes int) (any, []AssignmentContextSource, error) {
	out := assignmentContentBatch{ResourceKind: part.ResourceKind, SourceID: part.SourceID, Part: part.Part, Version: part.Version, Pages: []AssignmentContentPage{}}
	size := 0
	var invalid error
	visit := func(page AssignmentContentPage) bool {
		at := position + len(out.Pages)
		if at >= len(part.Cursors) || page.Cursor != part.Cursors[at] || page.ResourceKind != part.ResourceKind || page.SourceID != part.SourceID || page.Part != part.Part || page.Version != part.Version {
			invalid = errors.New("assignment content page unavailable")
			return false
		}
		encoded, err := json.Marshal(page)
		if err != nil {
			invalid = err
			return false
		}
		if len(out.Pages) > 0 && size+len(encoded) > maxBytes {
			return false
		}
		out.Pages = append(out.Pages, page)
		size += len(encoded)
		return true
	}
	count := len(part.Cursors) - position
	if loader, ok := store.(AssignmentContentRangeLoader); ok {
		if err := loader.VisitAssignmentContentPages(ctx, snapshot.AssignmentID, index, count, visit); err != nil {
			return nil, nil, err
		}
	} else {
		for i := range count {
			page, found, err := store.LoadAssignmentContentPage(ctx, snapshot.AssignmentID, index+i)
			if err != nil {
				return nil, nil, err
			}
			if !found {
				invalid = errors.New("assignment content page unavailable")
			}
			if !found || !visit(page) {
				break
			}
		}
	}
	if invalid != nil && len(out.Pages) == 0 {
		return nil, nil, invalid
	}
	if invalid != nil {
		// A damaged later page ends the batch; its cursor still reports the failure.
		invalid = nil
	}
	last := out.Pages[len(out.Pages)-1]
	out.Complete = last.Complete
	if next := position + len(out.Pages); next < len(part.Cursors) {
		out.NextCursor = part.Cursors[next]
	}
	return out, uniqueAssignmentContentSources(out.Pages), nil
}
