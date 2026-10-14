package riidoaiserver

import (
	"errors"
	"strings"
)

func assignmentActionTargetByAssignmentID(taskID, agentID, assignmentID string, threads []AIAgentTaskThreadRecord) (aiAgentAssignmentActionTarget, bool, error) {
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if strings.TrimSpace(thread.AssignmentID) != assignmentID {
			continue
		}
		if agentID != "" && strings.TrimSpace(thread.AgentID) != agentID {
			return aiAgentAssignmentActionTarget{}, false, errors.New("assignment_id does not belong to task agent")
		}
		return aiAgentAssignmentActionTarget{TaskID: taskID, AgentID: thread.AgentID, AssignmentID: assignmentID}, true, nil
	}
	return aiAgentAssignmentActionTarget{}, false, errors.New("assignment_id does not belong to task agent")
}

func activeAssignmentActionTarget(taskID string, threads []AIAgentTaskThreadRecord) (aiAgentAssignmentActionTarget, bool, error) {
	target, ok := preferredActiveAssignmentActionTarget(taskID, "", threads)
	return target, ok, nil
}

func actionTargetFromThread(taskID, agentID string, threads []AIAgentTaskThreadRecord, activeOnly bool) (aiAgentAssignmentActionTarget, bool) {
	if activeOnly {
		return preferredActiveAssignmentActionTarget(taskID, agentID, threads)
	}
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if strings.TrimSpace(thread.AgentID) != agentID || strings.TrimSpace(thread.AssignmentID) == "" {
			continue
		}
		return aiAgentAssignmentActionTarget{
			TaskID:       taskID,
			AgentID:      agentID,
			AssignmentID: thread.AssignmentID,
		}, true
	}
	return aiAgentAssignmentActionTarget{}, false
}
