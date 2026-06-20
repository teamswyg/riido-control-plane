package riidoaiserver

import "net/http"

func (s Server) handleAgentCatalogUpdate(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionUpdate, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionUpdate)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req UpdateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record.Visibility = req.Visibility
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}
