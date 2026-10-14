package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) agentActiveAssignmentDynamoDBItem(record AssignmentOperationRecord) map[string]map[string]string {
	leaseHeartbeatAt := record.RecordedAt.UTC()
	leaseExpiresAt := s.activeLeaseExpiresAt(leaseHeartbeatAt)
	return map[string]map[string]string{
		"pk":                    {"S": dynamoDBAgentActivePK(record.AgentID)},
		"sk":                    {"S": dynamoDBAgentActiveSK},
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": record.AgentID},
		"active_assignment_id":  {"S": record.AssignmentID},
		"task_id":               {"S": record.TaskID},
		"component_id":          {"S": record.Assignment.ComponentID},
		"lease_token":           {"S": record.Assignment.LeaseToken},
		"operation_id":          {"S": record.OperationID},
		"operation_type":        {"S": string(record.OperationType)},
		"runtime_provider":      {"S": record.Assignment.RuntimeProvider},
		"assignment_state":      {"S": string(record.Assignment.State)},
		"leased_at":             {"S": leaseHeartbeatAt.Format(time.RFC3339Nano)},
		"lease_heartbeat_at":    {"S": leaseHeartbeatAt.Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": leaseExpiresAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(leaseExpiresAt.UnixMilli(), 10)},
		"updated_at":            {"S": record.Assignment.UpdatedAt.UTC().Format(time.RFC3339Nano)},
		"last_event_seq":        {"N": strconv.FormatInt(assignmentOperationLastEventSeq(record), 10)},
	}
}

func (s *DynamoDBAssignmentOperationStore) activeLeaseExpiresAt(at time.Time) time.Time {
	return at.UTC().Add(s.activeLeaseDuration)
}

func (s *DynamoDBAssignmentOperationStore) deleteAgentActiveAssignmentPayload(record AssignmentOperationRecord) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		Key                       map[string]map[string]string `json:"Key"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	}{
		TableName:           s.tableName,
		Key:                 dynamoDBAgentActiveKey(record.AgentID),
		ConditionExpression: "active_assignment_id = :assignment_id",
		ExpressionAttributeValues: map[string]map[string]string{
			":assignment_id": {"S": record.AssignmentID},
		},
	}
	return json.Marshal(payload)
}
