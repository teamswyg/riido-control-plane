package riidoaiserver

type eventCmd struct {
	agentID string
	req     AgentEventRequest
	reply   chan eventResult
}

type eventResult struct {
	response AgentEventResponse
	err      error
}

type heartbeatCmd struct {
	agentID string
	req     AgentHeartbeatRequest
	reply   chan heartbeatResult
}

type heartbeatResult struct {
	response  AgentHeartbeatResponse
	mutations []heartbeatMutation
	err       error
}

type heartbeatMutation struct {
	assignment    Assignment
	operationType AssignmentOperationType
	events        []TaskEvent
}

type providerStatusCmd struct {
	agentID string
	req     ProviderStatusSyncRequest
	reply   chan providerStatusResult
}

type providerStatusResult struct {
	response ProviderStatusSyncResponse
	err      error
}

type getProviderStatusCmd struct {
	agentID string
	reply   chan getProviderStatusResult
}

type getProviderStatusResult struct {
	response ProviderStatusSyncResponse
	ok       bool
	err      error
}
