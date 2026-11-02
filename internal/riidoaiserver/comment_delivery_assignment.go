package riidoaiserver

import "context"

func (s Server) deliverCommentInvocationStages(ctx context.Context, principal AuthorizationResult, evidence verifiedCommentDelivery, request AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	defer s.scheduleCommentStatusReactions(ctx, principal, evidence.TaskID)
	response, found, lookupErr := s.findTaskRequestReceipt(ctx, principal, evidence.TaskID, request)
	if lookupErr != nil {
		return AIAgentTaskActionResponse{}, withFailureStage("receipt_lookup", lookupErr)
	}
	if found {
		return response, withFailureStage("receipt_context", s.ensureCommentInvocationContext(ctx, request.commentInvocation, response.AssignmentID))
	}
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, "", evidence.TaskID, request)
	if err != nil {
		return AIAgentTaskActionResponse{}, withFailureStage("compose", err)
	}
	if err := s.prepareCommentConversation(ctx, principal, evidence, &request, &composed.Request, composed.AgentOwnerPrincipalID); err != nil {
		return AIAgentTaskActionResponse{}, withFailureStage("conversation", err)
	}
	if err := s.appendCommentInvocationPrompt(ctx, principal, evidence.TaskID, &composed.Request); err != nil {
		return AIAgentTaskActionResponse{}, withFailureStage("prompt", err)
	}
	if call := request.commentInvocation; call != nil && call.OriginAssignmentID != "" && !call.HumanFollowup {
		composed.Request.Prompt += "\n\nYou are the recipient of this handoff. Carry out the request addressed to you in " + assignmentInstructionHeading +
			". Earlier instructions to select a recipient were for the caller. Consult the directory only if your assigned work requires another Agent."
	}
	if composed.IntentGateRequired {
		request.intentGateRequired = true
		response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
		if err != nil {
			return AIAgentTaskActionResponse{}, withFailureStage("intent_gate_assignment", err)
		}
		return response, nil
	}
	if call := request.commentInvocation; call != nil && call.OriginAssignmentID != "" {
		child := Assignment{ID: identifiedAssignmentID(evidence.TaskID, composed.Request), TaskID: evidence.TaskID, AgentID: composed.Request.AgentID, WorkspaceID: composed.Request.WorkspaceID, CreatedBy: composed.Request.CreatedBy}
		if err := s.InheritAssignmentContext(ctx, call.OriginAssignmentID, child, call.OriginMessageID); err != nil {
			return AIAgentTaskActionResponse{}, withFailureStage("inherit_context", err)
		}
	}
	creationContext, creationReceipt := withTaskAgentCreationReceipt(ctx)
	assignment, err := s.assignment.AssignTaskAdditive(creationContext, evidence.TaskID, composed.Request)
	if err != nil {
		return AIAgentTaskActionResponse{}, withFailureStage("assign", err)
	}
	request.AssignmentID, request.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		request.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
	if err != nil {
		s.recoverTaskAgentPublication(ctx, assignment, creationReceipt)
		return AIAgentTaskActionResponse{}, withFailureStage("client_assignment", err)
	}
	if request.commentInvocation == nil || request.commentInvocation.OriginAssignmentID == "" {
		s.captureAssignmentContext(ctx, assignment, composed.TaskContext, assignmentContextTrigger{Kind: AssignmentContextTriggerComment, ID: evidence.CommentID, RequesterAdmin: aiAgentIsAdmin(principal)})
	}
	return response, nil
}
