package riidoaiserver

type taskRecord struct {
	id                  string
	componentID         string
	currentAssignmentID string
}

type storeState struct {
	tasks                   map[string]taskRecord
	assignments             map[string]Assignment
	agentAssignments        map[string][]string
	providerStatuses        map[string]ProviderStatusSyncResponse
	agentCatalog            map[string]AgentCatalogRecord
	events                  map[string][]TaskEvent
	subscribers             map[string]map[int64]chan TaskEvent
	nextAssignmentSeq       int64
	nextEventSeq            int64
	nextSubscriberSeq       int64
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
		tasks:            map[string]taskRecord{},
		assignments:      map[string]Assignment{},
		agentAssignments: map[string][]string{},
		providerStatuses: map[string]ProviderStatusSyncResponse{},
		agentCatalog:     map[string]AgentCatalogRecord{},
		events:           map[string][]TaskEvent{},
		subscribers:      map[string]map[int64]chan TaskEvent{},
		pollActionsTotal: map[PollAction]int64{},
	}
}
