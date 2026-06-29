package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleComponentTasks(w http.ResponseWriter, r *http.Request) {
	taskID, suffix, ok := splitResourcePath(r.URL.Path, "/v1/component-tasks/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "assignment" && r.Method == http.MethodPost:
		s.handleComponentTaskAssign(w, r, taskID)
	case suffix == "events" && r.Method == http.MethodGet:
		s.handleComponentTaskEvents(w, r, taskID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleComponentTaskAssign(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceComponentTask, Action: AuthorizationActionAssign, TaskID: taskID}); !ok {
		return
	}
	var req AssignRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Prompt) == "" && s.taskContext != nil {
		composedReq, err := s.assignRequestWithTaskContextPrompt(r.Context(), taskID, req)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		req = composedReq
	}
	assignment, err := s.assignment.AssignTask(r.Context(), taskID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		SchemaVersion string     `json:"schema_version"`
		Assignment    Assignment `json:"assignment"`
	}{SchemaVersion: SchemaVersion, Assignment: assignment})
}

func (s Server) handleComponentTaskEvents(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceComponentTaskEvents, Action: AuthorizationActionEventsRead, TaskID: taskID}); !ok {
		return
	}
	s.streamTaskEvents(w, r, taskID)
}
