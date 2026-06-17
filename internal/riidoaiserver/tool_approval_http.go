package riidoaiserver

import (
	"net/http"
	"strings"
	"time"
)

func (s Server) toolApprovalStore(w http.ResponseWriter) (AssignmentToolApprovalStore, bool) {
	store, ok := s.assignment.(AssignmentToolApprovalStore)
	if s.assignment == nil || !ok {
		writeError(w, http.StatusServiceUnavailable, "tool approval store is not configured")
		return nil, false
	}
	return store, true
}

func (s Server) handleAgentToolApprovalCreate(w http.ResponseWriter, r *http.Request, agentID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID}); !ok {
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
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll, AgentID: agentID}); !ok {
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
	result, decision, err := store.WaitForToolApproval(r.Context(), agentID, req.AssignmentID, approvalID, hold, s.config.LongPollTick)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolApprovalWaitResponse{SchemaVersion: SchemaVersion, Result: result, Decision: decision})
}

func (s Server) handleAIAgentClientTaskToolApprovals(w http.ResponseWriter, r *http.Request, taskID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	if _, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID}); !ok {
		return
	}
	approvals, err := store.ListTaskToolApprovals(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolApprovalListResponse{SchemaVersion: SchemaVersion, Approvals: approvals})
}

func (s Server) handleAIAgentClientTaskToolApprovalDecision(w http.ResponseWriter, r *http.Request, taskID, approvalID string) {
	store, ok := s.toolApprovalStore(w)
	if !ok {
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionUpdate, TaskID: taskID})
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

func toolApprovalDecisionSuffixApprovalID(suffix string) (string, bool) {
	return toolApprovalNestedSuffixID(suffix, "tool-approvals/", "/decision")
}

func toolApprovalWaitSuffixApprovalID(suffix string) (string, bool) {
	return toolApprovalNestedSuffixID(suffix, "tool-approvals/", "/wait")
}

func toolApprovalNestedSuffixID(suffix, prefix, tail string) (string, bool) {
	if !strings.HasPrefix(suffix, prefix) || !strings.HasSuffix(suffix, tail) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(suffix, prefix), tail)
	id = strings.Trim(id, "/")
	return id, id != ""
}
