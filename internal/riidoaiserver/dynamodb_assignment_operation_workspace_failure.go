package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

// The state row identifies the exact operation that committed the failure,
// including older executions whose client thread dropped the diagnostic.
func (s *DynamoDBAssignmentOperationStore) readWorkspacePreparationFailure(ctx context.Context, credentials AWSCredentials, projection assignmentProjectionRecord, stateItem map[string]map[string]string) (bool, error) {
	recordedAt, err := time.Parse(time.RFC3339Nano, dynamoDBStringValue(stateItem, "recorded_at"))
	if err != nil {
		return false, fmt.Errorf("decode failed assignment projection recorded_at: %w", err)
	}
	key := assignmentOperationKey(recordedAt, projection.LastEventSeq, dynamoDBStringValue(stateItem, "operation_id"))
	item, err := s.getAssignmentJournalItem(ctx, credentials, dynamoDBAssignmentOperationPK, key)
	if err != nil {
		return false, fmt.Errorf("dynamodb get terminal assignment operation: %w", err)
	}
	if len(item) == 0 {
		return false, nil
	}
	record, err := assignmentOperationRecordFromDynamoDBItem(item)
	if err != nil {
		return false, err
	}
	for _, event := range record.Events {
		if event.Seq == projection.LastEventSeq && isWorkspacePreparationFailure(projection.Assignment, event) {
			return true, nil
		}
	}
	return false, nil
}
