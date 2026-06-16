package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

type Store struct {
	commands            chan any
	done                chan struct{}
	now                 func() time.Time
	activeLeaseDuration time.Duration
	outbox              EventSink
	snapshotStore       SnapshotStore
	operationStore      AssignmentOperationStore
	agentRegistry       AgentRegistry
	operationMetrics    *StoreOperationMetrics
	traceRecorder       TraceRecorder
}

type StoreConfig struct {
	Now                 func() time.Time
	ActiveLeaseDuration time.Duration
	Outbox              EventSink
	SnapshotStore       SnapshotStore
	OperationStore      AssignmentOperationStore
	AgentRegistry       AgentRegistry
	OperationMetrics    *StoreOperationMetrics
	TraceRecorder       TraceRecorder
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
	loadedSnapshot := false
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
			loadedSnapshot = true
		}
	}
	if !loadedSnapshot {
		if loader, ok := config.OperationStore.(AssignmentOperationLoader); ok {
			operations, err := loader.LoadAssignmentOperations(ctx)
			if err != nil {
				return nil, err
			}
			if len(operations) > 0 {
				loaded, err := stateFromAssignmentOperations(operations)
				if err != nil {
					return nil, err
				}
				state = loaded
			}
		}
	}
	return newStoreWithConfig(config, state), nil
}

