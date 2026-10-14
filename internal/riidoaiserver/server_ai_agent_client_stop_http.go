package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientStopTask(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req StopAIAgentTaskRequest
	if !decodeOptionalStopRequest(w, r, &req) {
		return
	}
	requestedAssignmentID := strings.TrimSpace(req.AssignmentID)
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, req.AgentID, req.AssignmentID, req.Reason); err != nil {
		writeAIAgentClientError(w, err)
		return
	} else if ok {
		req.AgentID = target.AgentID
		applyStopActionTarget(&req.AssignmentID, requestedAssignmentID, target)
		req.durableState = target.State
	}
	response, err := s.aiAgent.StopAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientStopTaskAgentAssignment(w http.ResponseWriter, r *http.Request, taskID, agentID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req AgentAssignmentActionRequest
	if !decodeOptionalStopRequest(w, r, &req) {
		return
	}
	response, err := s.stopAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
