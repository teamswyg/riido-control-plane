package riidoaiserver

import "net/http"

func (s Server) handleAgentCatalog(w http.ResponseWriter, r *http.Request) {
	if s.agentCatalog == nil {
		writeError(w, http.StatusServiceUnavailable, "agent catalog store is not configured")
		return
	}
	agentID, hasAgentID, ok := splitOptionalResourcePath(r.URL.Path, "/v1/agent-catalog")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case !hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogList(w, r)
	case !hasAgentID && r.Method == http.MethodPost:
		s.handleAgentCatalogCreate(w, r)
	case hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogGet(w, r, agentID)
	case hasAgentID && r.Method == http.MethodPatch:
		s.handleAgentCatalogUpdate(w, r, agentID)
	case hasAgentID && r.Method == http.MethodDelete:
		s.handleAgentCatalogDelete(w, r, agentID)
	default:
		writeMethodNotAllowed(w)
	}
}
