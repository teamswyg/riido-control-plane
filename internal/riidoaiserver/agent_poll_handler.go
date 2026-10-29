package riidoaiserver

import (
	"net/http"
	"time"
)

func (s Server) handleAgentPoll(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	auth, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll, AgentID: agentID})
	if !ok {
		return
	}
	var req PollRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, ok := authenticatedPollContext(w, r, req, auth)
	if !ok {
		return
	}
	var response PollResponse
	var err error
	if longPoll, ok := s.assignment.(AssignmentLongPollStore); ok && req.WaitMs > 0 {
		hold := time.Duration(req.WaitMs) * time.Millisecond
		if maxHold := s.config.LongPollMaxHold; maxHold > 0 && hold > maxHold {
			hold = maxHold
		}
		response, err = longPoll.WaitForAssignment(ctx, agentID, req, hold, s.config.LongPollTick)
	} else {
		response, err = s.assignment.PollAgent(ctx, agentID, req)
	}
	if err != nil {
		logAgentPollRejected(agentID, err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}
