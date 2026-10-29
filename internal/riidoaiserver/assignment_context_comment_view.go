package riidoaiserver

import (
	"context"
	"strconv"
	"strings"
)

// AssignmentContextCommentPage is the `comments` view: one page of the
// task's comments in priority order (trigger, recent, older).
type AssignmentContextCommentPage struct {
	Comments   []AIAgentTaskContextComment `json:"comments"`
	Total      int                         `json:"total"`
	Status     string                      `json:"status,omitempty"`
	Cursor     string                      `json:"cursor"`
	NextCursor string                      `json:"next_cursor,omitempty"`
}

// assignmentContextViewData answers every view; only comments need the
// stored pages, so only that view touches the store.
func (s Server) assignmentContextViewData(ctx context.Context, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest) (any, []AssignmentContextSource, bool, error) {
	if err := validateAssignmentReadScopeManifest(snapshot); err != nil {
		return nil, nil, false, err
	}
	if req.View == AssignmentContextViewContent {
		return s.assignmentContentView(ctx, snapshot, req)
	}
	if snapshot.Context.ReadScope != nil {
		return s.assignmentReadScopeView(ctx, snapshot, req)
	}
	if req.View != AssignmentContextViewComments {
		data, sources, ok := selectAssignmentContextView(snapshot.Context, req.View, req.Target)
		return data, sources, ok, nil
	}
	c := snapshot.Context
	if target := strings.TrimSpace(req.Target); target != "" && !assignmentContextTargetMatches(target, c.Component.ID, c.Component.KeyNumber) {
		return nil, nil, false, nil
	}
	page, err := strconv.Atoi(strings.TrimSpace(req.Cursor))
	if err != nil {
		page = 0
	}
	out := AssignmentContextCommentPage{Cursor: strconv.Itoa(page), Comments: []AIAgentTaskContextComment{}}
	if c.Extended != nil {
		out.Total, out.Status = c.Extended.CommentTotal, c.Extended.CommentStatus
	}
	if page >= 0 && page < snapshot.CommentPages {
		comments, found, err := s.config.AssignmentContext.LoadAssignmentContextCommentPage(ctx, snapshot.AssignmentID, page)
		if err != nil {
			return nil, nil, false, err
		}
		if found {
			out.Comments = comments
		}
		if page+1 < snapshot.CommentPages {
			out.NextCursor = strconv.Itoa(page + 1)
		}
	}
	return out, assignmentContextCommentSources(out.Comments), true, nil
}
