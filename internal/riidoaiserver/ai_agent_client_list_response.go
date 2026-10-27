package riidoaiserver

import "time"

type AgentClientListResponse struct {
	SchemaVersion string              `json:"schema_version"`
	Agents        []AgentClientRecord `json:"agents"`
	// Request-scoped projection for the authorized directory only; never wire or persistence data.
	agentDaemonLastSeenAt map[string]time.Time `json:"-"`
}
