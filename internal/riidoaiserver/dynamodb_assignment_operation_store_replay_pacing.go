package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

const dynamoDBAssignmentOperationReplayPageDelay = 100 * time.Millisecond

func loadDynamoDBAssignmentOperationPage(ctx context.Context, request dynamoDBRequest) ([]byte, error) {
	var body []byte
	err := retryStoreOpenTransient(ctx, func() error {
		page, err := doDynamoDBJSON(ctx, request)
		if err != nil {
			return fmt.Errorf("dynamodb query assignment operations: %w", err)
		}
		body = page
		return nil
	})
	return body, err
}

func paceDynamoDBAssignmentOperationReplay(ctx context.Context) error {
	return sleepStoreOpenRetry(ctx, dynamoDBAssignmentOperationReplayPageDelay)
}
