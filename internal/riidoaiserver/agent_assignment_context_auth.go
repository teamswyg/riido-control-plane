package riidoaiserver

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

type assignmentContextDevice struct {
	DeviceID string
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
	return assignmentContextDevice{DeviceID: strings.TrimSpace(r.Header.Get(deviceIDHeader))}, true
}

// leasedAssignmentForContext admits only the device bound to the agent, while
// the assignment is active and the caller presents its current lease token.
// The device's enrollment workspace is not compared: one device may run agents
// of other workspaces, and the poll path leases those to it the same way.
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
		subtle.ConstantTimeCompare([]byte(assignment.LeaseToken), []byte(strings.TrimSpace(req.LeaseToken))) != 1 {
		writeError(w, http.StatusForbidden, "forbidden")
		return Assignment{}, false
	}
	switch assignment.State {
	case AssignmentLeased, AssignmentReady, AssignmentRunning:
	default:
		writeError(w, http.StatusConflict, "assignment is not active")
		return Assignment{}, false
	}
	if checker, ok := s.assignment.(assignmentLeaseChecker); ok {
		current, err := checker.AssignmentLeaseCurrent(r.Context(), assignment)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "assignment store unavailable")
			return Assignment{}, false
		}
		if !current {
			writeError(w, http.StatusConflict, "assignment lease expired")
			return Assignment{}, false
		}
	}
	return assignment, true
}
