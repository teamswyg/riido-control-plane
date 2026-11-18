package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s Server) writeExplicitAssignmentRead(w http.ResponseWriter, r *http.Request, a Assignment, currentDevice assignmentContextDevice, req AgentAssignmentContextRequest, snapshot AssignmentContextSnapshot, result explicitAssignmentResult, data any) {
	raw, err := json.Marshal(data)
	if err != nil || len(raw) > 16<<20 {
		writeError(w, 503, "explicit source response too large")
		return
	}
	kind := result.Kind
	if kind == "subtask" {
		kind = "task"
	}
	sources := []AssignmentContextSource{{Kind: kind, ID: result.ID, Part: "metadata", Version: assignmentContextVersion(string(raw))}}
	now := time.Now().UTC()
	readStatus := AssignmentContextReadServed
	if result.Truncated {
		readStatus = AssignmentContextReadTruncated
	}
	reads := assignmentContextReadRecords(snapshot, sources, AssignmentContextViaCLI, readStatus, req.RequestID, now)
	for i := range reads {
		reads[i].Explicit = true
	}
	if s.config.AssignmentContext.AppendAssignmentContextReads(r.Context(), reads) != nil {
		writeError(w, 503, "assignment context read record unavailable")
		return
	}
	if _, ok := s.leasedAssignmentForContext(w, r, a.AgentID, currentDevice, req); !ok {
		return
	}
	s.publishAssignmentContextReads(r.Context(), reads)
	if _, ok := s.leasedAssignmentForContext(w, r, a.AgentID, currentDevice, req); !ok {
		return
	}
	writeJSON(w, 200, AgentAssignmentContextResponse{SchemaVersion: SchemaVersion, AssignmentID: a.ID, TaskID: a.TaskID, View: req.View, CapturedAt: now, Live: true, Truncated: result.Truncated, Data: data, Sources: sources})
}
