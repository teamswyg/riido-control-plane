package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) updateAgentActiveAssignmentPayload(assignment Assignment, heartbeatAt time.Time) ([]byte, error) {
	heartbeatAt = heartbeatAt.UTC()
	leaseExpiresAt := s.activeLeaseExpiresAt(heartbeatAt)
	payload := struct {
		TableName                 string                       `json:"TableName"`
		Key                       map[string]map[string]string `json:"Key"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		UpdateExpression          string                       `json:"UpdateExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	}{
		TableName:           s.tableName,
		Key:                 dynamoDBAgentActiveKey(assignment.AgentID),
		ConditionExpression: "active_assignment_id = :assignment_id AND lease_token = :lease_token",
		UpdateExpression:    "SET assignment_state = :assignment_state, lease_heartbeat_at = :heartbeat_at, lease_expires_at = :lease_expires_at, lease_expires_unix_ms = :lease_expires_unix_ms, updated_at = :heartbeat_at",
		ExpressionAttributeValues: map[string]map[string]string{
			":assignment_id":         {"S": assignment.ID},
			":lease_token":           {"S": assignment.LeaseToken},
			":assignment_state":      {"S": string(assignment.State)},
			":heartbeat_at":          {"S": heartbeatAt.Format(time.RFC3339Nano)},
			":lease_expires_at":      {"S": leaseExpiresAt.Format(time.RFC3339Nano)},
			":lease_expires_unix_ms": {"N": strconv.FormatInt(leaseExpiresAt.UnixMilli(), 10)},
		},
	}
	return json.Marshal(payload)
}

func dynamoDBAgentActiveKey(agentID string) map[string]map[string]string {
	return map[string]map[string]string{
		"pk": {"S": dynamoDBAgentActivePK(agentID)},
		"sk": {"S": dynamoDBAgentActiveSK},
	}
}
