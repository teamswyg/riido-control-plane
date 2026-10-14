package riidoaiserver

import (
	"context"
)

type AIAgentAssignmentEventRecorder interface {
	RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error
}
