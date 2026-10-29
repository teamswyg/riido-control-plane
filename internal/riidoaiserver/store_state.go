package riidoaiserver

type taskRecord struct {
	id                  string
	componentID         string
	currentAssignmentID string
}

type storeState struct {
	cliInstallationIntents  map[string]cliInstallationIntent
	cliInstallationLinks    map[string]cliInstallationLink
	devicePolicyRevisions   map[string]uint64
	executionTerminations   map[string]executionTerminationReceipt
	executionClaims         map[string]executionClaim
	tasks                   map[string]taskRecord
	assignments             map[string]Assignment
	agentAssignments        map[string][]string
	providerStatuses        map[string]ProviderStatusSyncResponse
	agentCatalog            map[string]AgentCatalogRecord
	events                  map[string][]TaskEvent
	assignmentEventACKs     map[string][]assignmentEventACKReceipt
	subscribers             map[string]map[int64]chan TaskEvent
	agentWaiters            map[string]map[int64]chan struct{}
	toolApprovals           map[string]ToolApprovalRequest
	toolApprovalDecisions   map[string]ToolApprovalDecision
	toolApprovalWaiters     map[string]map[int64]chan struct{}
	nextAssignmentSeq       int64
	nextEventSeq            int64
	nextSubscriberSeq       int64
	nextAgentWaiterSeq      int64
	nextToolApprovalWaiter  int64
	pollRequestsTotal       int64
	pollActionsTotal        map[PollAction]int64
	agentEventsTotal        int64
	providerStatusSyncTotal int64
	outboxErrorsTotal       int64
	eventAppendLatency      eventAppendLatencyMetrics
}

type eventAppendLatencyMetrics struct {
	samplesTotal      int64
	totalMilliseconds int64
	maxMilliseconds   int64
	lastMilliseconds  int64
}

func newStoreState() storeState {
	return storeState{
		cliInstallationIntents: map[string]cliInstallationIntent{},
		cliInstallationLinks:   map[string]cliInstallationLink{},
		devicePolicyRevisions:  map[string]uint64{},
		executionClaims:        map[string]executionClaim{},
		executionTerminations:  map[string]executionTerminationReceipt{},
		tasks:                  map[string]taskRecord{},
		assignments:            map[string]Assignment{},
		agentAssignments:       map[string][]string{},
		providerStatuses:       map[string]ProviderStatusSyncResponse{},
		agentCatalog:           map[string]AgentCatalogRecord{},
		events:                 map[string][]TaskEvent{},
		subscribers:            map[string]map[int64]chan TaskEvent{},
		agentWaiters:           map[string]map[int64]chan struct{}{},
		toolApprovals:          map[string]ToolApprovalRequest{},
		toolApprovalDecisions:  map[string]ToolApprovalDecision{},
		toolApprovalWaiters:    map[string]map[int64]chan struct{}{},
		pollActionsTotal:       map[PollAction]int64{},
	}
}
