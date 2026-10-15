package riidoaiserver

import "context"

func (s Server) findTaskRequestReceipt(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, bool, error) {
	if req.RequestID == "" {
		return AIAgentTaskActionResponse{}, false, nil
	}
	if reader, ok := s.aiAgent.(interface {
		FindTaskRequestReceipt(context.Context, AuthorizationResult, string, AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, bool, error)
	}); ok {
		response, found, err := reader.FindTaskRequestReceipt(ctx, principal, taskID, req)
		if err != nil || !found {
			return response, found, err
		}
		if projectionReader, ok := s.assignment.(AssignmentProjectionReader); ok {
			projection, exists, err := projectionReader.LoadAssignmentProjection(ctx, response.AssignmentID)
			if err != nil {
				return AIAgentTaskActionResponse{}, false, err
			}
			if exists && !sameRecoveryPolicy(projection.Assignment.RecoveryPolicy, req.RecoveryPolicy) {
				return AIAgentTaskActionResponse{}, false, errRecoveryPolicyChanged
			}
			if !exists && req.RecoveryPolicy != nil {
				return AIAgentTaskActionResponse{}, false, errRecoveryPolicyChanged
			}
		} else if req.RecoveryPolicy != nil {
			return AIAgentTaskActionResponse{}, false, errRecoveryPolicyChanged
		}
		return response, true, nil
	}
	return AIAgentTaskActionResponse{}, false, nil
}
