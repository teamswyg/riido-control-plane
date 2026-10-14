package riidoaiserver

import (
	"strconv"
	"time"
)

func staleBlockedActiveItem(now time.Time) map[string]map[string]string {
	expiredAt := now.Add(-time.Minute)
	return map[string]map[string]string{
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": "jykim-old"},
		"active_assignment_id":  {"S": "asn-000001"},
		"lease_token":           {"S": "lease-old"},
		"lease_heartbeat_at":    {"S": now.Add(-2 * time.Minute).Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": expiredAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(expiredAt.UnixMilli(), 10)},
	}
}
