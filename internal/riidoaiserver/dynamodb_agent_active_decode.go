package riidoaiserver

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func agentActiveAssignmentFromDynamoDBItem(item map[string]map[string]string) (AssignmentActiveLease, error) {
	if schema := dynamoDBStringValue(item, "schema_version"); schema != AssignmentAgentActiveSchemaVersion {
		return AssignmentActiveLease{}, fmt.Errorf("unsupported agent active assignment schema_version %q", schema)
	}
	heartbeatAt, err := parseDynamoDBOptionalTime(item, "lease_heartbeat_at")
	if err != nil {
		return AssignmentActiveLease{}, err
	}
	expiresAt, err := parseDynamoDBOptionalTime(item, "lease_expires_at")
	if err != nil {
		return AssignmentActiveLease{}, err
	}
	expiresUnixMS, err := parseDynamoDBOptionalInt64(item, "lease_expires_unix_ms")
	if err != nil {
		return AssignmentActiveLease{}, err
	}
	lease := AssignmentActiveLease{
		AgentID:            dynamoDBStringValue(item, "agent_id"),
		ActiveAssignmentID: dynamoDBStringValue(item, "active_assignment_id"),
		LeaseToken:         dynamoDBStringValue(item, "lease_token"),
		HeartbeatAt:        heartbeatAt,
		LeaseExpiresAt:     expiresAt,
		LeaseExpiresUnixMS: expiresUnixMS,
	}
	if lease.AgentID == "" || lease.ActiveAssignmentID == "" {
		return AssignmentActiveLease{}, errors.New("agent active assignment agent_id and active_assignment_id are required")
	}
	return lease, nil
}

func parseDynamoDBOptionalTime(item map[string]map[string]string, key string) (time.Time, error) {
	raw := dynamoDBStringValue(item, key)
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("decode agent active assignment %s: %w", key, err)
	}
	return parsed, nil
}

func parseDynamoDBOptionalInt64(item map[string]map[string]string, key string) (int64, error) {
	raw := dynamoDBNumberValue(item, key)
	if raw == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("decode agent active assignment %s: %w", key, err)
	}
	return parsed, nil
}
