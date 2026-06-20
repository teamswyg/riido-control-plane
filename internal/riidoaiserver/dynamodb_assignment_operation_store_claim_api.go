package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) ClaimNextAssignment(ctx context.Context, agentID string, at time.Time) (AssignmentClaimResult, error) {
	if s == nil {
		return AssignmentClaimResult{}, nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AssignmentClaimResult{}, errors.New("riidoaiserver: assignment claim agent_id is required")
	}
	if at.IsZero() {
		at = s.now()
	}
	reply := make(chan dynamoDBAssignmentClaimResult, 1)
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, claim: true, agentID: agentID, claimAt: at, claimReply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return AssignmentClaimResult{}, err
	}
	select {
	case result := <-reply:
		return result.result, result.err
	case <-s.done:
		return AssignmentClaimResult{}, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentClaimResult{}, ctx.Err()
	}
}
