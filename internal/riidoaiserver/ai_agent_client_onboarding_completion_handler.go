package riidoaiserver

import "net/http"

func (s Server) handleAIAgentClientCompleteOnboardingTask(w http.ResponseWriter, r *http.Request, taskID string) {
	workspaceID := aiAgentWorkspaceIDFromRequest(r)
	if workspaceID == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionAssign, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, ok := requestBearerToken(r)
	if !ok {
		writeUnauthorized(w)
		return
	}
	var req CompleteAIAgentOnboardingTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.verifyOnboardingFirstTaskBundleTask(r.Context(), bearerToken, workspaceID, principal.PrincipalID, taskID, req.TemplateKey); err != nil {
		writeOnboardingTaskProvenanceError(w, err)
		return
	}
	response, err := s.completeAIAgentOnboardingTaskUnderGate(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
