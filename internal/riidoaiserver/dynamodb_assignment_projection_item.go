package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) putAssignmentProjectionPayload(record AssignmentOperationRecord) ([]byte, error) {
	item, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	lastEventSeq := assignmentOperationLastEventSeq(record)
	payload := struct {
		TableName                 string                       `json:"TableName"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		Item                      map[string]map[string]string `json:"Item"`
	}{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(last_event_seq) OR last_event_seq <= :last_event_seq",
		ExpressionAttributeValues: map[string]map[string]string{
			":last_event_seq": {"N": strconv.FormatInt(lastEventSeq, 10)},
		},
		Item: item,
	}
	return json.Marshal(payload)
}

func assignmentProjectionDynamoDBItem(record AssignmentOperationRecord) (map[string]map[string]string, error) {
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		return nil, err
	}
	item := assignmentProjectionBaseItem(record, string(assignmentJSON))
	assignmentProjectionOptionalFields(item, record.Assignment)
	return item, nil
}

func assignmentProjectionBaseItem(record AssignmentOperationRecord, assignmentJSON string) map[string]map[string]string {
	return map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentProjectionPK(record.AssignmentID)},
		"sk":                 {"S": dynamoDBAssignmentProjectionSK},
		"schema_version":     {"S": AssignmentProjectionSchemaVersion},
		"assignment_id":      {"S": record.AssignmentID},
		"task_id":            {"S": record.TaskID},
		"component_id":       {"S": record.Assignment.ComponentID},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"operation_agent_id": {"S": record.AgentID},
		"assignment_state":   {"S": string(record.Assignment.State)},
		"runtime_provider":   {"S": record.Assignment.RuntimeProvider},
		"prompt":             {"S": record.Assignment.Prompt},
		"created_at":         {"S": record.Assignment.CreatedAt.UTC().Format(time.RFC3339Nano)},
		"updated_at":         {"S": record.Assignment.UpdatedAt.UTC().Format(time.RFC3339Nano)},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":     {"N": strconv.FormatInt(assignmentOperationLastEventSeq(record), 10)},
		"assignment_json":    {"S": assignmentJSON},
	}
}
