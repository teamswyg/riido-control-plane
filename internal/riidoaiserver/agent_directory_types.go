package riidoaiserver

import "time"

// The agent directory lets a running agent compare other agents before it
// mentions one (RIID-5892). It is read live, never from the assignment
// snapshot, and never carries private instructions, owners or credentials.
const (
	agentDirectoryPageSize = 50

	AgentDirectoryRoleDescribed = "described"
	AgentDirectoryRoleMissing   = "missing"

	AgentDirectoryConnectionOnline  = "online"
	AgentDirectoryConnectionOffline = "offline"

	AgentDirectoryReasonSelf           = "self"
	AgentDirectoryReasonRuntimeOffline = "runtime_offline"
)

type AgentDirectoryRuntime struct {
	Kind       RuntimeKind `json:"kind,omitempty"`
	Connection string      `json:"connection"`
}

// AgentDirectoryEntry separates work status (what the agent is doing) from
// runtime connection (whether its device can take a run). Callable reflects
// only the checks listed in UnavailableReasons; mention-time policy checks
// stay with the server that starts the run.
type AgentDirectoryEntry struct {
	AgentID            string                `json:"agent_id"`
	Name               string                `json:"name"`
	Description        string                `json:"description,omitempty"`
	RoleStatus         string                `json:"role_status"`
	WorkStatus         AgentWorkStatus       `json:"work_status"`
	Runtime            AgentDirectoryRuntime `json:"runtime"`
	Callable           bool                  `json:"callable"`
	UnavailableReasons []string              `json:"unavailable_reasons,omitempty"`
	IsCurrentAgent     bool                  `json:"is_current_agent,omitempty"`
	UpdatedAt          time.Time             `json:"updated_at,omitzero"`
}

type AgentDirectoryPage struct {
	Agents        []AgentDirectoryEntry `json:"agents"`
	Total         int                   `json:"total"`
	Cursor        string                `json:"cursor"`
	NextCursor    string                `json:"next_cursor,omitempty"`
	FetchedAt     time.Time             `json:"fetched_at"`
	PolicyChecked bool                  `json:"policy_checked"`
}

type AgentDirectoryDetail struct {
	Agent         AgentDirectoryEntry `json:"agent"`
	FetchedAt     time.Time           `json:"fetched_at"`
	PolicyChecked bool                `json:"policy_checked"`
}
