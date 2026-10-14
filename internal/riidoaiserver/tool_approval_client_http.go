package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientTaskToolApprovals(w http.ResponseWriter, r *http.Request, taskID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	if _, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID,
	}); !ok {
		return
	}
	approvals, err := store.ListTaskToolApprovals(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolApprovalListResponse{SchemaVersion: SchemaVersion, Approvals: approvals})
}

func (s Server) handleAIAgentClientTaskToolApprovalDecision(
	w http.ResponseWriter,
	r *http.Request,
	taskID string,
	approvalID string,
) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionUpdate, TaskID: taskID,
	})
	if !ok {
		return
	}
	var decision ToolApprovalDecision
	if err := decodeJSON(r, &decision); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	decision.ApprovalID = approvalID
	if strings.TrimSpace(decision.DecidedBy) == "" {
		decision.DecidedBy = principal.PrincipalID
	}
	result, saved, err := store.DecideToolApproval(r.Context(), taskID, decision)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolApprovalDecisionResponse{SchemaVersion: SchemaVersion, Result: result, Decision: saved})
}
