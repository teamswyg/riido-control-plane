package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

type Store struct {
	commands      chan any
	done          chan struct{}
	now           func() time.Time
	outbox        EventSink
	snapshotStore SnapshotStore
	agentRegistry AgentRegistry
}

type StoreConfig struct {
	Now           func() time.Time
	Outbox        EventSink
	SnapshotStore SnapshotStore
	AgentRegistry AgentRegistry
}

func NewStore() *Store {
	return NewStoreWithConfig(StoreConfig{})
}

func NewStoreWithClock(now func() time.Time) *Store {
	return NewStoreWithConfig(StoreConfig{Now: now})
}

func NewStoreWithConfig(config StoreConfig) *Store {
	return newStoreWithConfig(config, newStoreState())
}

func OpenStoreWithConfig(ctx context.Context, config StoreConfig) (*Store, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	state := newStoreState()
	if config.SnapshotStore != nil {
		snapshot, ok, err := config.SnapshotStore.LoadStoreSnapshot(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			loaded, err := stateFromSnapshot(snapshot)
			if err != nil {
				return nil, err
			}
			state = loaded
		}
	}
	return newStoreWithConfig(config, state), nil
}

func newStoreWithConfig(config StoreConfig, initial storeState) *Store {
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	s := &Store{
		commands:      make(chan any),
		done:          make(chan struct{}),
		now:           now,
		outbox:        config.Outbox,
		snapshotStore: config.SnapshotStore,
		agentRegistry: config.AgentRegistry,
	}
	go s.loop(initial)
	return s
}

func (s *Store) Close() {
	select {
	case <-s.done:
		return
	default:
	}
	reply := make(chan struct{})
	select {
	case s.commands <- closeCmd{reply: reply}:
		<-reply
	case <-s.done:
	}
}

func (s *Store) AssignTask(ctx context.Context, taskID string, req AssignRequest) (Assignment, error) {
	reply := make(chan assignResult, 1)
	if err := s.send(ctx, assignCmd{taskID: taskID, req: req, reply: reply}); err != nil {
		return Assignment{}, err
	}
	select {
	case res := <-reply:
		return res.assignment, res.err
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}

func (s *Store) PollAgent(ctx context.Context, agentID string, req PollRequest) (PollResponse, error) {
	reply := make(chan pollResult, 1)
	if err := s.send(ctx, pollCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return PollResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return PollResponse{}, ctx.Err()
	}
}

func (s *Store) HeartbeatAgent(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error) {
	reply := make(chan heartbeatResult, 1)
	if err := s.send(ctx, heartbeatCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentHeartbeatResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return AgentHeartbeatResponse{}, ctx.Err()
	}
}

func (s *Store) SyncProviderStatus(ctx context.Context, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error) {
	reply := make(chan providerStatusResult, 1)
	if err := s.send(ctx, providerStatusCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return ProviderStatusSyncResponse{}, ctx.Err()
	}
}

func (s *Store) GetProviderStatus(ctx context.Context, agentID string) (ProviderStatusSyncResponse, bool, error) {
	reply := make(chan getProviderStatusResult, 1)
	if err := s.send(ctx, getProviderStatusCmd{agentID: agentID, reply: reply}); err != nil {
		return ProviderStatusSyncResponse{}, false, err
	}
	select {
	case res := <-reply:
		return res.response, res.ok, res.err
	case <-ctx.Done():
		return ProviderStatusSyncResponse{}, false, ctx.Err()
	}
}

func (s *Store) RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (AgentEventResponse, error) {
	reply := make(chan eventResult, 1)
	if err := s.send(ctx, eventCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentEventResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return AgentEventResponse{}, ctx.Err()
	}
}

func (s *Store) SubscribeTask(ctx context.Context, taskID string) ([]TaskEvent, <-chan TaskEvent, func(), error) {
	reply := make(chan subscribeResult, 1)
	if err := s.send(ctx, subscribeCmd{taskID: taskID, reply: reply}); err != nil {
		return nil, nil, nil, err
	}
	select {
	case res := <-reply:
		if res.err != nil {
			return nil, nil, nil, res.err
		}
		cancel := func() {
			unsub := make(chan struct{}, 1)
			if err := s.send(context.Background(), unsubscribeCmd{taskID: taskID, subID: res.subID, reply: unsub}); err != nil {
				return
			}
			<-unsub
		}
		return res.history, res.events, cancel, nil
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	}
}

func (s *Store) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	reply := make(chan metricsResult, 1)
	if err := s.send(ctx, metricsCmd{reply: reply}); err != nil {
		return MetricsSnapshot{}, err
	}
	select {
	case res := <-reply:
		return res.snapshot, res.err
	case <-ctx.Done():
		return MetricsSnapshot{}, ctx.Err()
	}
}

