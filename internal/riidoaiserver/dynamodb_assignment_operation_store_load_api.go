package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentOperations(ctx context.Context) ([]AssignmentOperationRecord, error) {
	if s == nil {
		return nil, nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	reply := make(chan dynamoDBAssignmentOperationLoadResult, 1)
	if err := s.send(ctx, dynamoDBAssignmentOperationCommand{ctx: ctx, load: true, loadReply: reply}); err != nil {
		return nil, err
	}
	select {
	case result := <-reply:
		return result.records, result.err
	case <-s.done:
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
