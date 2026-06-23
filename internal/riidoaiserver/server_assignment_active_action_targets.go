package riidoaiserver

import "strings"

func preferredActiveAssignmentActionTarget(taskID, agentID string, threads []AIAgentTaskThreadRecord) (aiAgentAssignmentActionTarget, bool) {
	targets := activeAssignmentActionTargets(taskID, agentID, threads)
	if len(targets) == 0 {
		return aiAgentAssignmentActionTarget{}, false
	}
	return targets[0], true
}

func activeAssignmentActionTargetsForAgent(taskID, agentID string, threads []AIAgentTaskThreadRecord) []aiAgentAssignmentActionTarget {
	return activeAssignmentActionTargets(taskID, agentID, threads)
}

func activeAssignmentActionTargets(taskID, agentID string, threads []AIAgentTaskThreadRecord) []aiAgentAssignmentActionTarget {
	seen := map[string]bool{}
	targets := []aiAgentAssignmentActionTarget{}
	for _, state := range []AgentAssignmentState{AgentAssignmentStateRunning, AgentAssignmentStateQueued} {
		targets = appendActiveAssignmentActionTargets(targets, seen, taskID, agentID, threads, state)
	}
	return targets
}

func appendActiveAssignmentActionTargets(targets []aiAgentAssignmentActionTarget, seen map[string]bool, taskID, agentID string, threads []AIAgentTaskThreadRecord, state AgentAssignmentState) []aiAgentAssignmentActionTarget {
	for i := len(threads) - 1; i >= 0; i-- {
		target, ok := activeAssignmentActionTargetFromThread(taskID, agentID, threads[i], state)
		if !ok || seen[target.AssignmentID] {
			continue
		}
		seen[target.AssignmentID] = true
		targets = append(targets, target)
	}
	return targets
}

func activeAssignmentActionTargetFromThread(taskID, agentID string, thread AIAgentTaskThreadRecord, state AgentAssignmentState) (aiAgentAssignmentActionTarget, bool) {
	if strings.TrimSpace(agentID) != "" && strings.TrimSpace(thread.AgentID) != agentID {
		return aiAgentAssignmentActionTarget{}, false
	}
	if strings.TrimSpace(thread.AssignmentID) == "" || thread.AssignmentState != state {
		return aiAgentAssignmentActionTarget{}, false
	}
	return aiAgentAssignmentActionTarget{TaskID: taskID, AgentID: thread.AgentID, AssignmentID: thread.AssignmentID}, true
}
