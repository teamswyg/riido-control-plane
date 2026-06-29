package riidoaiserver

import (
	"net/http"
)

func (s Server) handleProviderStatusSync(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.provider == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusWrite, AgentID: agentID}); !ok {
		return
	}
	var req ProviderStatusSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.provider.SyncProviderStatus(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleProviderStatusRead(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.providerRead == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status reader is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusRead, AgentID: agentID}); !ok {
		return
	}
	response, found, err := s.providerRead.GetProviderStatus(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, response)
}
