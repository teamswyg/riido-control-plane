package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

// This boundary distinguishes a rejected write from a committed write whose
// acknowledgement was lost. It copies bytes so memory aliasing cannot fake durability.
type onboardingAgentCreateFaultSnapshot struct {
	mu              sync.Mutex
	raw             []byte
	failBefore      bool
	commitThenError bool
}

func (s *onboardingAgentCreateFaultSnapshot) LoadAIAgentClientSnapshot(context.Context) (AIAgentClientSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.raw) == 0 {
		return AIAgentClientSnapshot{}, false, nil
	}
	var snapshot AIAgentClientSnapshot
	err := json.Unmarshal(s.raw, &snapshot)
	return snapshot, true, err
}

func (s *onboardingAgentCreateFaultSnapshot) SaveAIAgentClientSnapshot(_ context.Context, snapshot AIAgentClientSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failBefore && len(snapshot.OnboardingAgentCreates) > 0 {
		return errors.New("fixture write rejected before commit")
	}
	if len(s.raw) != 0 {
		var committed AIAgentClientSnapshot
		if err := json.Unmarshal(s.raw, &committed); err != nil {
			return err
		}
		if committed.Version != snapshot.ExpectedVersion {
			return errAIAgentSnapshotConflict
		}
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	s.raw = raw
	if s.commitThenError && len(snapshot.OnboardingAgentCreates) > 0 {
		s.commitThenError = false
		return errors.New("fixture commit acknowledgement lost")
	}
	return nil
}
