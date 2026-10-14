package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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

	identity, err := s.newAssignmentID(state.nextAssignmentSeq + 1)
	if err != nil {
		return Assignment{}, err
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
		WorkspaceID:              req.WorkspaceID,
		ID:                       identity,
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
