package riidoaiserver

import "context"

func (s Server) deliverCommentInvocation(ctx context.Context, principal AuthorizationResult, evidence verifiedCommentDelivery, request AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	response, found, lookupErr := s.findTaskRequestReceipt(ctx, principal, evidence.TaskID, request)
	if lookupErr != nil {
		return AIAgentTaskActionResponse{}, lookupErr
	}
	if found {
		return response, s.ensureCommentInvocationContext(ctx, request.commentInvocation, response.AssignmentID)
	}
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, "", evidence.TaskID, request)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	if err := s.prepareCommentConversation(ctx, principal, evidence, &request, &composed.Request, composed.AgentOwnerPrincipalID); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	if err := s.appendCommentInvocationPrompt(ctx, principal, evidence.TaskID, &composed.Request); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	if composed.IntentGateRequired {
		request.intentGateRequired = true
		response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
		if err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		return response, nil
	}
	if call := request.commentInvocation; call != nil && call.OriginAssignmentID != "" {
		child := Assignment{ID: identifiedAssignmentID(evidence.TaskID, composed.Request), TaskID: evidence.TaskID, AgentID: composed.Request.AgentID, WorkspaceID: composed.Request.WorkspaceID, CreatedBy: composed.Request.CreatedBy}
		if err := s.InheritAssignmentContext(ctx, call.OriginAssignmentID, child, call.OriginMessageID); err != nil {
			return AIAgentTaskActionResponse{}, err
		}
	}
	creationContext, creationReceipt := withTaskAgentCreationReceipt(ctx)
	assignment, err := s.assignment.AssignTaskAdditive(creationContext, evidence.TaskID, composed.Request)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	request.AssignmentID, request.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		request.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
	if err != nil {
		s.recoverTaskAgentPublication(ctx, assignment, creationReceipt)
		return AIAgentTaskActionResponse{}, err
	}
	if request.commentInvocation == nil || request.commentInvocation.OriginAssignmentID == "" {
		s.captureAssignmentContext(ctx, assignment, composed.TaskContext, assignmentContextTrigger{Kind: AssignmentContextTriggerComment, ID: evidence.CommentID, RequesterAdmin: aiAgentIsAdmin(principal)})
	}
	return response, nil
}
