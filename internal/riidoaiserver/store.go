package riidoaiserver

import (
	"strings"
	"time"
)

func (s *Store) loop(state storeState) {
	defer close(s.done)
	for cmd := range s.commands {
		switch msg := cmd.(type) {
		case recoverTaskAgentPublicationCmd:
			msg.reply <- s.applyTaskAgentPublicationRecovery(&state, msg.ctx, msg.assignment)
		case assignCmd:
			if msg.allowConcurrentTaskAgents && strings.TrimSpace(msg.req.RequestID) != "" {
				before := state.nextAssignmentSeq
				assignment, err := s.assignIdentifiedRequest(&state, msg.taskID, msg.req)
				msg.creationReceipt.record(assignment, err == nil && state.nextAssignmentSeq != before)
				msg.reply <- assignResult{assignment: assignment, err: err}
				continue
			}
			startedAt := time.Now()
			_, taskExisted := state.tasks[msg.taskID]
			beforeEventSeq := state.nextEventSeq
			beforeAssignmentSeq := state.nextAssignmentSeq
			assignment, err := s.handleAssign(&state, msg.taskID, msg.req, msg.allowConcurrentTaskAgents, msg.forceNewAssignment)
			msg.creationReceipt.record(assignment, err == nil && state.nextAssignmentSeq != beforeAssignmentSeq)
			if err == nil {
				err = s.saveAssignmentMutationOperations(&state, AssignmentOperationAssignTask, assignment, eventsAfterSeq(&state, beforeEventSeq), state.nextAssignmentSeq != beforeAssignmentSeq)
			}
			if err == nil {
				err = s.saveSnapshot(&state)
			}
			if err == nil && !taskExisted {
				s.observeStoreOperation(StoreOperationCreateTask, startedAt, nil)
			}
			msg.reply <- assignResult{assignment: assignment, err: err}
		case cancelAssignmentCmd:
			if _, durable := s.operationStore.(AssignmentProjectionReader); durable && msg.req.AssignmentID != "" {
				assignment, err := s.mutateDurableAssignment(msg.ctx, &state, msg.req.AssignmentID, AssignmentOperationClientStop, func(quiet *Store, trial *storeState) (Assignment, error) {
					return quiet.handleCancelAssignment(trial, msg.taskID, msg.req)
				})
				msg.reply <- cancelAssignmentResult{assignment: assignment, err: err}
				continue
			}
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
			if err == nil {
				for _, event := range eventsAfterSeq(&state, beforeEventSeq) {
					logAssignmentFailure(&state, event)
				}
			}
			msg.reply <- pollResult{response: response, operationAlreadySaved: operationAlreadySaved, mutatedAssignment: mutatedAssignment, mutationOperationType: mutationOperationType, err: err}
		case eventCmd:
			if err := msg.ctx.Err(); err != nil {
				msg.reply <- eventResult{err: err}
				continue
			}
			if _, durable := s.operationStore.(AssignmentProjectionReader); durable {
				var response AgentEventResponse
				_, err := s.mutateDurableAssignment(msg.ctx, &state, msg.req.AssignmentID, AssignmentOperationAgentEvent, func(quiet *Store, trial *storeState) (Assignment, error) {
					var err error
					quiet.recoveryApproval = msg.approval
					if err = msg.ctx.Err(); err != nil {
						return Assignment{}, err
					}
					response, err = quiet.handleEvent(trial, msg.agentID, msg.req)
					if response.Assignment == nil {
						return Assignment{}, err
					}
					return *response.Assignment, err
				}, func(event TaskEvent) { observeCommittedAssignmentEvent(msg.ctx, &state, event) })
				msg.reply <- eventResult{response: response, err: err}
				continue
			}
			beforeEventSeq := state.nextEventSeq
			eventStore := *s
			eventStore.recoveryApproval = msg.approval
			response, err := eventStore.handleEvent(&state, msg.agentID, msg.req)
			if err == nil && state.nextEventSeq != beforeEventSeq {
				err = s.saveOperation(&state, AssignmentOperationAgentEvent, *response.Assignment, eventsAfterSeq(&state, beforeEventSeq))
				if err == nil {
					err = s.saveSnapshot(&state)
				}
			}
			if err == nil && state.nextEventSeq != beforeEventSeq {
				observeCommittedAssignmentEvent(msg.ctx, &state, response.Event)
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
		case agentDashboardSnapshotCmd:
			msg.reply <- agentDashboardSnapshotFromState(&state, msg.agentID, s.now())
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
