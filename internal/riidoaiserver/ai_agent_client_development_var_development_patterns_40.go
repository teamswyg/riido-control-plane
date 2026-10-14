package riidoaiserver

import (
	"errors"
)

var (
	ErrAIAgentNotFound           = errors.New("ai agent not found")
	ErrAIAgentAssigned           = errors.New("ai agent has assigned tasks")
	ErrAIAgentTaskThreadConflict = errors.New("task already has another active ai agent thread")
)
