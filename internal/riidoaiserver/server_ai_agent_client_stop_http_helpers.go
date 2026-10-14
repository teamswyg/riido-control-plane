package riidoaiserver

import "net/http"

func decodeOptionalStopRequest(w http.ResponseWriter, r *http.Request, req any) bool {
	if r.Body == nil || r.ContentLength == 0 {
		return true
	}
	if err := decodeJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func applyStopActionTarget(assignmentID *string, requestedAssignmentID string, target aiAgentAssignmentActionTarget) {
	if requestedAssignmentID != "" {
		*assignmentID = target.AssignmentID
	}
}
