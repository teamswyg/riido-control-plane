package riidoaiserver

import (
	"net/http"
	"time"
)

func (s Server) serveExplicitAssignmentRead(w http.ResponseWriter, r *http.Request, a Assignment, device assignmentContextDevice, req AgentAssignmentContextRequest) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.config.AssignmentReadScope || len(s.config.CommentDelivery.Key) < 32 || s.config.CommentDelivery.EvidenceBaseURL == "" {
		writeError(w, 503, "explicit source read is not configured")
		return
	}
	store := s.config.AssignmentContext
	snapshot, found, err := store.LoadAssignmentContextSnapshot(r.Context(), a.ID)
	if err != nil {
		writeError(w, 503, "assignment context store unavailable")
		return
	}
	if a.CreatedBy == "" || a.WorkspaceID == "" || a.TaskID == "" || (found && (snapshot.RequesterPrincipalID != a.CreatedBy || snapshot.WorkspaceID != a.WorkspaceID || snapshot.TaskID != a.TaskID || snapshot.AgentID != a.AgentID)) {
		writeError(w, 403, "original assignment requester unavailable")
		return
	}
	if !found {
		snapshot = AssignmentContextSnapshot{AssignmentID: a.ID, WorkspaceID: a.WorkspaceID, TaskID: a.TaskID, AgentID: a.AgentID, RequesterPrincipalID: a.CreatedBy}
	}
	result, status, err := s.readExplicitAssignment(r.Context(), a, req)
	if err != nil {
		message := "explicit source read unavailable"
		readStatus := AssignmentContextReadUnavailable
		if status == 403 {
			message = "current requester cannot read this resource"
			readStatus = AssignmentContextReadDenied
		}
		if status == 400 {
			message = "invalid explicit source read input"
		}
		if status == 412 {
			message = "source changed; read the first page again"
		}
		s.writeAssignmentContextRefusal(w, r, snapshot, req, time.Now().UTC(), readStatus, status, message)
		return
	}
	data := any(map[string]any{"resource": result.Data, "target_id": result.ID, "team_id": result.TeamID, "title": result.Title, "live": true})
	if req.View == "thread" || req.View == "threads" {
		data, err = s.explicitThreadHistory(r.Context(), a, result, req)
		if err != nil {
			s.writeAssignmentContextRefusal(w, r, snapshot, req, time.Now().UTC(), AssignmentContextReadDenied, 403, "requested retained conversation unavailable")
			return
		}
	}
	// Network and store reads can outlive the lease. Check the device credential,
	// binding, projection and lease again before recording or releasing any data.
	currentDevice, ok := s.authorizeAssignmentContextDevice(w, r, a.AgentID)
	if !ok {
		return
	}
	if currentDevice.DeviceID != device.DeviceID {
		writeError(w, 403, "forbidden")
		return
	}
	current, ok := s.leasedAssignmentForContext(w, r, a.AgentID, currentDevice, req)
	if !ok {
		return
	}
	if current.CreatedBy != a.CreatedBy || current.TaskID != a.TaskID || current.WorkspaceID != a.WorkspaceID {
		writeError(w, 403, "assignment provenance changed")
		return
	}
	s.writeExplicitAssignmentRead(w, r, a, currentDevice, req, snapshot, result, data)
}
