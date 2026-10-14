package riidoaiserver

import (
	"context"
	"strings"
)

func queueDiagnosticsFromAssignmentProjection(ctx context.Context, reader AssignmentProjectionReader, projection AssignmentProjection) (*AIAgentTaskThreadQueueDiagnostics, error) {
	assignment := projection.Assignment
	blockedByID := strings.TrimSpace(assignment.BlockedByAssignmentID)
	if assignment.State.Code() != AssignmentStateCodeQueued || blockedByID == "" {
		return nil, nil
	}
	diagnostics := &AIAgentTaskThreadQueueDiagnostics{
		Reason:                "blocked_by_assignment",
		BlockedByAssignmentID: blockedByID,
	}
	blocker, ok, err := reader.LoadAssignmentProjection(ctx, blockedByID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return diagnostics, nil
	}
	diagnostics.BlockerAgentID = blocker.Assignment.AgentID
	diagnostics.BlockerRuntimeProvider = blocker.Assignment.RuntimeProvider
	diagnostics.BlockerState = blocker.Assignment.State
	diagnostics.BlockerUpdatedAt = blocker.Assignment.UpdatedAt
	return diagnostics, nil
}
