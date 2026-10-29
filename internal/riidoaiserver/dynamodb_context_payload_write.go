package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// Captures fence both the lifecycle revision and assignment projection. A late
// capture retries against the committed terminal time instead of erasing TTL.
func (s *DynamoDBAssignmentOperationStore) saveContextPayloads(ctx context.Context, id string, items []map[string]map[string]string) error {
	return s.saveContextPayloadsWithPolicy(ctx, id, items, false)
}

func (s *DynamoDBAssignmentOperationStore) saveContextPayloadsWithPolicy(ctx context.Context, id string, items []map[string]map[string]string, immutable bool) error {
	if err := validateContextPayloadBudget(items); err != nil {
		return err
	}
	for range 3 {
		marker, err := s.contextRetentionItem(ctx, id, assignmentContextLifecycleSK)
		if err != nil {
			return err
		}
		credentials, err := s.credentialsProvider.Credentials(ctx)
		if err != nil {
			return err
		}
		projection, found, err := s.loadAssignmentProjection(ctx, id, credentials)
		if err != nil {
			return err
		}
		var terminal time.Time
		if found && isTerminal(projection.Assignment.State) {
			if projection.Assignment.UpdatedAt.IsZero() {
				return errors.New("terminal assignment timestamp missing")
			}
			terminal = projection.Assignment.UpdatedAt
		}
		expiry, err := assignmentContextExpiry(marker, terminal)
		if err != nil {
			return err
		}
		if expiry > 0 && !s.now().Before(time.Unix(expiry, 0)) {
			return errAssignmentContextExpired
		}
		fence, err := s.contextLifecycleFence(id, marker, expiry)
		if err != nil {
			return err
		}
		actions := []dynamoDBTransactWriteItem{fence, s.contextProjectionFence(id, projection, found)}
		actions = append(actions, s.contextPublicationFences(id, items, immutable)...)
		for _, item := range items {
			if expiry > 0 {
				item[assignmentContextTTLAttribute] = map[string]string{"N": strconv.FormatInt(expiry, 10)}
			}
		}
		actions = append(actions, s.contextPayloadPuts(items, immutable)...)
		_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if err == nil || !isDynamoDBTransactionContention(err) {
			return err
		}
	}
	return errAssignmentProjectionConflict
}
