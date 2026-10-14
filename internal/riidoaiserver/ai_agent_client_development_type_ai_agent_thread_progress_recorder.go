package riidoaiserver

import (
	"context"
)

type AIAgentThreadProgressRecorder interface {
	RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error)
}
