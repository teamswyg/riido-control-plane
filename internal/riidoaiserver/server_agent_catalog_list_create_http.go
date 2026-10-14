package riidoaiserver

import "net/http"

func (s Server) handleAgentCatalogList(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	records, err := s.agentCatalog.ListAgentCatalog(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogListResponse{
		SchemaVersion: SchemaVersion,
		Agents:        VisibleAgentCatalogRecords(principal, records),
	})
}

func (s Server) handleAgentCatalogCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record := AgentCatalogRecord{
		AgentID:          req.AgentID,
		OwnerPrincipalID: principal.PrincipalID,
		Visibility:       req.Visibility,
	}
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}
