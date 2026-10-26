package riidoaiserver

import (
	"net/http"
	"strings"
	"time"
)

// handleAgentAssignmentContext serves the assignment's context snapshot to
// the daemon that holds its lease. Every answer, including refusals inside an
// authorized lease, is written as a read record before it is returned.
func (s Server) handleAgentAssignmentContext(w http.ResponseWriter, r *http.Request, agentID string) {
	store := s.config.AssignmentContext
	if store == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context is not configured")
		return
	}
	var req AgentAssignmentContextRequest
	device, ok := s.authorizeAssignmentContextDevice(w, r, agentID)
	if !ok {
		return
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.View = strings.TrimSpace(req.View)
	if req.View == "" {
		req.View = AssignmentContextViewBase
	}
	if strings.TrimSpace(req.AssignmentID) == "" || strings.TrimSpace(req.LeaseToken) == "" || !validAssignmentContextView(req.View) {
		writeError(w, http.StatusBadRequest, "assignment_id, lease_token and a supported view are required")
		return
	}
	assignment, ok := s.leasedAssignmentForContext(w, r, agentID, device, req)
	if !ok {
		return
	}
	now := time.Now().UTC()
	snapshot, found, err := store.LoadAssignmentContextSnapshot(r.Context(), assignment.ID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context store unavailable")
		return
	}
	if !found {
		snapshot = AssignmentContextSnapshot{AssignmentID: assignment.ID, TaskID: assignment.TaskID, AgentID: assignment.AgentID, RequesterPrincipalID: assignment.CreatedBy}
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadUnavailable, http.StatusNotFound, "assignment context snapshot unavailable")
		return
	}
	data, sources, inScope := selectAssignmentContextView(snapshot.Context, req.View, req.Target)
	if !inScope {
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadDenied, http.StatusForbidden, "target is outside this assignment's context")
		return
	}
	status := AssignmentContextReadServed
	if snapshot.Truncated {
		status = AssignmentContextReadTruncated
	}
	// Fail closed: an answer the server could not record is not served.
	if store.AppendAssignmentContextReads(r.Context(), assignmentContextReadRecords(snapshot, sources, AssignmentContextViaCLI, status, req.RequestID, now)) != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context read record unavailable")
		return
	}
	writeJSON(w, http.StatusOK, AgentAssignmentContextResponse{
		SchemaVersion: SchemaVersion, AssignmentID: snapshot.AssignmentID, TaskID: snapshot.TaskID, View: req.View,
		CapturedAt: snapshot.CapturedAt, Truncated: snapshot.Truncated, Data: data, Sources: sources,
	})
}
