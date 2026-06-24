package main

type threadHistoryV3 struct {
	ReadEndpoint       endpointRule   `json:"read_endpoint"`
	ActionEndpoints    []endpointRule `json:"action_endpoints"`
	SSEEndpoint        endpointRule   `json:"sse_endpoint"`
	IdentityRules      []identityRule `json:"identity_rules"`
	GroupingRules      []string       `json:"grouping_rules"`
	AgentSnapshotRules []string       `json:"agent_snapshot_rules"`
	MessageRoles       []messageRole  `json:"message_roles"`
	SSEHandlingRules   []string       `json:"sse_handling_rules"`
	TerminalStates     []string       `json:"terminal_states"`
	Checklist          []string       `json:"checklist"`
}

type endpointRule struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`
	TruthRole   string `json:"truth_role"`
	Optimistic  bool   `json:"optimistic,omitempty"`
	RequestBody string `json:"request_body,omitempty"`
}

type identityRule struct {
	Name    string `json:"name"`
	Meaning string `json:"meaning"`
	Use     string `json:"use"`
}

type messageRole struct {
	Role    string `json:"role"`
	Meaning string `json:"meaning"`
}
