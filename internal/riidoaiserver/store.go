package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
			assignment, err := s.handleAssign(&state, msg.taskID, msg.req, msg.allowConcurrentTaskAgents, msg.forceNewAssignment)
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
			result, decision, approval, mutated, err := s.handleReadToolApproval(&state, msg.agentID, msg.assignmentID, msg.approvalID)
			if err == nil && mutated {
				err = s.saveSnapshot(&state)
			}
			msg.reply <- toolApprovalDecisionResult{approval: approval, result: result, decision: decision, mutated: mutated, err: err}
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

func (s *Store) handleAssign(state *storeState, taskID string, req AssignRequest, allowConcurrentTaskAgents, forceNewAssignment bool) (Assignment, error) {
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
			if current.AgentID == req.AgentID && !forceNewAssignment {
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
		if response, ok := deletedAgentPollCancelResponse(state, agentID, req, count); ok {
			return response, false, nil, "", nil
		}
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
