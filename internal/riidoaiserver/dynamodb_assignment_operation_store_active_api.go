package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) LoadAgentActiveAssignment(ctx context.Context, agentID string) (AssignmentActiveLease, bool, error) {
	if s == nil {
		return AssignmentActiveLease{}, false, nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AssignmentActiveLease{}, false, errors.New("riidoaiserver: active lease agent_id is required")
	}
	reply := make(chan dynamoDBAssignmentActiveLeaseResult, 1)
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, active: true, agentID: agentID, activeReply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return AssignmentActiveLease{}, false, err
	}
	select {
	case result := <-reply:
		return result.lease, result.found, result.err
	case <-s.done:
		return AssignmentActiveLease{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentActiveLease{}, false, ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) RefreshAgentActiveAssignment(ctx context.Context, assignment Assignment, at time.Time) error {
	if s == nil {
		return nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	if at.IsZero() {
		at = s.now()
	}
	assignmentCopy := assignment
	reply := make(chan error, 1)
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, refresh: true, assignment: &assignmentCopy, refreshAt: at, reply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return err
	}
	return s.receiveError(ctx, reply)
}
