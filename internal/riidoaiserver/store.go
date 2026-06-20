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
		case createToolApprovalCmd:
			approval, err := s.handleCreateToolApproval(&state, msg.agentID, msg.req)
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- toolApprovalCreateResult{approval: approval, err: err}
		case listToolApprovalsCmd:
			approvals, err := s.handleListTaskToolApprovals(&state, msg.taskID)
			msg.reply <- toolApprovalListResult{approvals: approvals, err: err}
		case decideToolApprovalCmd:
			result, decision, err := s.handleDecideToolApproval(&state, msg.taskID, msg.decision)
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- toolApprovalDecisionResult{result: result, decision: decision, err: err}
		case readToolApprovalCmd:
			result, decision, mutated, err := s.handleReadToolApproval(&state, msg.agentID, msg.assignmentID, msg.approvalID)
			if err == nil && mutated {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- toolApprovalDecisionResult{result: result, decision: decision, mutated: mutated, err: err}
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
		case registerToolApprovalWaiterCmd:
			ch, id := s.handleRegisterToolApprovalWaiter(&state, msg.key)
			msg.reply <- registerToolApprovalWaiterResult{ch: ch, id: id}
		case unregisterToolApprovalWaiterCmd:
			s.handleUnregisterToolApprovalWaiter(&state, msg.key, msg.id)
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
		if existing, ok := duplicateAssignmentEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, req.Metadata); ok {
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

func duplicateAssignmentEvent(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	if existing, ok := duplicateAssignmentEventKey(state, taskID, assignmentID, agentID, eventType, metadata); ok {
		return existing, true
	}
	return duplicateThreadProgressEvent(state, taskID, assignmentID, agentID, eventType, metadata)
}

func duplicateAssignmentEventKey(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	key := strings.TrimSpace(metadata[metadatakeys.AssignmentEventKey.String()])
	if key == "" {
		return TaskEvent{}, false
	}
	for _, event := range slices.Backward(state.events[taskID]) {
		if event.AssignmentID != assignmentID ||
			event.AgentID != agentID ||
			event.Type != eventType ||
			strings.TrimSpace(event.Metadata[metadatakeys.AssignmentEventKey.String()]) != key {
			continue
		}
		return event, true
	}
	return TaskEvent{}, false
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
