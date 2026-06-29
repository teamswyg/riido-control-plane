package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientTasks(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	taskID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/tasks/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "assignable-agents" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskAssignableAgents(w, r, taskID)
	case suffix == "assignment" && r.Method == http.MethodPost:
		s.handleAIAgentClientAssignTask(w, r, taskID)
	case suffix == "assignment" && r.Method == http.MethodDelete:
		s.handleAIAgentClientUnassignTask(w, r, taskID)
	case suffix == "agent-assignments" && r.Method == http.MethodPost:
		s.handleAIAgentClientCreateTaskAgentAssignment(w, r, taskID)
	case strings.HasPrefix(suffix, "agent-assignments/") && strings.HasSuffix(suffix, "/stop") && r.Method == http.MethodPost:
		agentID, ok := agentAssignmentStopSuffixAgentID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientStopTaskAgentAssignment(w, r, taskID, agentID)
	case strings.HasPrefix(suffix, "agent-assignments/") && r.Method == http.MethodDelete:
		agentID, ok := agentAssignmentSuffixAgentID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientDeleteTaskAgentAssignment(w, r, taskID, agentID)
	case suffix == "threads" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreads(w, r, taskID)
	case suffix == "tool-approvals" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskToolApprovals(w, r, taskID)
	case strings.HasPrefix(suffix, "tool-approvals/") && strings.HasSuffix(suffix, "/decision") && r.Method == http.MethodPost:
		approvalID, ok := toolApprovalDecisionSuffixApprovalID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientTaskToolApprovalDecision(w, r, taskID, approvalID)
	case suffix == "thread-stream-subscription" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreadStreamSubscription(w, r, taskID)
	case suffix == "comments" && r.Method == http.MethodPost:
		s.handleAIAgentClientSubmitTaskComment(w, r, taskID)
	case strings.HasPrefix(suffix, "threads/") && strings.HasSuffix(suffix, "/messages") && r.Method == http.MethodPost:
		threadID, ok := threadMessageSuffixThreadID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientCreateTaskThreadMessage(w, r, taskID, threadID)
	case suffix == "stop" && r.Method == http.MethodPost:
		s.handleAIAgentClientStopTask(w, r, taskID)
	default:
		writeMethodNotAllowed(w)
	}
}