func (s *Store) send(ctx context.Context, cmd any) error {
	select {
	case s.commands <- cmd:
		return nil
	case <-s.done:
		return errors.New("riido-control-plane store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

type assignCmd struct {
	taskID string
	req    AssignRequest
	reply  chan assignResult
}

type assignResult struct {
	assignment Assignment
	err        error
}

type pollCmd struct {
	agentID string
	req     PollRequest
	reply   chan pollResult
}

type pollResult struct {
	response PollResponse
	err      error
}

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
	response AgentHeartbeatResponse
	err      error
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

type subscribeCmd struct {
	taskID string
	reply  chan subscribeResult
}

type subscribeResult struct {
	history []TaskEvent
	events  <-chan TaskEvent
	subID   int64
	err     error
}

type unsubscribeCmd struct {
	taskID string
	subID  int64
	reply  chan struct{}
}

type metricsCmd struct {
	reply chan metricsResult
}

type metricsResult struct {
	snapshot MetricsSnapshot
	err      error
}

type closeCmd struct {
	reply chan struct{}
}

func (s *Store) loop(state storeState) {
	defer close(s.done)
	for cmd := range s.commands {
		switch msg := cmd.(type) {
		case assignCmd:
			assignment, err := s.handleAssign(&state, msg.taskID, msg.req)
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- assignResult{assignment: assignment, err: err}
		case pollCmd:
			response, err := s.handlePoll(&state, msg.agentID, msg.req)
			if err == nil && response.Action == PollStart {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- pollResult{response: response, err: err}
		case eventCmd:
			response, err := s.handleEvent(&state, msg.agentID, msg.req)
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- eventResult{response: response, err: err}
		case heartbeatCmd:
			response, err := s.handleHeartbeat(&state, msg.agentID, msg.req)
			msg.reply <- heartbeatResult{response: response, err: err}
		case providerStatusCmd:
			response, err := s.handleProviderStatusSync(&state, msg.agentID, msg.req)
			msg.reply <- providerStatusResult{response: response, err: err}
		case getProviderStatusCmd:
			response, ok, err := handleGetProviderStatus(&state, msg.agentID)
			msg.reply <- getProviderStatusResult{response: response, ok: ok, err: err}
		case subscribeCmd:
			history, events, subID, err := s.handleSubscribe(&state, msg.taskID)
			msg.reply <- subscribeResult{history: history, events: events, subID: subID, err: err}
		case unsubscribeCmd:
			s.handleUnsubscribe(&state, msg.taskID, msg.subID)
			msg.reply <- struct{}{}
		case metricsCmd:
			msg.reply <- metricsResult{snapshot: s.handleMetrics(&state)}
		case closeCmd:
			_ = s.saveSnapshot(&state)
			if s.outbox != nil {
				_ = s.outbox.Close()
			}
			if s.snapshotStore != nil {
				_ = s.snapshotStore.Close()
			}
			for _, subs := range state.subscribers {
				for _, ch := range subs {
					close(ch)
				}
			}
			close(msg.reply)
			return
		}
	}
}

func (s *Store) saveSnapshot(state *storeState) error {
	if s.snapshotStore == nil {
		return nil
	}
	return s.snapshotStore.SaveStoreSnapshot(context.Background(), snapshotFromState(state, s.now()))
}

func (s *Store) handleAssign(state *storeState, taskID string, req AssignRequest) (Assignment, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.RuntimeProvider = strings.TrimSpace(req.RuntimeProvider)
	req.Prompt = strings.TrimSpace(req.Prompt)
	if taskID == "" {
		return Assignment{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return Assignment{}, errors.New("agent_id is required")
	}
	if req.RuntimeProvider == "" {
		return Assignment{}, errors.New("runtime_provider is required")
	}
	if req.Prompt == "" {
		return Assignment{}, errors.New("prompt is required")
	}
	if err := validateAssignmentBinding(s.agentRegistry, req.AgentID, req.RuntimeProvider); err != nil {
		return Assignment{}, err
	}
	if status, ok := state.providerStatuses[req.AgentID]; ok {
		decision, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{
			RuntimeProvider:  capability.ProviderKind(req.RuntimeProvider),
			ProviderStatuses: cloneProviderStatusRecords(status.Providers),
		})
		if err != nil {
			return Assignment{}, err
		}
		if !decision.Allowed {
			return Assignment{}, fmt.Errorf("provider %s cannot be assigned: %s", req.RuntimeProvider, decision.Reason)
		}
	}

	now := s.now()
	task := state.tasks[taskID]
	if task.id == "" {
		task = taskRecord{id: taskID, componentID: req.ComponentID}
	}

	replacesID := ""
	blockedByID := ""
	if task.currentAssignmentID != "" {
		current := state.assignments[task.currentAssignmentID]
		if !isTerminal(current.State) {
			if current.AgentID == req.AgentID {
				return current, nil
			}
			current.State = AssignmentCancelling
			current.UpdatedAt = now
			state.assignments[current.ID] = current
			s.appendEvent(state, current.TaskID, current.ID, current.AgentID, EventAssignmentCancelling, current.State, "task reassigned to another agent", nil, now)
			replacesID = current.ID
			blockedByID = current.ID
		}
	}

	state.nextAssignmentSeq++
	assignment := Assignment{
		ID:                    fmt.Sprintf("asn-%06d", state.nextAssignmentSeq),
		TaskID:                taskID,
		ComponentID:           req.ComponentID,
		AgentID:               req.AgentID,
		RuntimeProvider:       req.RuntimeProvider,
		Prompt:                req.Prompt,
		State:                 AssignmentQueued,
		ReplacesAssignmentID:  replacesID,
		BlockedByAssignmentID: blockedByID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	state.assignments[assignment.ID] = assignment
	state.agentAssignments[assignment.AgentID] = append(state.agentAssignments[assignment.AgentID], assignment.ID)
	task.currentAssignmentID = assignment.ID
	task.componentID = req.ComponentID
	state.tasks[taskID] = task
	s.appendEvent(state, taskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "", nil, now)
	return assignment, nil
}

func (s *Store) handlePoll(state *storeState, agentID string, req PollRequest) (PollResponse, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return PollResponse{}, errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, req); err != nil {
		return PollResponse{}, err
	}
	state.pollRequestsTotal++
	response := PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}
	assignIDs := state.agentAssignments[agentID]
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State == AssignmentCancelling {
			response.Action = PollCancel
			response.Assignment = copyAssignment(assignment)
			state.pollActionsTotal[response.Action]++
			return response, nil
		}
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if isAgentActive(assignment.State) {
			response.Action = PollActive
			response.Assignment = copyAssignment(assignment)
			state.pollActionsTotal[response.Action]++
			return response, nil
		}
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State != AssignmentQueued || !assignmentBlockerCleared(state, assignment) {
			continue
		}
		now := s.now()
		assignment.State = AssignmentLeased
		assignment.LeaseToken = fmt.Sprintf("%s:%d", assignment.ID, now.UnixNano())
		assignment.UpdatedAt = now
		state.assignments[id] = assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentLeased, assignment.State, "", map[string]string{"lease_token": assignment.LeaseToken}, now)
		response.Action = PollStart
		response.Assignment = copyAssignment(assignment)
		state.pollActionsTotal[response.Action]++
		return response, nil
	}
	state.pollActionsTotal[response.Action]++
	return response, nil
}

func (s *Store) handleHeartbeat(state *storeState, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentHeartbeatResponse{}, errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return AgentHeartbeatResponse{}, err
	}
	assignmentIDs := heartbeatAssignmentIDs(state, agentID, req)
	response := AgentHeartbeatResponse{SchemaVersion: SchemaVersion}
	if len(assignmentIDs) == 0 {
		return response, nil
	}
	now := s.now()
	for _, assignmentID := range assignmentIDs {
		assignment, ok := state.assignments[assignmentID]
		if !ok {
			return AgentHeartbeatResponse{}, fmt.Errorf("assignment %s not found", assignmentID)
		}
		if assignment.AgentID != agentID {
			return AgentHeartbeatResponse{}, fmt.Errorf("assignment %s belongs to agent %s", assignmentID, assignment.AgentID)
		}
		if !assignmentHoldsActiveLease(assignment.State) {
			continue
		}
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = assignment
		response.RefreshedAssignments = append(response.RefreshedAssignments, assignment)
	}
	return response, nil
}

