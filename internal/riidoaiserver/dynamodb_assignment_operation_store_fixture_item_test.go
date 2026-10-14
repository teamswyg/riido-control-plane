package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func sampleAssignmentOperationDynamoDBItem(t *testing.T, record AssignmentOperationRecord) map[string]map[string]string {
	t.Helper()
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		t.Fatalf("marshal assignment: %v", err)
	}
	eventsJSON, err := json.Marshal(record.Events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	return map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentOperationPK},
		"sk":                 {"S": assignmentOperationSortKey(record)},
		"schema_version":     {"S": record.SchemaVersion},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"task_id":            {"S": record.TaskID},
		"assignment_id":      {"S": record.AssignmentID},
		"operation_agent_id": {"S": record.AgentID},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"assignment_json":    {"S": string(assignmentJSON)},
		"events_json":        {"S": string(eventsJSON)},
	}
}

func sampleAssignmentProjectionDynamoDBItem(t *testing.T, assignment Assignment, lastEventSeq int64) map[string]map[string]string {
	t.Helper()
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		t.Fatalf("marshal assignment: %v", err)
	}
	return map[string]map[string]string{
		"pk":              {"S": dynamoDBAssignmentProjectionPK(assignment.ID)},
		"sk":              {"S": dynamoDBAssignmentProjectionSK},
		"schema_version":  {"S": AssignmentProjectionSchemaVersion},
		"assignment_id":   {"S": assignment.ID},
		"agent_id":        {"S": assignment.AgentID},
		"assignment_sort": {"S": assignmentQueueSort(assignment)},
		"last_event_seq":  {"N": strconv.FormatInt(lastEventSeq, 10)},
		"assignment_json": {"S": string(assignmentJSON)},
	}
}
