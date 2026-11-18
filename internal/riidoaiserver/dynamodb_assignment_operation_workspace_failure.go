package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

// The state row identifies the exact operation that committed the terminal
// result, including older executions whose client projection lost its body.
func (s *DynamoDBAssignmentOperationStore) readTerminalAssignmentEvent(ctx context.Context, credentials AWSCredentials, projection assignmentProjectionRecord, stateItem map[string]map[string]string) (*TaskEvent, error) {
	recordedAt, err := time.Parse(time.RFC3339Nano, dynamoDBStringValue(stateItem, "recorded_at"))
	if err != nil {
		return nil, fmt.Errorf("decode terminal assignment projection recorded_at: %w", err)
	}
	key := assignmentOperationKey(recordedAt, projection.LastEventSeq, dynamoDBStringValue(stateItem, "operation_id"))
	item, err := s.getAssignmentJournalItem(ctx, credentials, dynamoDBAssignmentOperationPK, key)
	if err != nil {
		return nil, fmt.Errorf("dynamodb get terminal assignment operation: %w", err)
	}
	if len(item) == 0 {
		return nil, nil
	}
	record, err := assignmentOperationRecordFromDynamoDBItem(item)
	if err != nil {
		return nil, err
	}
	for _, event := range record.Events {
		if event.Seq == projection.LastEventSeq && assignmentTerminalEventMatches(projection.Assignment, event) {
			return &event, nil
		}
	}
	return nil, nil
}
