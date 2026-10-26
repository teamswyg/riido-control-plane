package riidoaiserver

import (
	"errors"
	"net/http"
	"strconv"
)

func (s Server) handleAIAgentClientRuns(w http.ResponseWriter, r *http.Request) {
	runID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/runs/")
	if !ok || suffix != "reads" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if _, ok := requestBearerToken(r); !ok {
		writeUnauthorized(w)
		return
	}
	// Authorize and read with the same viewer identity, even if a legacy agent
	// token header is also present. Neither that token nor service keys can widen access.
	r = r.Clone(r.Context())
	r.Header.Del(aiAgentTokenHeader)
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		var err error
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
	}
	runs, ok := s.aiAgent.(aiAgentRunReader)
	if !ok || s.config.AssignmentContext == nil {
		writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		return
	}
	thread, err := runs.FindAIAgentTaskThreadByRunID(r.Context(), principal.WorkspaceID, runID)
	if err != nil {
		if errors.Is(err, ErrAIAgentNotFound) {
			writeError(w, http.StatusNotFound, "not found")
		} else {
			writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		}
		return
	}
	if _, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: thread.TaskID}); !ok {
		return
	}
	visible, ok := s.currentRunReadSources(w, r, principal.WorkspaceID, thread.TaskID)
	if !ok {
		return
	}
	reads, err := s.config.AssignmentContext.ListAssignmentContextReads(r.Context(), thread.AssignmentID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		return
	}
	response, err := buildAIAgentRunReads(thread, reads, visible, page)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		return
	}
	writeJSON(w, http.StatusOK, response)
}
