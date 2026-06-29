package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientProfileThumbnailUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if s.aiAgentProfileThumbnails == nil {
		writeError(w, http.StatusServiceUnavailable, "profile thumbnail upload service is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentProfileThumbnailUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgentProfileThumbnails.CreateAIAgentProfileThumbnailUpload(r.Context(), principal, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
