package riidoaiserver

import "strconv"

func nextTaskThreadSequence(threads []AIAgentTaskThreadRecord) string {
	return strconv.Itoa(len(threads) + 1)
}

func taskAssignmentIDForRequest(taskID, sequence string, req AssignAIAgentTaskRequest) string {
	if req.AssignmentID != "" {
		return req.AssignmentID
	}
	if req.intentGateRequired {
		return intentGateAssignmentID(taskID, sequence)
	}
	return "asn-dev-assignment-" + taskID + "-" + sequence
}

func applyAssignmentStartProjection(
	response *AIAgentTaskActionResponse,
	agentStatus AgentWorkStatus,
	req AssignAIAgentTaskRequest,
) {
	if req.intentGateRequired {
		response.WorkStatus = AgentWorkStatusWaitingForUser
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentRuntimeProgress
		response.Message = clientMessageNeedUserInput
		return
	}
	if agentStatus == AgentWorkStatusRunning ||
		agentStatus == AgentWorkStatusWaitingForUser ||
		agentStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = clientMessageAgentBusyQueued
	}
}
