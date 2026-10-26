package riidoaiserver

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"
)

type assignmentContextDevice struct {
	DeviceID    string
	WorkspaceID string
}

// authorizeAssignmentContextDevice accepts only the daemon's device
// credential. A user Bearer is never enough to read an assignment snapshot.
func (s Server) authorizeAssignmentContextDevice(w http.ResponseWriter, r *http.Request, agentID string) (assignmentContextDevice, bool) {
	result, handled := s.authorizeDeviceCredential(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionRead, AgentID: agentID})
	if !handled {
		writeUnauthorized(w)
		return assignmentContextDevice{}, false
	}
	if strings.TrimSpace(result.PrincipalID) == "" {
		return assignmentContextDevice{}, false
	}
	return assignmentContextDevice{DeviceID: strings.TrimSpace(r.Header.Get(deviceIDHeader)), WorkspaceID: result.WorkspaceID}, true
}

// leasedAssignmentForContext admits only the device bound to the agent, while
// the assignment is active and the caller presents its current lease token.
func (s Server) leasedAssignmentForContext(w http.ResponseWriter, r *http.Request, agentID string, device assignmentContextDevice, req AgentAssignmentContextRequest) (Assignment, bool) {
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "assignment projection reader is not configured")
		return Assignment{}, false
	}
	projection, found, err := reader.LoadAssignmentProjection(r.Context(), strings.TrimSpace(req.AssignmentID))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store unavailable")
		return Assignment{}, false
	}
	assignment := projection.Assignment
	if !found || assignment.AgentID != agentID {
		writeError(w, http.StatusNotFound, "assignment not found")
		return Assignment{}, false
	}
	binding, _, err := s.resolveAgentRuntimeFact(agentID)
	if err != nil || binding.DeviceID == "" || binding.DeviceID != device.DeviceID ||
		(assignment.WorkspaceID != "" && device.WorkspaceID != assignment.WorkspaceID) ||
		subtle.ConstantTimeCompare([]byte(assignment.LeaseToken), []byte(strings.TrimSpace(req.LeaseToken))) != 1 {
		writeError(w, http.StatusForbidden, "forbidden")
		return Assignment{}, false
	}
	switch assignment.State {
	case AssignmentLeased, AssignmentReady, AssignmentRunning:
		return assignment, true
	default:
		writeError(w, http.StatusConflict, "assignment is not active")
		return Assignment{}, false
	}
}

func (s Server) writeAssignmentContextRefusal(w http.ResponseWriter, r *http.Request, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest, at time.Time, status string, httpStatus int, message string) {
	source := AssignmentContextSource{Kind: "task", ID: snapshot.TaskID, Part: req.View}
	if target := strings.TrimSpace(req.Target); target != "" {
		source = AssignmentContextSource{Kind: "target", ID: target, Part: req.View}
	}
	reads := assignmentContextReadRecords(snapshot, []AssignmentContextSource{source}, AssignmentContextViaCLI, status, req.RequestID, at)
	if s.config.AssignmentContext.AppendAssignmentContextReads(r.Context(), reads) != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context read record unavailable")
		return
	}
	s.publishAssignmentContextReads(r.Context(), reads)
	writeError(w, httpStatus, message)
}