func heartbeatAssignmentIDs(state *storeState, agentID string, req AgentHeartbeatRequest) []string {
	seen := map[string]bool{}
	var ids []string
	appendID := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	for _, id := range req.ActiveAssignmentIDs {
		appendID(id)
	}
	for _, taskID := range req.RunningTaskIDs {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			continue
		}
		for _, assignmentID := range state.agentAssignments[agentID] {
			assignment := state.assignments[assignmentID]
			if assignment.TaskID == taskID && assignmentHoldsActiveLease(assignment.State) {
				appendID(assignment.ID)
			}
		}
	}
	return ids
}

func (s *Store) handleEvent(state *storeState, agentID string, req AgentEventRequest) (AgentEventResponse, error) {
	agentID = strings.TrimSpace(agentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	if agentID == "" {
		return AgentEventResponse{}, errors.New("agent_id is required")
	}
	if req.AssignmentID == "" {
		return AgentEventResponse{}, errors.New("assignment_id is required")
	}
	assignment, ok := state.assignments[req.AssignmentID]
	if !ok {
		return AgentEventResponse{}, fmt.Errorf("assignment %s not found", req.AssignmentID)
	}
	if assignment.AgentID != agentID {
		return AgentEventResponse{}, fmt.Errorf("assignment %s belongs to agent %s", req.AssignmentID, assignment.AgentID)
	}
	if req.TaskID != "" && req.TaskID != assignment.TaskID {
		return AgentEventResponse{}, fmt.Errorf("assignment %s belongs to task %s", req.AssignmentID, assignment.TaskID)
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return AgentEventResponse{}, err
	}
	state.agentEventsTotal++
	now := s.now()
	if req.State != "" && req.State != assignment.State {
		if !canTransitionAssignment(assignment.State, req.State) {
			return AgentEventResponse{}, fmt.Errorf("invalid assignment transition %s -> %s", assignment.State, req.State)
		}
		assignment.State = req.State
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = assignment
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		if req.State != "" {
			eventType = EventAssignmentStateUpdated
		}
		if eventType == "" {
			eventType = EventRiidoLog
		}
	}
	event := s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, assignment.State, req.Message, req.Metadata, now)
	return AgentEventResponse{
		SchemaVersion: SchemaVersion,
		Assignment:    copyAssignment(assignment),
		Event:         event,
	}, nil
}

