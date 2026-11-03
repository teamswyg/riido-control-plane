package riidoaiserver

import "context"

// Older executions never persisted the new binding. A conversation string alone
// is not provenance: every parent edge must reach the same original assignment.
func verifiedDirectConversationRoot(selected AIAgentTaskThreadRecord, rows []AIAgentTaskThreadRecord) (AIAgentTaskThreadRecord, bool) {
	byID := make(map[string]AIAgentTaskThreadRecord, len(rows))
	for _, row := range rows {
		byID[row.ThreadID] = row
	}
	conversation := taskThreadConversationID(selected)
	for id := selected.ThreadID; id != ""; {
		row, ok := byID[id]
		if !ok || row.TaskID != selected.TaskID || taskThreadConversationID(row) != conversation {
			break
		}
		if row.ParentThreadID == "" {
			return row, row.ThreadID == conversation && row.SourceCommentID == ""
		}
		delete(byID, id)
		id = row.ParentThreadID
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s Server) recoverDirectInvocation(ctx context.Context, p AuthorizationResult, root AIAgentTaskThreadRecord) (*CommentInvocation, error) {
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok || s.config.AssignmentContext == nil {
		return nil, nil
	}
	snap, found, err := s.config.AssignmentContext.LoadAssignmentContextSnapshot(ctx, root.AssignmentID)
	if err != nil {
		return nil, err
	}
	if !found || snap.TriggerKind != AssignmentContextTriggerAssign || snap.AssignmentID != root.AssignmentID || snap.RequesterPrincipalID != p.PrincipalID || snap.WorkspaceID != p.WorkspaceID || snap.TaskID != root.TaskID || snap.AgentID != root.AgentID {
		return nil, nil
	}
	proj, found, err := reader.LoadAssignmentProjection(ctx, root.AssignmentID)
	if err != nil {
		return nil, err
	}
	a := proj.Assignment
	if !found || a.ID != root.AssignmentID || a.CreatedBy != p.PrincipalID || a.TeamID == "" || a.WorkspaceID != p.WorkspaceID || a.TaskID != root.TaskID || a.AgentID != root.AgentID {
		return nil, nil
	}
	return newDirectInvocation(actionResponseFromThread(root, p.WorkspaceID), p, a.TeamID), nil
}
