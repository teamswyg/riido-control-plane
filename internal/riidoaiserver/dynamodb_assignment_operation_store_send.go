package riidoaiserver

import (
	"context"
	"errors"
)

func dynamoDBAssignmentOperationContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (s *DynamoDBAssignmentOperationStore) send(ctx context.Context, cmd dynamoDBAssignmentOperationCommand) error {
	select {
	case s.commands <- cmd:
		return nil
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *DynamoDBAssignmentOperationStore) receiveError(ctx context.Context, reply <-chan error) error {
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errors.New("riidoaiserver: DynamoDB assignment operation store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}
