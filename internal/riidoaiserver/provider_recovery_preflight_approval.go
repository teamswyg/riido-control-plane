package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const recoveryCandidateTimeout = 2 * time.Second

type recoveryCandidateApproval struct {
	fingerprint string
	provider    string
	expiresAt   time.Time
}

func candidateAssignmentFingerprint(current Assignment) (string, error) {
	current.UpdatedAt = time.Time{}
	data, err := json.Marshal(current)
	if err != nil {
		return "", err
	}
	return recoveryHash(string(data)), nil
}

func (s *Store) checkRecoveryCandidateApproval(current Assignment, provider string) error {
	fingerprint, err := candidateAssignmentFingerprint(current)
	if err != nil {
		return err
	}
	approval := s.recoveryApproval
	if approval == nil || approval.provider != provider || time.Now().After(approval.expiresAt) || approval.fingerprint != fingerprint {
		return errors.New("provider recovery candidate approval changed or expired; retry the request")
	}
	return nil
}

func newRecoveryCandidateApproval(current Assignment, provider string) (*recoveryCandidateApproval, error) {
	fingerprint, err := candidateAssignmentFingerprint(current)
	if err != nil {
		return nil, err
	}
	return &recoveryCandidateApproval{fingerprint: fingerprint, provider: provider, expiresAt: time.Now().Add(recoveryCandidateTimeout)}, nil
}

func (s *Store) validateRecoveryPreflightIdentity(agentID string, req AgentEventRequest, current Assignment) error {
	if current.AgentID != strings.TrimSpace(agentID) || (req.TaskID != "" && strings.TrimSpace(req.TaskID) != current.TaskID) {
		return ErrAuthorizationForbidden
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return err
	}
	if req.DaemonID == "" || req.DeviceID == "" || req.RuntimeID == "" || current.LeaseToken == "" || req.Metadata["lease_token"] != current.LeaseToken || current.State == AssignmentQueued || isTerminal(current.State) || current.State == AssignmentCancelling {
		return ErrAuthorizationForbidden
	}
	if (req.State != "" && req.State != current.State) || (req.ProviderSessionID != "" && req.ProviderSessionID != current.ProviderSessionID) {
		return ErrAuthorizationForbidden
	}
	return nil
}