func (s *Store) handleProviderStatusSync(state *storeState, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error) {
	agentID = strings.TrimSpace(agentID)
	req, err := normalizeProviderStatusSync(agentID, req)
	if err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	state.providerStatusSyncTotal++
	response := ProviderStatusSyncResponse{
		SchemaVersion:       SchemaVersion,
		AgentID:             agentID,
		DaemonID:            req.DaemonID,
		DeviceID:            req.DeviceID,
		RuntimeID:           req.RuntimeID,
		DistributionChannel: req.DistributionChannel,
		AppVersion:          req.AppVersion,
		Providers:           cloneProviderStatusRecords(req.Providers),
		SyncedAt:            s.now(),
	}
	state.providerStatuses[agentID] = response
	return cloneProviderStatusResponse(response), nil
}

func handleGetProviderStatus(state *storeState, agentID string) (ProviderStatusSyncResponse, bool, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return ProviderStatusSyncResponse{}, false, errors.New("agent_id is required")
	}
	response, ok := state.providerStatuses[agentID]
	if !ok {
		return ProviderStatusSyncResponse{}, false, nil
	}
	return cloneProviderStatusResponse(response), true, nil
}

func (s *Store) handleSubscribe(state *storeState, taskID string) ([]TaskEvent, <-chan TaskEvent, int64, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, nil, 0, errors.New("task_id is required")
	}
	state.nextSubscriberSeq++
	subID := state.nextSubscriberSeq
	ch := make(chan TaskEvent, 32)
	if state.subscribers[taskID] == nil {
		state.subscribers[taskID] = map[int64]chan TaskEvent{}
	}
	state.subscribers[taskID][subID] = ch
	history := append([]TaskEvent(nil), state.events[taskID]...)
	return history, ch, subID, nil
}

func (s *Store) handleUnsubscribe(state *storeState, taskID string, subID int64) {
	subs := state.subscribers[taskID]
	if subs == nil {
		return
	}
	if ch, ok := subs[subID]; ok {
		close(ch)
		delete(subs, subID)
	}
	if len(subs) == 0 {
		delete(state.subscribers, taskID)
	}
}

