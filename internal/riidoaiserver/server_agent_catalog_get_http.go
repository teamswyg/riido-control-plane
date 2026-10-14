package riidoaiserver

import "net/http"

func (s Server) handleAgentCatalogGet(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead, AgentID: agentID})
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
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionRead)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: record})
}
