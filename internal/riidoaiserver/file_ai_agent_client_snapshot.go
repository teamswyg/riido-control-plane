package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// FileAIAgentClientSnapshot is a single-process local-development store.
type FileAIAgentClientSnapshot struct {
	path string
	mu   sync.Mutex
}

func NewFileAIAgentClientSnapshot(path string) *FileAIAgentClientSnapshot {
	return &FileAIAgentClientSnapshot{path: path}
}
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
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
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
