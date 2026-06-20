package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) SaveAssignmentOperation(ctx context.Context, record AssignmentOperationRecord) error {
	if s == nil {
		return nil
	}
	ctx = dynamoDBAssignmentOperationContext(ctx)
	reply := make(chan error, 1)
	recordCopy := record
	cmd := dynamoDBAssignmentOperationCommand{ctx: ctx, record: &recordCopy, reply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}
