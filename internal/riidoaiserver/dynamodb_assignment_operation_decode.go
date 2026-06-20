package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"time"
)

func assignmentOperationRecordFromDynamoDBItem(item map[string]map[string]string) (AssignmentOperationRecord, error) {
	var assignment Assignment
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "assignment_json")), &assignment); err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation assignment_json: %w", err)
	}
	var events []TaskEvent
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "events_json")), &events); err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation events_json: %w", err)
	}
	recordedAt, err := time.Parse(time.RFC3339Nano, dynamoDBStringValue(item, "recorded_at"))
	if err != nil {
		return AssignmentOperationRecord{}, fmt.Errorf("decode assignment operation recorded_at: %w", err)
	}
	record := AssignmentOperationRecord{
		SchemaVersion: dynamoDBStringValue(item, "schema_version"),
		OperationID:   dynamoDBStringValue(item, "operation_id"),
		OperationType: AssignmentOperationType(dynamoDBStringValue(item, "operation_type")),
		TaskID:        dynamoDBStringValue(item, "task_id"),
		AssignmentID:  dynamoDBStringValue(item, "assignment_id"),
		AgentID:       dynamoDBStringValue(item, "operation_agent_id"),
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    recordedAt,
	}
	if record.AgentID == "" {
		record.AgentID = dynamoDBStringValue(item, "agent_id")
	}
	if err := validateAssignmentOperationRecord(record); err != nil {
		return AssignmentOperationRecord{}, err
	}
	return record, nil
}
