package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) putOperationPayload(record AssignmentOperationRecord) ([]byte, error) {
	item, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	payload := struct {
		TableName           string                       `json:"TableName"`
		ConditionExpression string                       `json:"ConditionExpression"`
		Item                map[string]map[string]string `json:"Item"`
	}{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		Item:                item,
	}
	return json.Marshal(payload)
}

func assignmentOperationDynamoDBItem(record AssignmentOperationRecord) (map[string]map[string]string, error) {
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		return nil, err
	}
	eventsJSON, err := json.Marshal(record.Events)
	if err != nil {
		return nil, err
	}
	item := map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentOperationPK},
		"sk":                 {"S": assignmentOperationSortKey(record)},
		"schema_version":     {"S": record.SchemaVersion},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"task_id":            {"S": record.TaskID},
		"assignment_id":      {"S": record.AssignmentID},
		"operation_agent_id": {"S": record.AgentID},
		"assignment_state":   {"S": string(record.Assignment.State)},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":     {"N": strconv.FormatInt(assignmentOperationLastEventSeq(record), 10)},
		"event_count":        {"N": strconv.Itoa(len(record.Events))},
		"assignment_json":    {"S": string(assignmentJSON)},
		"events_json":        {"S": string(eventsJSON)},
	}
	return item, nil
}