func newStoreWithConfig(config StoreConfig, initial storeState) *Store {
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	activeLeaseDuration := config.ActiveLeaseDuration
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	s := &Store{
		commands:            make(chan any, 64),
		done:                make(chan struct{}),
		now:                 now,
		activeLeaseDuration: activeLeaseDuration,
		outbox:              config.Outbox,
		snapshotStore:       config.SnapshotStore,
		operationStore:      config.OperationStore,
		agentRegistry:       config.AgentRegistry,
		operationMetrics:    config.OperationMetrics,
		traceRecorder:       config.TraceRecorder,
	}
	if s.operationMetrics == nil {
		s.operationMetrics = NewStoreOperationMetrics()
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

func (s *Store) AssignTask(ctx context.Context, taskID string, req AssignRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCreateAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCreateAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
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

func (s *Store) AssignTaskAdditive(ctx context.Context, taskID string, req AssignRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCreateAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCreateAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan assignResult, 1)
	if err := s.send(ctx, assignCmd{taskID: taskID, req: req, allowConcurrentTaskAgents: true, reply: reply}); err != nil {
		return Assignment{}, err
	}
	select {
	case res := <-reply:
		return res.assignment, res.err
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}

func (s *Store) CancelAssignment(ctx context.Context, taskID string, req CancelAssignmentRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCancelAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCancelAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan cancelAssignmentResult, 1)
	if err := s.send(ctx, cancelAssignmentCmd{taskID: taskID, req: req, reply: reply}); err != nil {
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
	return s.pollAgent(ctx, agentID, req, true)
}

// pollAgent runs a single point-in-time evaluation. count records whether this
// evaluation should be tallied as a daemon poll request (true for a real
// short-poll or a long-poll's first evaluation; false for the internal
// re-evaluations a long-poll performs while held).
func (s *Store) pollAgent(ctx context.Context, agentID string, req PollRequest, count bool) (response PollResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationPollAssignment)
	startedAt := time.Now()
	defer func() {
		if err == nil {
			span.SetAttributes(StringTraceAttribute(metadatakeys.RiidoPollAction.String(), string(response.Action)))
		}
		if !count {
			FinishTraceSpan(span, err)
			return
		}
		s.observeStoreOperation(StoreOperationPollAssignment, startedAt, err)
		if err == nil && response.Action == PollStart {
			s.observeStoreOperation(StoreOperationLeaseAssignment, startedAt, nil)
		}
		FinishTraceSpan(span, err)
	}()
	reply := make(chan pollResult, 1)
	if err := s.send(ctx, pollCmd{agentID: agentID, req: req, countRequest: count, reply: reply}); err != nil {
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
	response, _, err := s.HeartbeatAgentWithEvents(ctx, agentID, req)
	return response, err
}

func (s *Store) HeartbeatAgentWithEvents(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []TaskEvent, error) {
	reply := make(chan heartbeatResult, 1)
	if err := s.send(ctx, heartbeatCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentHeartbeatResponse{}, nil, err
	}
	select {
	case res := <-reply:
		if res.err != nil {
			return AgentHeartbeatResponse{}, nil, res.err
		}
		var events []TaskEvent
		for _, mutation := range res.mutations {
			events = append(events, mutation.events...)
		}
		return res.response, events, nil
	case <-ctx.Done():
		return AgentHeartbeatResponse{}, nil, ctx.Err()
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

func (s *Store) RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (response AgentEventResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationAppendEvent)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationAppendEvent, startedAt, err)
		FinishTraceSpan(span, err)
	}()
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

// WaitForAssignment is the long-poll claim. It returns immediately when work is
// already available for the agent; otherwise it registers a per-agent waiter and
// blocks until a queued assignment is signaled, a cross-instance re-poll tick
// finds one (covers assignments queued on another control-plane instance), the
// hold budget elapses (returns action=none, like a normal empty poll), or ctx is
// cancelled (client disconnect / shutdown).
//
// The wait loop runs entirely on the caller's goroutine — never on the store
// actor loop — so the actor keeps servicing assign/heartbeat/poll commands while
// a request is held.
func (s *Store) WaitForAssignment(ctx context.Context, agentID string, req PollRequest, hold, tick time.Duration) (response PollResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationWaitAssignment)
	startedAt := time.Now()
	defer func() {
		if err == nil {
			span.SetAttributes(StringTraceAttribute(metadatakeys.RiidoPollAction.String(), string(response.Action)))
		}
		s.observeStoreOperation(StoreOperationWaitAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	// First evaluation is the one that counts as a daemon poll request and
	// short-circuits when work is already queued.
	resp, err := s.PollAgent(ctx, agentID, req)
	if err != nil || resp.Action != PollNone {
		return resp, err
	}
	if hold <= 0 {
		return resp, nil
	}

	signal, release, err := s.registerWaiter(ctx, agentID)
	if err != nil {
		return PollResponse{}, err
	}
	defer release()

	// Re-evaluate AFTER registering so an assignment queued between the first
	// evaluation and registration is not missed: registration runs on the actor
	// goroutine, so any signal fired after it is guaranteed to reach this waiter,
	// and anything queued before it is caught here. This closes the lost-wakeup
	// window. Internal re-evaluations pass count=false so they do not inflate the
	// daemon-poll-request metric.
	resp, err = s.pollAgent(ctx, agentID, req, false)
	if err != nil || resp.Action != PollNone {
		return resp, err
	}

	if tick <= 0 || tick > hold {
		tick = hold
	}
	deadline := time.NewTimer(hold)
	defer deadline.Stop()
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-signal:
			if resp, err = s.pollAgent(ctx, agentID, req, false); err != nil || resp.Action != PollNone {
				return resp, err
			}
		case <-ticker.C:
			if resp, err = s.pollAgent(ctx, agentID, req, false); err != nil || resp.Action != PollNone {
				return resp, err
			}
		case <-deadline.C:
			return PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}, nil
		case <-ctx.Done():
			return PollResponse{}, ctx.Err()
		}
	}
}

func (s *Store) registerWaiter(ctx context.Context, agentID string) (<-chan struct{}, func(), error) {
	reply := make(chan registerWaiterResult, 1)
	if err := s.send(ctx, registerWaiterCmd{agentID: agentID, reply: reply}); err != nil {
		return nil, nil, err
	}
	select {
	case res := <-reply:
		release := func() {
			done := make(chan struct{}, 1)
			if err := s.send(context.Background(), unregisterWaiterCmd{agentID: agentID, id: res.id, reply: done}); err != nil {
				return
			}
			<-done
		}
		return res.ch, release, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
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

func (s *Store) LoadAssignmentProjection(ctx context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("assignment_id is required")
	}
	if reader, ok := s.operationStore.(AssignmentProjectionReader); ok {
		return reader.LoadAssignmentProjection(ctx, assignmentID)
	}
	reply := make(chan assignmentProjectionResult, 1)
	if err := s.send(ctx, assignmentProjectionCmd{assignmentID: assignmentID, reply: reply}); err != nil {
		return AssignmentProjection{}, false, err
	}
	select {
	case res := <-reply:
		return res.projection, res.found, res.err
	case <-ctx.Done():
		return AssignmentProjection{}, false, ctx.Err()
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

func (s *Store) startStoreOperationTrace(ctx context.Context, operation StoreOperationName) (context.Context, TraceSpan) {
	return StartTraceSpan(ctx, s.traceRecorder, TraceSpanStart{
		Name: "store." + operation.String(),
		Kind: TraceSpanKindInternal,
		Attributes: []TraceAttribute{
			StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), operation.String()),
			StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "assignment_store"),
		},
	})
}

type assignCmd struct {
	taskID                    string
	req                       AssignRequest
	allowConcurrentTaskAgents bool
	reply                     chan assignResult
}

type assignResult struct {
	assignment Assignment
	err        error
}

type CancelAssignmentRequest struct {
	AgentID      string
	AssignmentID string
	Reason       string
}

type cancelAssignmentCmd struct {
	taskID string
	req    CancelAssignmentRequest
	reply  chan cancelAssignmentResult
}

type cancelAssignmentResult struct {
	assignment Assignment
	err        error
}

type pollCmd struct {
	agentID string
	req     PollRequest
	// countRequest increments pollRequestsTotal for this evaluation. A daemon
	// long-poll counts exactly once (its first evaluation) so the metric still
	// means "daemon poll requests", not internal re-evaluations during a hold.
	countRequest bool
	reply        chan pollResult
}

type pollResult struct {
	response              PollResponse
	operationAlreadySaved bool
	mutatedAssignment     *Assignment
	mutationOperationType AssignmentOperationType
	err                   error
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

// registerWaiterCmd / unregisterWaiterCmd implement per-agent long-poll waiters,
// mirroring the per-task subscriber pub/sub. A waiter is a buffered (cap 1)
// signal channel woken when an assignment becomes claimable for the agent.
type registerWaiterCmd struct {
	agentID string
	reply   chan registerWaiterResult
}

type registerWaiterResult struct {
	ch chan struct{}
	id int64
}

type unregisterWaiterCmd struct {
	agentID string
	id      int64
	reply   chan struct{}
}

type metricsCmd struct {
	reply chan metricsResult
}

type metricsResult struct {
	snapshot MetricsSnapshot
	err      error
}

type assignmentProjectionCmd struct {
	assignmentID string
	reply        chan assignmentProjectionResult
}

type assignmentProjectionResult struct {
	projection AssignmentProjection
	found      bool
	err        error
}

type closeCmd struct {
	reply chan struct{}
}

func (s *Store) loop(state storeState) {
	defer close(s.done)
	for cmd := range s.commands {
		switch msg := cmd.(type) {
		case assignCmd:
			startedAt := time.Now()
			_, taskExisted := state.tasks[msg.taskID]
			beforeEventSeq := state.nextEventSeq
			assignment, err := s.handleAssign(&state, msg.taskID, msg.req, msg.allowConcurrentTaskAgents)
			if err == nil {
				err = s.saveAssignmentMutationOperations(&state, AssignmentOperationAssignTask, assignment, eventsAfterSeq(&state, beforeEventSeq))
			}
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			if err == nil && !taskExisted {
				s.observeStoreOperation(StoreOperationCreateTask, startedAt, nil)
			}
			msg.reply <- assignResult{assignment: assignment, err: err}
		case cancelAssignmentCmd:
			beforeEventSeq := state.nextEventSeq
			assignment, err := s.handleCancelAssignment(&state, msg.taskID, msg.req)
			if err == nil {
				err = s.saveOperation(&state, AssignmentOperationClientStop, assignment, eventsAfterSeq(&state, beforeEventSeq))
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			msg.reply <- cancelAssignmentResult{assignment: assignment, err: err}
		case pollCmd:
			beforeEventSeq := state.nextEventSeq
			response, operationAlreadySaved, mutatedAssignment, mutationOperationType, err := s.handlePoll(&state, msg.agentID, msg.req, msg.countRequest)
			if err == nil && mutatedAssignment != nil {
				err = s.saveOperation(&state, mutationOperationType, *mutatedAssignment, eventsAfterSeq(&state, beforeEventSeq))
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			if err == nil && mutatedAssignment == nil && response.Action == PollStart {
				if !operationAlreadySaved {
					err = s.saveAssignmentMutationOperations(&state, AssignmentOperationPollStart, *response.Assignment, eventsAfterSeq(&state, beforeEventSeq))
				}
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			msg.reply <- pollResult{response: response, operationAlreadySaved: operationAlreadySaved, mutatedAssignment: mutatedAssignment, mutationOperationType: mutationOperationType, err: err}
		case eventCmd:
			beforeEventSeq := state.nextEventSeq
			response, err := s.handleEvent(&state, msg.agentID, msg.req)
			if err == nil && state.nextEventSeq != beforeEventSeq {
				err = s.saveOperation(&state, AssignmentOperationAgentEvent, *response.Assignment, eventsAfterSeq(&state, beforeEventSeq))
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			msg.reply <- eventResult{response: response, err: err}
		case heartbeatCmd:
			response, mutations, err := s.handleHeartbeat(&state, msg.agentID, msg.req)
			if err == nil && len(mutations) > 0 {
				for _, mutation := range mutations {
					err = s.saveOperation(&state, mutation.operationType, mutation.assignment, mutation.events)
					if err != nil {
						break
					}
				}
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			msg.reply <- heartbeatResult{response: response, mutations: mutations, err: err}
		case providerStatusCmd:
			response, err := s.handleProviderStatusSync(&state, msg.agentID, msg.req)
			msg.reply <- providerStatusResult{response: response, err: err}
		case getProviderStatusCmd:
			response, ok, err := handleGetProviderStatus(&state, msg.agentID)
			msg.reply <- getProviderStatusResult{response: response, ok: ok, err: err}
		case listAgentCatalogCmd:
			msg.reply <- listAgentCatalogResult{records: handleListAgentCatalog(&state)}
		case getAgentCatalogCmd:
			record, ok, err := handleGetAgentCatalog(&state, msg.agentID)
			msg.reply <- getAgentCatalogResult{record: record, ok: ok, err: err}
		case saveAgentCatalogCmd:
			record, err := handleSaveAgentCatalog(&state, msg.record)
			msg.reply <- saveAgentCatalogResult{record: record, err: err}
		case deleteAgentCatalogCmd:
			deleted, err := handleDeleteAgentCatalog(&state, msg.agentID)
			msg.reply <- deleteAgentCatalogResult{deleted: deleted, err: err}
		case applyReviewAccountProvisioningCmd:
			err := s.handleApplyReviewAccountProvisioning(&state, msg.provisioning)
			msg.reply <- applyReviewAccountProvisioningResult{err: err}
		case subscribeCmd:
			history, events, subID, err := s.handleSubscribe(&state, msg.taskID)
			msg.reply <- subscribeResult{history: history, events: events, subID: subID, err: err}
		case unsubscribeCmd:
			s.handleUnsubscribe(&state, msg.taskID, msg.subID)
			msg.reply <- struct{}{}
		case registerWaiterCmd:
			ch, id := s.handleRegisterWaiter(&state, msg.agentID)
			msg.reply <- registerWaiterResult{ch: ch, id: id}
		case unregisterWaiterCmd:
			s.handleUnregisterWaiter(&state, msg.agentID, msg.id)
			msg.reply <- struct{}{}
		case metricsCmd:
			msg.reply <- metricsResult{snapshot: s.handleMetrics(&state)}
		case assignmentProjectionCmd:
			projection, found, err := handleLoadAssignmentProjection(&state, msg.assignmentID)
			msg.reply <- assignmentProjectionResult{projection: projection, found: found, err: err}
		case closeCmd:
			_ = s.saveSnapshot(&state)
			if s.outbox != nil {
				_ = s.outbox.Close()
			}
			if s.snapshotStore != nil {
				_ = s.snapshotStore.Close()
			}
			if s.operationStore != nil {
				_ = s.operationStore.Close()
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

func (s *Store) saveOperation(state *storeState, operationType AssignmentOperationType, assignment Assignment, events []TaskEvent) error {
	if s.operationStore == nil || len(events) == 0 {
		return nil
	}
	recordedAt := s.now()
	return s.operationStore.SaveAssignmentOperation(context.Background(), AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(operationType, assignment, events),
		OperationType: operationType,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        append([]TaskEvent(nil), events...),
		RecordedAt:    recordedAt,
	})
}

func (s *Store) handleAssign(state *storeState, taskID string, req AssignRequest, allowConcurrentTaskAgents bool) (Assignment, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.RuntimeProvider = strings.TrimSpace(req.RuntimeProvider)
	req.ModelID = strings.TrimSpace(req.ModelID)
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.AgentInstruction = strings.TrimSpace(req.AgentInstruction)
	req.ResumeSessionID = strings.TrimSpace(req.ResumeSessionID)
	req.Worktree = normalizeAssignmentWorktree(req.Worktree)
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
	if utf8.RuneCountInString(req.AgentInstruction) > AgentInstructionMaxCharacters {
		return Assignment{}, errors.New("agent_instruction must be 1000 characters or fewer")
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

	if allowConcurrentTaskAgents {
		for _, assignmentID := range state.agentAssignments[req.AgentID] {
			current := state.assignments[assignmentID]
			if current.TaskID == taskID && !isTerminal(current.State) {
				s.cancelQueuedBlockerForAssignment(state, &current, now)
				return current, nil
			}
		}
	}

	replacesID := ""
	blockedByID := ""
	if !allowConcurrentTaskAgents && task.currentAssignmentID != "" {
		current := state.assignments[task.currentAssignmentID]
		if !isTerminal(current.State) {
			if current.AgentID == req.AgentID {
				s.cancelQueuedBlockerForAssignment(state, &current, now)
				return current, nil
			}
			replacesID = current.ID
			if current.State.Code() == AssignmentStateCodeQueued {
				current.State = AssignmentCancelled
				current.UpdatedAt = now
				state.assignments[current.ID] = current
				s.appendEvent(state, current.TaskID, current.ID, current.AgentID, EventAssignmentCancelled, current.State, "queued assignment was replaced before daemon lease", nil, now)
			} else {
				current.State = AssignmentCancelling
				current.UpdatedAt = now
				state.assignments[current.ID] = current
				s.appendEvent(state, current.TaskID, current.ID, current.AgentID, EventAssignmentCancelling, current.State, "task reassigned to another agent", nil, now)
				blockedByID = current.ID
			}
		}
	}

	state.nextAssignmentSeq++
	assignment := Assignment{
		ID:                       fmt.Sprintf("asn-%06d", state.nextAssignmentSeq),
		TaskID:                   taskID,
		ComponentID:              req.ComponentID,
		AgentID:                  req.AgentID,
		RuntimeProvider:          req.RuntimeProvider,
		ModelID:                  req.ModelID,
		Prompt:                   req.Prompt,
		AgentInstruction:         req.AgentInstruction,
		AllowExperimentalRuntime: req.AllowExperimentalRuntime,
		ResumeSessionID:          req.ResumeSessionID,
		Worktree:                 cloneAssignmentWorktree(req.Worktree),
		State:                    AssignmentQueued,
		ReplacesAssignmentID:     replacesID,
		BlockedByAssignmentID:    blockedByID,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	state.assignments[assignment.ID] = assignment
	state.agentAssignments[assignment.AgentID] = append(state.agentAssignments[assignment.AgentID], assignment.ID)
	task.currentAssignmentID = assignment.ID
	task.componentID = req.ComponentID
	state.tasks[taskID] = task
	s.appendEvent(state, taskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "", nil, now)
	s.signalAgentWaiters(state, assignment.AgentID)
	return assignment, nil
}

func normalizeAssignmentWorktree(worktree *AssignmentWorktree) *AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := &AssignmentWorktree{
		RepositoryFullName: safeAIAgentRepositoryFullName(worktree.RepositoryFullName),
		RepositoryURL:      safeAIAgentRepositoryURL(worktree.RepositoryURL),
		BranchName:         strings.TrimSpace(worktree.BranchName),
		IsPrivate:          worktree.IsPrivate,
		Source:             strings.TrimSpace(worktree.Source),
	}
	if out.RepositoryFullName == "" && out.RepositoryURL == "" {
		return nil
	}
	return out
}

func cloneAssignmentWorktree(worktree *AssignmentWorktree) *AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := *worktree
	return &out
}

func (s *Store) handleCancelAssignment(state *storeState, taskID string, req CancelAssignmentRequest) (Assignment, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.Reason = strings.TrimSpace(req.Reason)
	if taskID == "" {
		return Assignment{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return Assignment{}, errors.New("agent_id is required")
	}
	assignment, ok := state.assignmentForClientStop(taskID, req)
	if !ok {
		if req.AssignmentID != "" {
			return Assignment{}, fmt.Errorf("assignment %s not found", req.AssignmentID)
		}
		return Assignment{}, fmt.Errorf("active assignment for task %s and agent %s not found", taskID, req.AgentID)
	}
	if assignment.TaskID != taskID {
		return Assignment{}, fmt.Errorf("assignment %s belongs to task %s", assignment.ID, assignment.TaskID)
	}
	if assignment.AgentID != req.AgentID {
		return Assignment{}, fmt.Errorf("assignment %s belongs to agent %s", assignment.ID, assignment.AgentID)
	}
	if isTerminal(assignment.State) || assignment.State.Code() == AssignmentStateCodeCancelling {
		return assignment, nil
	}

	nextState := AssignmentCancelling
	eventType := EventAssignmentCancelling
	message := req.Reason
	if message == "" {
		message = "assignment cancellation requested by client"
	}
	if assignment.State.Code() == AssignmentStateCodeQueued {
		nextState = AssignmentCancelled
		eventType = EventAssignmentCancelled
		if req.Reason == "" {
			message = "queued assignment was cancelled by client"
		}
	}
	if !canTransitionAssignment(assignment.State, nextState) {
		return Assignment{}, fmt.Errorf("invalid assignment transition %s -> %s", assignment.State, nextState)
	}
	now := s.now()
	assignment.State = nextState
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, assignment.State, message, map[string]string{"source": "client_stop"}, now)
	// Wake a daemon parked on a long-poll so it observes the cancel (PollCancel)
	// within the hold instead of waiting for the budget to elapse.
	s.signalAgentWaiters(state, assignment.AgentID)
	return assignment, nil
}

func (s *Store) saveAssignmentMutationOperations(state *storeState, primaryType AssignmentOperationType, primary Assignment, events []TaskEvent) error {
	if len(events) == 0 {
		return nil
	}
	eventsByAssignment := map[string][]TaskEvent{}
	var assignmentIDs []string
	for _, event := range events {
		assignmentID := strings.TrimSpace(event.AssignmentID)
		if assignmentID == "" {
			continue
		}
		if _, ok := eventsByAssignment[assignmentID]; !ok {
			assignmentIDs = append(assignmentIDs, assignmentID)
		}
		eventsByAssignment[assignmentID] = append(eventsByAssignment[assignmentID], event)
	}
	for _, assignmentID := range assignmentIDs {
		assignment := state.assignments[assignmentID]
		if assignment.ID == "" {
			return fmt.Errorf("assignment %s not found for mutation events", assignmentID)
		}
		operationType := AssignmentOperationAgentEvent
		if assignmentID == primary.ID {
			operationType = primaryType
			assignment = primary
		}
		if err := s.saveOperation(state, operationType, assignment, eventsByAssignment[assignmentID]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) cancelQueuedBlockerForAssignment(state *storeState, assignment *Assignment, now time.Time) bool {
	if assignment == nil || strings.TrimSpace(assignment.BlockedByAssignmentID) == "" {
		return false
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	if blocker.ID == "" || blocker.State.Code() != AssignmentStateCodeQueued {
		return false
	}
	blocker.State = AssignmentCancelled
	blocker.UpdatedAt = now
	state.assignments[blocker.ID] = blocker
	s.appendEvent(state, blocker.TaskID, blocker.ID, blocker.AgentID, EventAssignmentCancelled, blocker.State, "queued blocker was cancelled before daemon lease", nil, now)

	assignment.BlockedByAssignmentID = ""
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = *assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "queued blocker cleared before daemon lease", nil, now)
	s.signalAgentWaiters(state, assignment.AgentID)
	return true
}

func (s *Store) handlePoll(state *storeState, agentID string, req PollRequest, count bool) (PollResponse, bool, *Assignment, AssignmentOperationType, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return PollResponse{}, false, nil, "", errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, req); err != nil {
		return PollResponse{}, false, nil, "", err
	}
	if count {
		state.pollRequestsTotal++
	}
	response := PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}
	assignIDs := state.agentAssignments[agentID]
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State.Code() == AssignmentStateCodeCancelling {
			expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
			if err != nil {
				return PollResponse{}, false, nil, "", err
			}
			if expired {
				stale := s.failStaleAssignment(state, assignment)
				recordPollAction(state, response.Action, count)
				return response, false, &stale, AssignmentOperationAgentEvent, nil
			}
			response.Action = PollCancel
			response.Assignment = copyAssignment(assignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	durableAssignment, staleAssignment, ok, err := s.loadDurableActiveAssignment(state, agentID, s.now())
	if err != nil {
		return PollResponse{}, false, nil, "", err
	}
	if staleAssignment != nil {
		recordPollAction(state, response.Action, count)
		return response, false, staleAssignment, AssignmentOperationAgentEvent, nil
	}
	if ok {
		switch {
		case durableAssignment.State.Code() == AssignmentStateCodeCancelling:
			response.Action = PollCancel
			response.Assignment = copyAssignment(durableAssignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		case isAgentActive(durableAssignment.State):
			response.Action = PollActive
			response.Assignment = copyAssignment(durableAssignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if isAgentActive(assignment.State) {
			expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
			if err != nil {
				return PollResponse{}, false, nil, "", err
			}
			if expired {
				stale := s.failStaleAssignment(state, assignment)
				recordPollAction(state, response.Action, count)
				return response, false, &stale, AssignmentOperationAgentEvent, nil
			}
			response.Action = PollActive
			response.Assignment = copyAssignment(assignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	if claimer, ok := s.operationStore.(AssignmentClaimer); ok {
		claim, err := claimer.ClaimNextAssignment(context.Background(), agentID, s.now())
		if err != nil {
			return PollResponse{}, false, nil, "", err
		}
		if claim.Claimed {
			if err := s.applyClaimedAssignment(state, claim); err != nil {
				return PollResponse{}, false, nil, "", err
			}
			response.Action = PollStart
			response.Assignment = copyAssignment(claim.Assignment)
			recordPollAction(state, response.Action, count)
			return response, true, nil, "", nil
		}
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State.Code() != AssignmentStateCodeQueued {
			continue
		}
		if err := s.repairQueuedAssignmentBlockerForClaim(state, &assignment); err != nil {
			return PollResponse{}, false, nil, "", err
		}
		if !assignmentBlockerCleared(state, assignment) {
			continue
		}
		now := s.now()
		assignment.BlockedByAssignmentID = ""
		assignment.State = AssignmentLeased
		assignment.LeaseToken = fmt.Sprintf("%s:%d", assignment.ID, now.UnixNano())
		assignment.UpdatedAt = now
		state.assignments[id] = assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentLeased, assignment.State, "", map[string]string{"lease_token": assignment.LeaseToken}, now)
		response.Action = PollStart
		response.Assignment = copyAssignment(assignment)
		recordPollAction(state, response.Action, count)
		return response, false, nil, "", nil
	}
	recordPollAction(state, response.Action, count)
	return response, false, nil, "", nil
}

func recordPollAction(state *storeState, action PollAction, count bool) {
	if count {
		state.pollActionsTotal[action]++
	}
}

func (s *Store) loadDurableActiveAssignment(state *storeState, agentID string, at time.Time) (Assignment, *Assignment, bool, error) {
	leaseStore, hasLeaseStore := s.operationStore.(AssignmentActiveLeaseStore)
	projectionReader, hasProjectionReader := s.operationStore.(AssignmentProjectionReader)
	if !hasLeaseStore || !hasProjectionReader {
		return Assignment{}, nil, false, nil
	}
	lease, exists, err := leaseStore.LoadAgentActiveAssignment(context.Background(), agentID)
	if err != nil {
		return Assignment{}, nil, false, err
	}
	if !exists || lease.ActiveAssignmentID == "" {
		return Assignment{}, nil, false, nil
	}
	projection, exists, err := projectionReader.LoadAssignmentProjection(context.Background(), lease.ActiveAssignmentID)
	if err != nil {
		return Assignment{}, nil, false, err
	}
	if !exists {
		return Assignment{}, nil, false, nil
	}
	assignment := projection.Assignment
	if assignment.AgentID != agentID {
		return Assignment{}, nil, false, nil
	}
	applyAssignmentProjectionToState(state, projection)
	if !assignmentHoldsActiveLease(assignment.State) {
		return assignment, nil, true, nil
	}
	if lease.Expired(at) {
		stale := s.failStaleAssignment(state, assignment)
		return Assignment{}, &stale, false, nil
	}
	if !lease.HeartbeatAt.IsZero() && lease.HeartbeatAt.After(assignment.UpdatedAt) {
		assignment.UpdatedAt = lease.HeartbeatAt
		state.assignments[assignment.ID] = assignment
	}
	return assignment, nil, true, nil
}

func (s *Store) assignmentActiveLeaseExpired(state *storeState, assignment Assignment, at time.Time) (bool, error) {
	if !assignmentHoldsActiveLease(assignment.State) {
		return false, nil
	}
	if !assignment.UpdatedAt.IsZero() && assignment.UpdatedAt.Add(s.activeLeaseDuration).After(at) {
		return false, nil
	}
	leaseStore, ok := s.operationStore.(AssignmentActiveLeaseStore)
	if !ok {
		return false, nil
	}
	lease, exists, err := leaseStore.LoadAgentActiveAssignment(context.Background(), assignment.AgentID)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	if lease.ActiveAssignmentID != assignment.ID {
		return true, nil
	}
	if lease.Expired(at) {
		return true, nil
	}
	if !lease.HeartbeatAt.IsZero() {
		assignment.UpdatedAt = lease.HeartbeatAt
		state.assignments[assignment.ID] = assignment
	}
	return false, nil
}

func (s *Store) failStaleAssignment(state *storeState, assignment Assignment) Assignment {
	return s.failStaleAssignmentWithMessage(state, assignment, "active assignment lease expired", nil)
}

func (s *Store) failStaleAssignmentWithMessage(state *storeState, assignment Assignment, message string, metadata map[string]string) Assignment {
	now := s.now()
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	eventMetadata := map[string]string{"lease_token": assignment.LeaseToken}
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			eventMetadata[key] = value
		}
	}
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentFailed, assignment.State, message, eventMetadata, now)
	return assignment
}

func (s *Store) repairQueuedAssignmentBlockerForClaim(state *storeState, assignment *Assignment) error {
	if assignment == nil || assignment.State.Code() != AssignmentStateCodeQueued || strings.TrimSpace(assignment.BlockedByAssignmentID) == "" {
		return nil
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	if blocker.ID == "" {
		blockedByID := assignment.BlockedByAssignmentID
		now := s.now()
		assignment.BlockedByAssignmentID = ""
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = *assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "missing blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": blockedByID}, now)
		s.signalAgentWaiters(state, assignment.AgentID)
		return nil
	}
	if isTerminal(blocker.State) {
		return nil
	}
	now := s.now()
	if blocker.State.Code() == AssignmentStateCodeQueued {
		blocker.State = AssignmentCancelled
		blocker.UpdatedAt = now
		state.assignments[blocker.ID] = blocker
		s.appendEvent(state, blocker.TaskID, blocker.ID, blocker.AgentID, EventAssignmentCancelled, blocker.State, "queued blocker was cancelled before queued assignment claim", map[string]string{"blocked_assignment_id": assignment.ID}, now)
		assignment.BlockedByAssignmentID = ""
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = *assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "queued blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": blocker.ID}, now)
		s.signalAgentWaiters(state, assignment.AgentID)
		return nil
	}
	if !assignmentHoldsActiveLease(blocker.State) {
		return nil
	}
	expired, err := s.assignmentActiveLeaseExpired(state, blocker, now)
	if err != nil {
		return err
	}
	if !expired {
		return nil
	}
	stale := s.failStaleAssignmentWithMessage(state, blocker, "blocked queued assignment repaired after stale blocker lease expired", map[string]string{
		"blocked_assignment_id": assignment.ID,
	})
	assignment.BlockedByAssignmentID = ""
	assignment.UpdatedAt = stale.UpdatedAt
	state.assignments[assignment.ID] = *assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "stale blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": stale.ID}, stale.UpdatedAt)
	s.signalAgentWaiters(state, assignment.AgentID)
	return nil
}

func (s *Store) handleHeartbeat(state *storeState, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []heartbeatMutation, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentHeartbeatResponse{}, nil, errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return AgentHeartbeatResponse{}, nil, err
	}
	assignmentIDs := heartbeatAssignmentIDs(state, agentID, req)
	response := AgentHeartbeatResponse{SchemaVersion: SchemaVersion}
	if len(assignmentIDs) == 0 {
		return response, nil, nil
	}
	now := s.now()
	var mutations []heartbeatMutation
	leaseStore, _ := s.operationStore.(AssignmentActiveLeaseStore)
	for _, assignmentID := range assignmentIDs {
		assignment, ok := state.assignments[assignmentID]
		if !ok {
			return AgentHeartbeatResponse{}, nil, fmt.Errorf("assignment %s not found", assignmentID)
		}
		if assignment.AgentID != agentID {
			return AgentHeartbeatResponse{}, nil, fmt.Errorf("assignment %s belongs to agent %s", assignmentID, assignment.AgentID)
		}
		if !assignmentHoldsActiveLease(assignment.State) {
			continue
		}
		expired, err := s.assignmentActiveLeaseExpired(state, assignment, now)
		if err != nil {
			return AgentHeartbeatResponse{}, nil, err
		}
		if expired {
			beforeEventSeq := state.nextEventSeq
			stale := s.failStaleAssignment(state, assignment)
			mutations = append(mutations, heartbeatMutation{
				assignment:    stale,
				operationType: AssignmentOperationAgentEvent,
				events:        eventsAfterSeq(state, beforeEventSeq),
			})
			continue
		}
		if leaseStore != nil {
			if err := leaseStore.RefreshAgentActiveAssignment(context.Background(), assignment, now); err != nil {
				return AgentHeartbeatResponse{}, nil, err
			}
		}
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = assignment
		response.RefreshedAssignments = append(response.RefreshedAssignments, assignment)
	}
	return response, mutations, nil
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
	req.ProviderSessionID = strings.TrimSpace(req.ProviderSessionID)
	assignmentChanged := false
	if req.State != "" && req.State != assignment.State {
		if !canTransitionAssignment(assignment.State, req.State) {
			return AgentEventResponse{}, fmt.Errorf("invalid assignment transition %s -> %s", assignment.State, req.State)
		}
		assignment.State = req.State
		assignment.UpdatedAt = now
		assignmentChanged = true
	}
	if req.ProviderSessionID != "" && req.ProviderSessionID != assignment.ProviderSessionID {
		assignment.ProviderSessionID = req.ProviderSessionID
		assignment.UpdatedAt = now
		assignmentChanged = true
	}
	if assignmentChanged {
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
	if !assignmentChanged {
		if existing, ok := duplicateThreadProgressEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, req.Metadata); ok {
			return AgentEventResponse{
				SchemaVersion: SchemaVersion,
				Assignment:    copyAssignment(assignment),
				Event:         existing,
			}, nil
		}
	}
	event := s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, assignment.State, req.Message, req.Metadata, now)
	return AgentEventResponse{
		SchemaVersion: SchemaVersion,
		Assignment:    copyAssignment(assignment),
		Event:         event,
	}, nil
}

func duplicateThreadProgressEvent(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	if eventType != EventRiidoLog {
		return TaskEvent{}, false
	}
	seq := strings.TrimSpace(metadata[metadatakeys.ThreadProgressSeq.String()])
	if seq == "" {
		return TaskEvent{}, false
	}
	for _, event := range slices.Backward(state.events[taskID]) {
		if event.AssignmentID != assignmentID ||
			event.AgentID != agentID ||
			event.Type != EventRiidoLog ||
			strings.TrimSpace(event.Metadata[metadatakeys.ThreadProgressSeq.String()]) != seq {
			continue
		}
		return event, true
	}
	return TaskEvent{}, false
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

// handleRegisterWaiter / handleUnregisterWaiter / signalAgentWaiters run only on
// the actor goroutine, so all access to state.agentWaiters is serialized and
// signal sends never race a register/unregister. The signal channel is buffered
// (cap 1) so a wake between two waiter selects is never lost; we never close it
// (the waiting goroutine only receives), so a stray late signal cannot panic.
func (s *Store) handleRegisterWaiter(state *storeState, agentID string) (chan struct{}, int64) {
	state.nextAgentWaiterSeq++
	id := state.nextAgentWaiterSeq
	ch := make(chan struct{}, 1)
	if state.agentWaiters[agentID] == nil {
		state.agentWaiters[agentID] = map[int64]chan struct{}{}
	}
	state.agentWaiters[agentID][id] = ch
	return ch, id
}

func (s *Store) handleUnregisterWaiter(state *storeState, agentID string, id int64) {
	waiters := state.agentWaiters[agentID]
	if waiters == nil {
		return
	}
	delete(waiters, id)
	if len(waiters) == 0 {
		delete(state.agentWaiters, agentID)
	}
}

// signalAgentWaiters wakes every long-poll parked on agentID. It is a one-shot,
// non-blocking broadcast: a waiter whose buffer is already full is left as-is
// (it will re-evaluate anyway). Called from the queued-producing transitions.
func (s *Store) signalAgentWaiters(state *storeState, agentID string) {
	for _, ch := range state.agentWaiters[agentID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *Store) handleMetrics(state *storeState) MetricsSnapshot {
	rebuildStateMetricsFromHistory(state)
	assignmentsByState := map[AssignmentState]int{}
	for _, assignment := range state.assignments {
		assignmentsByState[assignment.State]++
	}
	pollActions := make(map[PollAction]int64, len(state.pollActionsTotal))
	maps.Copy(pollActions, state.pollActionsTotal)
	snapshot := MetricsSnapshot{
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
	return s.operationMetrics.ApplyToMetricsSnapshot(snapshot)
}

func (s *Store) observeStoreOperation(operation StoreOperationName, startedAt time.Time, err error) {
	if s == nil || s.operationMetrics == nil {
		return
	}
	s.operationMetrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: operation,
		Duration:  time.Since(startedAt),
		Err:       err,
	})
}

func (s *Store) appendEvent(state *storeState, taskID, assignmentID, agentID, eventType string, assignmentState AssignmentState, message string, metadata map[string]string, at time.Time) TaskEvent {
	startedAt := time.Now()
	defer func() {
		recordEventAppendLatency(state, time.Since(startedAt))
	}()
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

func (s *Store) applyClaimedAssignment(state *storeState, claim AssignmentClaimResult) error {
	if !claim.Claimed {
		return nil
	}
	if claim.Assignment.ID == "" {
		return errors.New("claimed assignment_id is required")
	}
	if claim.Operation.OperationID == "" {
		return errors.New("claimed assignment operation_id is required")
	}
	if err := validateAssignmentOperationRecord(claim.Operation); err != nil {
		return err
	}
	if claim.Operation.OperationType != AssignmentOperationPollStart {
		return fmt.Errorf("claimed assignment operation_type = %q", claim.Operation.OperationType)
	}
	if claim.Operation.AssignmentID != claim.Assignment.ID {
		return fmt.Errorf("claimed assignment operation assignment_id mismatch %q", claim.Operation.AssignmentID)
	}
	if claim.Operation.Assignment != claim.Assignment {
		return errors.New("claimed assignment operation assignment mismatch")
	}
	operations := claim.Operations
	if len(operations) == 0 {
		operations = []AssignmentOperationRecord{claim.Operation}
	}
	sawPrimary := false
	for _, operation := range operations {
		if err := validateAssignmentOperationRecord(operation); err != nil {
			return err
		}
		if operation.OperationID == claim.Operation.OperationID {
			sawPrimary = true
		}
		applyAssignmentToState(state, operation.Assignment)
		for _, event := range operation.Events {
			s.appendRecordedEvent(state, event)
		}
	}
	if !sawPrimary {
		return errors.New("claimed assignment operations missing primary claim operation")
	}
	return nil
}

func applyAssignmentToState(state *storeState, assignment Assignment) {
	state.assignments[assignment.ID] = assignment
	if seq := assignmentSequence(assignment.ID); seq > state.nextAssignmentSeq {
		state.nextAssignmentSeq = seq
	}
	if !assignmentIDInAgentQueue(state.agentAssignments[assignment.AgentID], assignment.ID) {
		state.agentAssignments[assignment.AgentID] = append(state.agentAssignments[assignment.AgentID], assignment.ID)
	}
	task := state.tasks[assignment.TaskID]
	current := state.assignments[task.currentAssignmentID]
	if task.id == "" || assignmentIsNewer(assignment, current) || task.currentAssignmentID == assignment.ID {
		state.tasks[assignment.TaskID] = taskRecord{
			id:                  assignment.TaskID,
			componentID:         assignment.ComponentID,
			currentAssignmentID: assignment.ID,
		}
	}
}

func applyAssignmentProjectionToState(state *storeState, projection AssignmentProjection) {
	applyAssignmentToState(state, projection.Assignment)
	if projection.LastEventSeq > state.nextEventSeq {
		state.nextEventSeq = projection.LastEventSeq
	}
}

func handleLoadAssignmentProjection(state *storeState, assignmentID string) (AssignmentProjection, bool, error) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("assignment_id is required")
	}
	assignment, ok := state.assignments[assignmentID]
	if !ok {
		return AssignmentProjection{}, false, nil
	}
	lastEventSeq := int64(0)
	for _, event := range state.events[assignment.TaskID] {
		if event.AssignmentID == assignmentID && event.Seq > lastEventSeq {
			lastEventSeq = event.Seq
		}
	}
	return AssignmentProjection{Assignment: assignment, LastEventSeq: lastEventSeq}, true, nil
}

func assignmentIDInAgentQueue(ids []string, id string) bool {
	return slices.Contains(ids, id)
}

func (s *Store) appendRecordedEvent(state *storeState, event TaskEvent) {
	startedAt := time.Now()
	defer func() {
		recordEventAppendLatency(state, time.Since(startedAt))
	}()
	if event.Seq <= 0 || event.TaskID == "" {
		return
	}
	for _, existing := range state.events[event.TaskID] {
		if existing.Seq == event.Seq {
			return
		}
	}
	recorded := event
	recorded.Metadata = cloneMetadata(event.Metadata)
	state.events[recorded.TaskID] = append(state.events[recorded.TaskID], recorded)
	sort.Slice(state.events[recorded.TaskID], func(i, j int) bool {
		return state.events[recorded.TaskID][i].Seq < state.events[recorded.TaskID][j].Seq
	})
	if recorded.Seq > state.nextEventSeq {
		state.nextEventSeq = recorded.Seq
	}
	if recorded.AssignmentID != "" && recorded.State != "" {
		assignment := state.assignments[recorded.AssignmentID]
		if assignment.ID != "" {
			assignment.State = recorded.State
			if !recorded.At.IsZero() {
				assignment.UpdatedAt = recorded.At
			}
			state.assignments[assignment.ID] = assignment
		}
	}
	if s.outbox != nil {
		s.appendTaskEventToOutbox(state, recorded)
	}
	for _, ch := range state.subscribers[recorded.TaskID] {
		select {
		case ch <- recorded:
		default:
		}
	}
}

func (s *Store) appendTaskEventToOutbox(state *storeState, event TaskEvent) {
	err := s.outbox.AppendTaskEvent(context.Background(), event)
	if err != nil {
		state.outboxErrorsTotal++
	}
}

func recordEventAppendLatency(state *storeState, duration time.Duration) {
	milliseconds := durationMilliseconds(duration)
	state.eventAppendLatency.samplesTotal++
	state.eventAppendLatency.totalMilliseconds += milliseconds
	state.eventAppendLatency.lastMilliseconds = milliseconds
	if milliseconds > state.eventAppendLatency.maxMilliseconds {
		state.eventAppendLatency.maxMilliseconds = milliseconds
	}
}

func assignmentHoldsActiveLease(state AssignmentState) bool {
	code := state.Code()
	return code.IsAgentActive() || code == AssignmentStateCodeCancelling
}

func assignmentBlockerCleared(state *storeState, assignment Assignment) bool {
	if assignment.BlockedByAssignmentID == "" {
		return true
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	return isTerminal(blocker.State)
}

func (state *storeState) assignmentForClientStop(taskID string, req CancelAssignmentRequest) (Assignment, bool) {
	if req.AssignmentID != "" {
		assignment := state.assignments[req.AssignmentID]
		return assignment, assignment.ID != ""
	}
	assignmentIDs := state.agentAssignments[req.AgentID]
	for _, assignmentID := range slices.Backward(assignmentIDs) {
		assignment := state.assignments[assignmentID]
		if assignment.TaskID != taskID || isTerminal(assignment.State) {
			continue
		}
		return assignment, true
	}
	return Assignment{}, false
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
	maps.Copy(out, in)
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

func eventsAfterSeq(state *storeState, seq int64) []TaskEvent {
	var events []TaskEvent
	for _, taskEvents := range state.events {
		for _, event := range taskEvents {
			if event.Seq > seq {
				events = append(events, event)
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Seq < events[j].Seq })
	return events
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

var (
	_ AssignmentStore      = (*Store)(nil)
	_ ProviderStatusStore  = (*Store)(nil)
	_ ProviderStatusReader = (*Store)(nil)
)
