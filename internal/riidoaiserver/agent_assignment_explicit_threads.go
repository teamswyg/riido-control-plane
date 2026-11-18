package riidoaiserver

import (
	"context"
	"errors"
)

func (s Server) explicitThreadHistory(ctx context.Context, a Assignment, v explicitAssignmentResult, req AgentAssignmentContextRequest) (any, error) {
	reader, ok := s.aiAgent.(interface {
		ListAIAgentTaskThreadHistory(context.Context, AuthorizationResult, string) (AIAgentTaskThreadHistoryCollectionResponse, error)
	})
	if !ok {
		return nil, errors.New("thread reader unavailable")
	}
	principal := AuthorizationResult{PrincipalID: a.CreatedBy, WorkspaceID: a.WorkspaceID}
	if v.Admin {
		principal.Roles = []AgentCatalogRole{AgentCatalogRoleAdmin}
	}
	history, err := reader.ListAIAgentTaskThreadHistory(ctx, principal, v.ID)
	if err != nil {
		return nil, err
	}
	data := map[string]any{"workspace_id": a.WorkspaceID, "task_id": v.ID, "history_scope": "server_retained", "projection_complete": true, "complete": false, "live": true}
	if req.View == "thread" {
		for _, t := range history.Threads {
			if t.ThreadID == req.ThreadID && t.TaskID == v.ID {
				data["thread"] = t
				data["agent_snapshot"] = history.AgentSnapshots[t.AgentSnapshotID]
				return data, nil
			}
		}
		return nil, ErrAuthorizationForbidden
	}
	summaries := []map[string]any{}
	for _, t := range history.Threads {
		summaries = append(summaries, map[string]any{"thread_id": t.ThreadID, "conversation_id": t.ConversationID, "task_id": t.TaskID, "agent_id": t.AgentID, "run_id": t.RunID, "work_status": t.WorkStatus, "message_count": len(t.Messages)})
	}
	data["threads"] = summaries
	return data, nil
}
