package riidoaiserver

import (
	"context"
)

type AIAgentTaskThreadProjectionReconciler interface {
	ReconcileAIAgentActiveThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string, reader AssignmentProjectionReader) (bool, error)
}
