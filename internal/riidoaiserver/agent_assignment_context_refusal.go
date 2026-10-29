package riidoaiserver

import (
	"net/http"
	"strings"
	"time"
)

func (s Server) writeAssignmentContextRefusal(w http.ResponseWriter, r *http.Request, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest, at time.Time, status string, httpStatus int, message string) {
	source := AssignmentContextSource{Kind: "task", ID: snapshot.TaskID, Part: req.View}
	if target := strings.TrimSpace(req.Target); target != "" {
		source = AssignmentContextSource{Kind: "target", ID: target, Part: req.View}
	}
	if req.View == AssignmentContextViewContent {
		source.Part = "metadata" // No captured leaf was served; retain the public audit enum.
	}
	reads := assignmentContextReadRecords(snapshot, []AssignmentContextSource{source}, AssignmentContextViaCLI, status, req.RequestID, at)
	if s.config.AssignmentContext.AppendAssignmentContextReads(r.Context(), reads) != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context read record unavailable")
		return
	}
	s.publishAssignmentContextReads(r.Context(), reads)
	writeError(w, httpStatus, message)
}
