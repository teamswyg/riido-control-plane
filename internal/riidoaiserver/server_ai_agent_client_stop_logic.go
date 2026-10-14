package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) stopAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	requestedAssignmentID := strings.TrimSpace(req.AssignmentID)
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(ctx, principal, taskID, agentID, req.AssignmentID, req.Reason); err != nil {
		return AIAgentTaskActionResponse{}, err
	} else if ok {
		applyStopActionTarget(&req.AssignmentID, requestedAssignmentID, target)
		req.durableState = target.State
	}
	return s.aiAgent.StopAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
}

func (s Server) stopAIAgentWorkspaceAgentAssignments(ctx context.Context, principal AuthorizationResult, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	resolver, ok := s.aiAgent.(AIAgentActiveTaskThreadsForAgentResolver)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	threads, err := resolver.ActiveAIAgentTaskThreadsForAgent(ctx, principal, agentID)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	if len(threads) == 0 {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	var response AIAgentTaskActionResponse
	for _, taskID := range uniqueTaskIDsFromThreads(threads) {
		next, err := s.stopAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
		if err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		if response.TaskID == "" {
			response = next
		}
	}
	return response, nil
}

func uniqueTaskIDsFromThreads(threads []AIAgentTaskThreadRecord) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, thread := range threads {
		taskID := strings.TrimSpace(thread.TaskID)
		if taskID == "" || seen[taskID] {
			continue
		}
		seen[taskID] = true
		out = append(out, taskID)
	}
	return out
}
