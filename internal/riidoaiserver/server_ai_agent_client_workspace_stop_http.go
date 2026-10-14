package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientAgentAssignments(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/v1/client/ai-agent/")
	if !strings.HasPrefix(suffix, "agent-assignments/") || !strings.HasSuffix(suffix, "/stop") || r.Method != http.MethodPost {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	agentID, ok := agentAssignmentStopSuffixAgentID(suffix)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	s.handleAIAgentClientStopWorkspaceAgentAssignment(w, r, agentID)
}

func (s Server) handleAIAgentClientStopWorkspaceAgentAssignment(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, AgentID: agentID})
	if !ok {
		return
	}
	var req AgentAssignmentActionRequest
	if !decodeOptionalStopRequest(w, r, &req) {
		return
	}
	response, err := s.stopAIAgentWorkspaceAgentAssignments(r.Context(), principal, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
