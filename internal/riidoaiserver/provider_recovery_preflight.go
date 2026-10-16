package riidoaiserver

import (
	"context"
	"errors"
	"strings"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) preflightRecoveryCandidate(ctx context.Context, agentID string, req AgentEventRequest) (*recoveryCandidateApproval, error) {
	recovery := assignment.RecoveryContextFromMetadata(req.Metadata)
	if recovery == nil || (recovery.State != "checking" && recovery.State != "switching") {
		return nil, nil
	}
	key := strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()])
	if strings.TrimSpace(req.EventType) != EventRiidoLog || key == "" || len(key) > 256 {
		return nil, errors.New("recovery event requires a bounded stable event key")
	}
	ctx, cancel := context.WithTimeout(ctx, recoveryCandidateTimeout)
	defer cancel()
	projection, found, err := s.LoadAssignmentProjection(ctx, req.AssignmentID)
	if err != nil {
		return nil, assignmentDependencyFailure(err)
	}
	if !found {
		return nil, errors.New("assignment not found")
	}
	current := projection.Assignment
	if err := s.validateRecoveryPreflightIdentity(agentID, req, current); err != nil {
		return nil, err
	}
	if projection.Recovery != nil {
		hash := recoveryHash(strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()]))
		for _, receipt := range projection.Recovery.Receipts {
			if receipt.KeyHash == hash && receipt.ContextHash == recoveryContextHash(*recovery) {
				return newRecoveryCandidateApproval(current, recovery.Provider)
			}
		}
	}
	if recovery.Provider == current.RuntimeProvider || recovery.Provider == "unknown" {
		return nil, nil
	}
	if err := s.recheckRecoveryCandidate(ctx, current, recovery.Provider); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return newRecoveryCandidateApproval(current, recovery.Provider)
}
