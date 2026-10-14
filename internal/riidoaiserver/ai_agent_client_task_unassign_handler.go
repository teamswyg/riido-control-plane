package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientUnassignTask(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req UnassignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fence, err := s.prepareAIAgentMembershipRemoval(r.Context(), principal, taskID, req.AgentID, req.AssignmentID, req.Reason)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	release, err := s.taskAgentExecutionGate.beginRemoval(r.Context(), taskID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	defer release()
	req.membershipRemovalFence = fence
	if fence != nil {
		req.AgentID = fence.agentID
	}
	response, err := s.aiAgent.UnassignAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientDeleteTaskAgentAssignment(w http.ResponseWriter, r *http.Request, taskID, agentID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req AgentAssignmentActionRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	fence, err := s.prepareAIAgentMembershipRemoval(r.Context(), principal, taskID, agentID, req.AssignmentID, req.Reason)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	release, err := s.taskAgentExecutionGate.beginRemoval(r.Context(), taskID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	defer release()
	req.membershipRemovalFence = fence
	response, err := s.aiAgent.DeleteAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
