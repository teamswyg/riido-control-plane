package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	errAssignmentContextParentMissing = errors.New("parent assignment context snapshot unavailable")
	errAssignmentContextParentScope   = errors.New("parent assignment context belongs to another task, workspace or requester")
)

// InheritAssignmentContext gives a run started by another agent's mention
// (RIID-5892) the snapshot of the run that mentioned it. The child has no
// requester Bearer of its own, so it may only reuse what the parent run was
// already allowed to read, for the same task, workspace and requester. No
// prompt reads are recorded: the child's prompt is built by its caller.
func (s Server) InheritAssignmentContext(ctx context.Context, parentAssignmentID string, child Assignment, triggerID string) error {
	store := s.config.AssignmentContext
	if store == nil {
		return nil
	}
	if strings.TrimSpace(child.ID) == "" {
		return errors.New("child assignment_id is required")
	}
	if _, exists, err := store.LoadAssignmentContextSnapshot(ctx, child.ID); err != nil || exists {
		return err
	}
	parent, found, err := store.LoadAssignmentContextSnapshot(ctx, strings.TrimSpace(parentAssignmentID))
	if err != nil {
		return err
	}
	if !found {
		return errAssignmentContextParentMissing
	}
	requester := strings.TrimSpace(child.CreatedBy)
	if parent.TaskID != child.TaskID || parent.WorkspaceID != child.WorkspaceID || (requester != "" && requester != parent.RequesterPrincipalID) {
		return errAssignmentContextParentScope
	}
	pages := make([][]AIAgentTaskContextComment, 0, parent.CommentPages)
	for page := range parent.CommentPages {
		comments, found, err := store.LoadAssignmentContextCommentPage(ctx, parent.AssignmentID, page)
		if err != nil {
			return err
		}
		if !found {
			return errAssignmentContextParentMissing
		}
		pages = append(pages, comments)
	}
	if len(pages) > 0 {
		if err := store.SaveAssignmentContextComments(ctx, child.ID, pages); err != nil {
			return err
		}
	}
	snapshot := parent
	snapshot.AssignmentID, snapshot.AgentID = child.ID, child.AgentID
	snapshot.TriggerKind, snapshot.TriggerID = AssignmentContextTriggerAgentMention, strings.TrimSpace(triggerID)
	snapshot.CapturedAt, snapshot.CommentPages = time.Now().UTC(), len(pages)
	return store.SaveAssignmentContextSnapshot(ctx, snapshot)
}
