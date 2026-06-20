package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ReconcileAIAgentActiveThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string, reader AssignmentProjectionReader) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if reader == nil {
		return false, nil
	}
	taskID = strings.TrimSpace(taskID)
	candidates := s.reconcilableActiveTaskThreadProjectionCandidates(principal, taskID)
	changed := false
	for _, thread := range candidates {
		if err := ctx.Err(); err != nil {
			return changed, err
		}
		projection, ok, err := reader.LoadAssignmentProjection(ctx, thread.AssignmentID)
		if err != nil {
			return changed, err
		}
		if !ok ||
			!assignmentProjectionMatchesTaskThread(thread, projection) ||
			!assignmentStateCanRepairTaskThread(projection.Assignment.State) {
			continue
		}
		diagnostics, err := queueDiagnosticsFromAssignmentProjection(ctx, reader, projection)
		if err != nil {
			return changed, err
		}
		if s.applyAssignmentProjectionToTaskThread(thread, projection, diagnostics) {
			changed = true
		}
	}
	return changed, nil
}
