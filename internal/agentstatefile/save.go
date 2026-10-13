package agentstatefile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func (s *FileAIAgentClientSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists, err := s.load(ctx)
	if err != nil {
		return err
	}
	if exists && current.Version != snapshot.ExpectedVersion {
		return errors.New("local Agent snapshot version conflict")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".agent-state-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(snapshot)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), s.path)
}
