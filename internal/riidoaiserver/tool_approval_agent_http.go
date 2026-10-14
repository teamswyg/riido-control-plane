package riidoaiserver

import (
	"net/http"
	"time"
)

func (s Server) handleAgentToolApprovalCreate(w http.ResponseWriter, r *http.Request, agentID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID,
	}); !ok {
		return
	}
	var req ToolApprovalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	approval, err := store.CreateToolApproval(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ToolApprovalCreateResponse{SchemaVersion: SchemaVersion, Approval: approval})
}

func (s Server) handleAgentToolApprovalWait(w http.ResponseWriter, r *http.Request, agentID, approvalID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll, AgentID: agentID,
	}); !ok {
		return
	}
	var req ToolApprovalWaitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hold := time.Duration(req.WaitMs) * time.Millisecond
	if maxHold := s.config.LongPollMaxHold; maxHold > 0 && hold > maxHold {
		hold = maxHold
	}
	result, decision, err := store.WaitForToolApproval(
		r.Context(), agentID, req.AssignmentID, approvalID, hold, s.config.LongPollTick,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolApprovalWaitResponse{SchemaVersion: SchemaVersion, Result: result, Decision: decision})
}
