package riidoaiserver

import (
	"strconv"
	"time"
)

func blockedClaimActiveItem(now time.Time) map[string]map[string]string {
	activeExpiresAt := now.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second)
	return map[string]map[string]string{
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": "jykim-old"},
		"active_assignment_id":  {"S": "asn-000000"},
		"lease_token":           {"S": "lease-old"},
		"lease_heartbeat_at":    {"S": now.Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": activeExpiresAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(activeExpiresAt.UnixMilli(), 10)},
	}
}
