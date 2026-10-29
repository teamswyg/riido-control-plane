package riidoaiserver

import (
	"errors"
	"net/http"
	"time"
)

var errAssignmentScopedInlineCursor = errors.New("scoped inline cursor requires context part")

func (s Server) writeAssignmentContextViewError(w http.ResponseWriter, r *http.Request, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest, now time.Time, err error) {
	if errors.Is(err, errAssignmentScopedInlineCursor) {
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, AssignmentContextReadDenied, http.StatusBadRequest, "use context part with the captured cursor")
		return
	}
	writeError(w, http.StatusServiceUnavailable, "assignment context store unavailable")
}
