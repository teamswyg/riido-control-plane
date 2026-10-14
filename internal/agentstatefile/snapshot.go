package agentstatefile

import (
	"sync"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

// FileAIAgentClientSnapshot is a single-process local-development store.
type AIAgentClientSnapshot = riidoaiserver.AIAgentClientSnapshot

type FileAIAgentClientSnapshot struct {
	path string
	mu   sync.Mutex
}

func NewFileAIAgentClientSnapshot(path string) *FileAIAgentClientSnapshot {
	return &FileAIAgentClientSnapshot{path: path}
}
