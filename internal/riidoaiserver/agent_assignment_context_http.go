package riidoaiserver

import (
	"net/http"
	"time"
)

// handleAgentAssignmentContext serves the lease holder's snapshot or live views;
// every answer, refusals included, is recorded as a read before it returns.
func (s Server) handleAgentAssignmentContext(w http.ResponseWriter, r *http.Request, agentID string) {
	store := s.config.AssignmentContext
	if store == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context is not configured")
		return
	}
	device, ok := s.authorizeAssignmentContextDevice(w, r, agentID)
	if !ok {
		return
	}
	var req AgentAssignmentContextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !normalizeAssignmentContextRequest(&req) {
		writeError(w, http.StatusBadRequest, "assignment_id, lease_token and a supported view are required")
		return
	}
	assignment, ok := s.leasedAssignmentForContext(w, r, agentID, device, req)
	if !ok || !s.consumeAssignmentContextRead(w, r, assignment.ID) {
		return
	}
	if assignmentContextLiveView(req.View) {
		s.serveLiveAssignmentContext(w, r, assignment, req)
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
	if s.refuseUnavailableMeeting(w, r, snapshot, req, now) {
		return
	}
	snapshot.Truncated = snapshot.Truncated || assignmentContextMeetingTruncated(snapshot.Context, req)
	data, sources, inScope, err := s.assignmentContextViewData(r.Context(), snapshot, req)
	if err != nil {
		s.writeAssignmentContextViewError(w, r, snapshot, req, now, err)
		return
	}
	if !inScope {
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadDenied, http.StatusForbidden, "target is outside this assignment's context")
		return
	}
	status := AssignmentContextReadServed
	if snapshot.Truncated {
		status = AssignmentContextReadTruncated
	}
	// Fail closed: an answer the server could not record is not served.
	reads := assignmentContextReadRecords(snapshot, sources, AssignmentContextViaCLI, status, req.RequestID, now)
	if store.AppendAssignmentContextReads(r.Context(), reads) != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context read record unavailable")
		return
	}
	s.publishAssignmentContextReads(r.Context(), reads)
	writeJSON(w, http.StatusOK, AgentAssignmentContextResponse{
		SchemaVersion: SchemaVersion, AssignmentID: snapshot.AssignmentID, TaskID: snapshot.TaskID, View: req.View,
		CapturedAt: snapshot.CapturedAt, Truncated: snapshot.Truncated, Data: data, Sources: sources,
	})
}
