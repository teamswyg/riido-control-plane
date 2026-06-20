package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) LoadAgentQueueAssignments(ctx context.Context, agentID string) ([]Assignment, error) {
	if s == nil {
		return nil, nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return nil, errors.New("riidoaiserver: assignment queue agent_id is required")
	}
	reply := make(chan dynamoDBAssignmentQueueResult, 1)
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, queue: true, agentID: agentID, queueReply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return nil, err
	}
	select {
	case result := <-reply:
		return result.assignments, result.err
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
