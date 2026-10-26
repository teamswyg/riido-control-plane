package riidoaiserver

import (
	"errors"
	"net/http"
	"time"
)

// serveLiveAssignmentContext answers live views for a leased assignment. A
// snapshot is not required: runs started by another agent's mention may not
// have one, and live data never comes from it. Records still fail closed.
func (s Server) serveLiveAssignmentContext(w http.ResponseWriter, r *http.Request, assignment Assignment, req AgentAssignmentContextRequest) {
	store := s.config.AssignmentContext
	now := time.Now().UTC()
	snapshot, found, err := store.LoadAssignmentContextSnapshot(r.Context(), assignment.ID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context store unavailable")
		return
	}
	if !found {
		snapshot = AssignmentContextSnapshot{AssignmentID: assignment.ID, TaskID: assignment.TaskID, WorkspaceID: assignment.WorkspaceID, AgentID: assignment.AgentID, RequesterPrincipalID: assignment.CreatedBy}
	}
	data, sources, inScope, err := s.assignmentAgentDirectoryView(r.Context(), assignment, found && snapshot.RequesterAdmin, req, now)
	if errors.Is(err, errAgentDirectoryRequesterUnknown) {
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadUnavailable, http.StatusNotFound, "assignment requester unavailable")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "agent directory unavailable")
		return
	}
	if !inScope {
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadDenied, http.StatusForbidden, "target is outside this assignment's context")
		return
	}
	reads := assignmentContextReadRecords(snapshot, sources, AssignmentContextViaCLI, AssignmentContextReadServed, req.RequestID, now)
	if len(reads) > 0 && store.AppendAssignmentContextReads(r.Context(), reads) != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context read record unavailable")
		return
	}
	s.publishAssignmentContextReads(r.Context(), reads)
	writeJSON(w, http.StatusOK, AgentAssignmentContextResponse{
		SchemaVersion: SchemaVersion, AssignmentID: assignment.ID, TaskID: assignment.TaskID, View: req.View,
		CapturedAt: now, Live: true, Data: data, Sources: sources,
	})
}
