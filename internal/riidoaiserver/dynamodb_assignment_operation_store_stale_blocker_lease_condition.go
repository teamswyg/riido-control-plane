package riidoaiserver

import (
	"context"
	"strconv"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) staleBlockerLeaseCondition(ctx context.Context, blocker Assignment, at time.Time, credentials AWSCredentials) (bool, *dynamoDBTransactWriteConditionCheckAction, error) {
	if !assignmentHoldsActiveLease(blocker.State) {
		return false, nil, nil
	}
	key := map[string]map[string]string{
		"pk": {"S": dynamoDBAgentActivePK(blocker.AgentID)},
		"sk": {"S": dynamoDBAgentActiveSK},
	}
	lease, exists, err := s.loadAgentActiveAssignment(ctx, blocker.AgentID, credentials)
	if err != nil {
		return false, nil, err
	}
	if !exists {
		return true, &dynamoDBTransactWriteConditionCheckAction{
			TableName:           s.tableName,
			Key:                 key,
			ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		}, nil
	}
	if lease.ActiveAssignmentID != blocker.ID {
		return true, &dynamoDBTransactWriteConditionCheckAction{
			TableName:           s.tableName,
			Key:                 key,
			ConditionExpression: "active_assignment_id <> :blocked_assignment_id",
			ExpressionAttributeValues: map[string]map[string]string{
				":blocked_assignment_id": {"S": blocker.ID},
			},
		}, nil
	}
	if !lease.Expired(at) {
		return false, nil, nil
	}
	return true, &dynamoDBTransactWriteConditionCheckAction{
		TableName: s.tableName,
		Key:       key,
		ConditionExpression: "active_assignment_id = :blocked_assignment_id AND " +
			"((attribute_exists(lease_expires_unix_ms) AND lease_expires_unix_ms <= :claim_started_unix_ms) OR " +
			"(attribute_not_exists(lease_expires_unix_ms) AND lease_expires_at <= :claim_started_at))",
		ExpressionAttributeValues: map[string]map[string]string{
			":blocked_assignment_id": {"S": blocker.ID},
			":claim_started_unix_ms": {"N": strconv.FormatInt(at.UTC().UnixMilli(), 10)},
			":claim_started_at":      {"S": at.UTC().Format(time.RFC3339Nano)},
		},
	}, nil
}

func (s *DynamoDBAssignmentOperationStore) loadAgentQueue(ctx context.Context, agentID string, credentials AWSCredentials) ([]Assignment, error) {
	records, err := s.loadAgentQueueRecords(ctx, agentID, credentials)
	if err != nil {
		return nil, err
	}
	assignments := make([]Assignment, 0, len(records))
	for _, record := range records {
		assignments = append(assignments, record.Assignment)
	}
	return assignments, nil
}
