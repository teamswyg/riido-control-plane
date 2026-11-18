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
	req, cache, ok := parseAssignmentContextRequest(w, r)
	if !ok {
		return
	}
	assignment, ok := s.leasedAssignmentForContext(w, r, agentID, device, req)
	if !ok || !s.consumeAssignmentContextRead(w, r, assignment.ID) {
		return
	}
	if explicitAssignmentView(req.View) || len(req.Options) > 0 {
		s.serveExplicitAssignmentRead(w, r, assignment, device, req)
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
		s.handleMissingAssignmentSnapshot(w, r, assignment, device, req, now)
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
		s.handleOutsideAssignmentSnapshot(w, r, assignment, device, req, snapshot, now)
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
	response := AgentAssignmentContextResponse{
		SchemaVersion: SchemaVersion, AssignmentID: snapshot.AssignmentID, TaskID: snapshot.TaskID, View: req.View,
		CapturedAt: snapshot.CapturedAt, Truncated: snapshot.Truncated, Data: data, Sources: sources,
	}
	writeAssignmentContextRepresentation(w, cache, agentID, req, response)
}
