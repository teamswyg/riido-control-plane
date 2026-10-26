package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) terminalContextActions(ctx context.Context, assignment Assignment) ([]dynamoDBTransactWriteItem, error) {
	marker, err := s.contextRetentionItem(ctx, assignment.ID, assignmentContextLifecycleSK)
	if err != nil {
		return nil, err
	}
	items, err := s.contextRetentionPayloads(ctx, assignment.ID)
	if err != nil {
		return nil, err
	}
	if assignment.UpdatedAt.IsZero() {
		return nil, errors.New("terminal assignment timestamp missing")
	}
	expiry, err := assignmentContextExpiry(marker, assignment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	// Keep one small marker even before capture: replayed terminal projections
	// must not move the first expiry for a later capture. No payload rows are created.
	fence, err := s.contextLifecycleFence(assignment.ID, marker, expiry)
	if err != nil {
		return nil, err
	}
	actions := []dynamoDBTransactWriteItem{fence}
	for _, item := range items {
		actions = append(actions, dynamoDBTransactWriteItem{Update: &dynamoDBTransactWriteUpdateAction{
			TableName: s.tableName, Key: map[string]map[string]string{"pk": item["pk"], "sk": item["sk"]},
			UpdateExpression: "SET " + assignmentContextTTLAttribute + " = :expiry", ConditionExpression: "attribute_exists(pk)",
			ExpressionAttributeValues: map[string]map[string]string{":expiry": {"N": strconv.FormatInt(expiry, 10)}},
		}})
	}
	return actions, nil
}