func (s *Store) handleMetrics(state *storeState) MetricsSnapshot {
	assignmentsByState := map[AssignmentState]int{}
	for _, assignment := range state.assignments {
		assignmentsByState[assignment.State]++
	}
	pollActions := make(map[PollAction]int64, len(state.pollActionsTotal))
	for action, count := range state.pollActionsTotal {
		pollActions[action] = count
	}
	return MetricsSnapshot{
		SchemaVersion:                       MetricsSchemaVersion,
		GeneratedAt:                         s.now(),
		TasksTotal:                          len(state.tasks),
		AssignmentsTotal:                    len(state.assignments),
		AssignmentsByState:                  assignmentsByState,
		PollRequestsTotal:                   state.pollRequestsTotal,
		PollActionsTotal:                    pollActions,
		AgentEventsTotal:                    state.agentEventsTotal,
		TaskEventsTotal:                     countTaskEvents(state),
		SSESubscribers:                      countSubscribers(state),
		OutboxErrorsTotal:                   state.outboxErrorsTotal,
		EventAppendLatencySamplesTotal:      state.eventAppendLatency.samplesTotal,
		EventAppendLatencyTotalMilliseconds: state.eventAppendLatency.totalMilliseconds,
		EventAppendLatencyMaxMilliseconds:   state.eventAppendLatency.maxMilliseconds,
		EventAppendLatencyLastMilliseconds:  state.eventAppendLatency.lastMilliseconds,
	}
}

func (s *Store) appendEvent(state *storeState, taskID, assignmentID, agentID, eventType string, assignmentState AssignmentState, message string, metadata map[string]string, at time.Time) TaskEvent {
	state.nextEventSeq++
	event := TaskEvent{
		Seq:          state.nextEventSeq,
		TaskID:       taskID,
		AssignmentID: assignmentID,
		AgentID:      agentID,
		Type:         eventType,
		State:        assignmentState,
		Message:      message,
		Metadata:     cloneMetadata(metadata),
		At:           at,
	}
	state.events[taskID] = append(state.events[taskID], event)
	if s.outbox != nil {
		s.appendTaskEventToOutbox(state, event)
	}
	for _, ch := range state.subscribers[taskID] {
		select {
		case ch <- event:
		default:
		}
	}
	return event
}

func (s *Store) appendTaskEventToOutbox(state *storeState, event TaskEvent) {
	startedAt := s.now()
	err := s.outbox.AppendTaskEvent(context.Background(), event)
	recordEventAppendLatency(state, s.now().Sub(startedAt))
	if err != nil {
		state.outboxErrorsTotal++
	}
}

func recordEventAppendLatency(state *storeState, duration time.Duration) {
	if duration < 0 {
		duration = 0
	}
	milliseconds := duration.Milliseconds()
	state.eventAppendLatency.samplesTotal++
	state.eventAppendLatency.totalMilliseconds += milliseconds
	state.eventAppendLatency.lastMilliseconds = milliseconds
	if milliseconds > state.eventAppendLatency.maxMilliseconds {
		state.eventAppendLatency.maxMilliseconds = milliseconds
	}
}

func assignmentHoldsActiveLease(state AssignmentState) bool {
	return isAgentActive(state) || state == AssignmentCancelling
}

func assignmentBlockerCleared(state *storeState, assignment Assignment) bool {
	if assignment.BlockedByAssignmentID == "" {
		return true
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	return isTerminal(blocker.State)
}

func copyAssignment(a Assignment) *Assignment {
	cp := a
	return &cp
}

func cloneMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneProviderStatusRecords(in []ProviderStatusRecord) []ProviderStatusRecord {
	if len(in) == 0 {
		return nil
	}
	return append([]ProviderStatusRecord(nil), in...)
}

func cloneProviderStatusResponse(in ProviderStatusSyncResponse) ProviderStatusSyncResponse {
	in.Providers = cloneProviderStatusRecords(in.Providers)
	return in
}

func countTaskEvents(state *storeState) int64 {
	var total int64
	for _, events := range state.events {
		total += int64(len(events))
	}
	return total
}

func countSubscribers(state *storeState) int {
	total := 0
	for _, subs := range state.subscribers {
		total += len(subs)
	}
	return total
}

var _ AssignmentStore = (*Store)(nil)
var _ ProviderStatusStore = (*Store)(nil)
var _ ProviderStatusReader = (*Store)(nil)
