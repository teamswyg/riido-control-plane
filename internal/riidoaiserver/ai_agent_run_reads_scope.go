package riidoaiserver

import (
	"context"
	"errors"
)

// Stored identity chooses the API shape only. Visibility still comes entirely
// from a fresh authorized request carrying the current viewer's own token.
func (s Server) currentRunReadRequiresScope(ctx context.Context, assignmentID, workspaceID, taskID string) (bool, error) {
	if s.config.AssignmentContext == nil || assignmentID == "" {
		return false, nil
	}
	snapshot, found, err := s.config.AssignmentContext.LoadAssignmentContextSnapshot(ctx, assignmentID)
	if err != nil {
		return false, err
	}
	if !found || snapshot.Context.ReadScope == nil {
		return false, nil
	}
	if snapshot.AssignmentID != assignmentID || snapshot.WorkspaceID != workspaceID || snapshot.TaskID != taskID || snapshot.Context.ReadScope.RootID != taskID || snapshot.Context.ReadScope.SchemaVersion != taskReadScopeSchema {
		return false, errors.New("captured read scope identity unavailable")
	}
	return true, nil
}
