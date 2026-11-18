package riidoaiserver

import (
	"net/http"
	"time"
)

func (s Server) canReadExplicitTarget(req AgentAssignmentContextRequest) bool {
	return (req.View == AssignmentContextViewTask || req.View == AssignmentContextViewComments) && s.config.AssignmentReadScope && s.config.CommentDelivery.Key != ""
}

func (s Server) handleMissingAssignmentSnapshot(w http.ResponseWriter, r *http.Request, assignment Assignment, device assignmentContextDevice, req AgentAssignmentContextRequest, now time.Time) {
	if s.canReadExplicitTarget(req) {
		s.serveExplicitAssignmentRead(w, r, assignment, device, req)
		return
	}
	snapshot := AssignmentContextSnapshot{AssignmentID: assignment.ID, TaskID: assignment.TaskID, AgentID: assignment.AgentID, RequesterPrincipalID: assignment.CreatedBy}
	s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadUnavailable, http.StatusNotFound, "assignment context snapshot unavailable")
}

func (s Server) handleOutsideAssignmentSnapshot(w http.ResponseWriter, r *http.Request, a Assignment, device assignmentContextDevice, req AgentAssignmentContextRequest, snapshot AssignmentContextSnapshot, now time.Time) {
	if s.canReadExplicitTarget(req) {
		s.serveExplicitAssignmentRead(w, r, a, device, req)
		return
	}
	s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadDenied, http.StatusForbidden, "target is outside this assignment's context")
}
