package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) prepareAIAgentMembershipRemoval(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID, reason string) (*taskAgentMembershipRemovalFence, error) {
	// Preserve fixture adapters without durable cancellation. Production injects
	// an AssignmentCancellationStore; its captured targets and fence must agree.
	if _, ok := s.assignment.(AssignmentCancellationStore); !ok {
		if err := s.validateAIAgentMembershipRemoval(ctx, principal, taskID, agentID, assignmentID); err != nil {
			return nil, err
		}
		return nil, nil
	}
	taskID, agentID, assignmentID = strings.TrimSpace(taskID), strings.TrimSpace(agentID), strings.TrimSpace(assignmentID)
	var threads AIAgentTaskThreadCollectionResponse
	var observation *taskAgentMembershipRemovalFence
	var err error
	if reader, ok := s.aiAgent.(taskAgentMembershipObservationReader); ok {
		threads, observation, err = reader.observeTaskAgentMembershipRemoval(ctx, principal, taskID, agentID)
	} else {
		threads, err = s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	}
	if err != nil {
		return nil, err
	}
	selector := assignmentID
	if observation != nil && observation.durableGeneration != "" && isNonDurableAssignmentID(assignmentID) && observation.intentGateAssignmentID != assignmentID {
		return nil, ErrAIAgentMembershipChanged
	}
	if observation != nil && observation.durableGeneration != "" && !isNonDurableAssignmentID(assignmentID) {
		if assignmentID != "" && strings.SplitN(observation.durableGeneration, "@", 2)[0] != assignmentID {
			return nil, ErrAIAgentMembershipChanged
		}
		selector = ""
	}
	targets, found, err := assignmentActionTargetsFromThreads(taskID, agentID, selector, threads.Threads)
	if err != nil {
		return nil, err
	}
	if agentID == "" && found {
		agentID = targets[0].AgentID
	}
	if err := validateMembershipRemovalAssignment(agentID, selector, threads.Threads); err != nil {
		return nil, err
	}
	if observation == nil {
		observation = observeTaskAgentMembershipRemoval(taskID, agentID, threads.Threads)
	} else if observation.agentID != agentID {
		reader := s.aiAgent.(taskAgentMembershipObservationReader)
		threads, observation, err = reader.observeTaskAgentMembershipRemoval(ctx, principal, taskID, agentID)
		if err != nil {
			return nil, err
		}
	}
	// The observed ID guards membership removal; cancellation covers this member.
	targets = membershipRemovalActiveTargets(taskID, agentID, threads.Threads)
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() {
		reader := s.assignment.(taskAgentDurableExecutionReader)
		generation, active, err := reader.loadTaskAgentDurableExecutions(ctx, taskID, agentID)
		if err != nil {
			return nil, err
		}
		if generation != observation.durableGeneration || (assignmentID != "" && !isNonDurableAssignmentID(assignmentID) && generation != "" && strings.SplitN(generation, "@", 2)[0] != assignmentID) {
			return nil, ErrAIAgentMembershipChanged
		}
		legacy, err := s.assignment.(interface {
			loadLegacyTaskAgentDurableExecutions(context.Context, string, string) ([]AssignmentProjection, error)
		}).loadLegacyTaskAgentDurableExecutions(ctx, taskID, agentID)
		if err != nil {
			return nil, err
		}
		active = append(active, legacy...)
		readerProjection := s.assignment.(AssignmentProjectionReader)
		for _, target := range targets {
			projection, found, err := readerProjection.LoadAssignmentProjection(ctx, target.AssignmentID)
			if err != nil {
				return nil, err
			}
			if found && projection.Assignment.TaskID == taskID && projection.Assignment.AgentID == agentID && !isTerminal(projection.Assignment.State) {
				active = append(active, projection)
			}
		}
		targets = nil
		seen := map[string]bool{}
		for _, projection := range active {
			if seen[projection.Assignment.ID] {
				continue
			}
			seen[projection.Assignment.ID] = true
			targets = append(targets, aiAgentAssignmentActionTarget{TaskID: taskID, AgentID: agentID, AssignmentID: projection.Assignment.ID, State: projection.Assignment.State, durableProjection: &projection})
		}
	}
	if len(targets) > 0 {
		if _, _, err := s.cancelAIAgentAssignmentTargets(ctx, targets, reason); err != nil {
			return nil, err
		}
	}
	return observation, nil
}

// ListAIAgentTaskThreads hides queued state from clients; its empty state still
// identifies a cancellable retained execution. Terminal history is excluded.
func membershipRemovalActiveTargets(taskID, agentID string, threads []AIAgentTaskThreadRecord) []aiAgentAssignmentActionTarget {
	targets := []aiAgentAssignmentActionTarget{}
	seen := map[string]bool{}
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if thread.AgentID != agentID || thread.AssignmentID == "" || seen[thread.AssignmentID] {
			continue
		}
		if thread.AssignmentState != "" && !taskThreadHasActiveStream(thread) {
			continue
		}
		seen[thread.AssignmentID] = true
		targets = append(targets, aiAgentAssignmentActionTarget{TaskID: taskID, AgentID: agentID, AssignmentID: thread.AssignmentID})
	}
	return targets
}
