package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentProjection(ctx context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	if s == nil {
		return AssignmentProjection{}, false, nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("riidoaiserver: assignment projection assignment_id is required")
	}
	reply := make(chan dynamoDBAssignmentProjectionResult, 1)
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, projection: true, assignmentID: assignmentID, projectionReply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return AssignmentProjection{}, false, err
	}
	select {
	case result := <-reply:
		return result.projection, result.found, result.err
	case <-s.done:
		return AssignmentProjection{}, false, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return AssignmentProjection{}, false, ctx.Err()
	}
}
