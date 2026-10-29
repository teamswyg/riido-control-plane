package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

func (s Server) inheritAssignmentContent(ctx context.Context, parent AssignmentContextSnapshot, child Assignment, triggerID string) error {
	store, ok := s.config.AssignmentContext.(AssignmentContentStore)
	if !ok {
		return errAssignmentContextParentMissing
	}
	var pages []AssignmentContentPage
	for _, part := range parent.Content {
		for range part.Cursors {
			page, found, err := store.LoadAssignmentContentPage(ctx, parent.AssignmentID, len(pages))
			if err != nil {
				return err
			}
			if !found {
				return errAssignmentContextParentMissing
			}
			pages = append(pages, page)
		}
	}
	snapshot := parent
	snapshot.AssignmentID, snapshot.AgentID = child.ID, child.AgentID
	snapshot.TriggerKind, snapshot.TriggerID = AssignmentContextTriggerAgentMention, strings.TrimSpace(triggerID)
	snapshot.CapturedAt = time.Now().UTC()
	return store.SaveAssignmentContentBundle(ctx, snapshot, pages)
}
