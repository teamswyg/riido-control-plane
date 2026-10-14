package agentstatefile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
)

func (s *FileAIAgentClientSnapshot) load(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentClientSnapshot{}, false, err
	}
	body, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return AIAgentClientSnapshot{}, false, nil
	}
	if err != nil {
		return AIAgentClientSnapshot{}, false, err
	}
	var snapshot AIAgentClientSnapshot
	err = json.Unmarshal(body, &snapshot)
	return snapshot, err == nil, err
}

func (s *FileAIAgentClientSnapshot) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(ctx)
}
