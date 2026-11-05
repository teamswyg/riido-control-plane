package riidoaiserver

import "net/http"

// Only user-actionable transitions notify. Runtime logs and recoverable provider
// errors keep their original diagnostic channel and never become notifications.
func (s *DevelopmentAIAgentClientStore) appendTaskNotificationLocked(event TaskEvent, response AIAgentTaskActionResponse) {
	if event.EventID == "" || event.CreatedBy == "" || event.WorkspaceID == "" {
		return
	}
	for _, thread := range s.taskThreads[event.TaskID] {
		if thread.AssignmentID == event.AssignmentID && thread.CommentInvocation != nil && thread.CommentInvocation.OriginMessageID != "" {
			return
		}
	}
	if event.State == AssignmentCancelled || event.State == AssignmentCancelling {
		return
	}
	kind := ""
	switch response.WorkStatus {
	case AgentWorkStatusWaitingForUser:
		kind = response.ConfirmationReason
		if kind == "" && assignmentMetadataNeedsInput(event.Metadata) {
			kind = "answer_required"
		}
	case AgentWorkStatusCompleted:
		kind = "result_ready"
	case AgentWorkStatusFailed:
		kind = "failed"
	case AgentWorkStatusQueued, AgentWorkStatusRunning:
		for _, thread := range s.taskThreads[event.TaskID] {
			sourceID := thread.SourceMessageID
			if thread.ReactionSourceMessageID != nil {
				sourceID = *thread.ReactionSourceMessageID
			}
			if thread.AssignmentID == event.AssignmentID && (sourceID != "" || thread.SourceCommentID != "") {
				kind = "started"
				break
			}
		}
	case AgentWorkStatusIdle, AgentWorkStatusOffline:
		return
	}
	if kind == "" {
		return
	}

	for _, previous := range s.events {
		notification, ok := previous.Payload.(AgentTaskNotificationEvent)
		if !ok {
			continue
		}
		if notification.EventID == event.EventID || ((isTerminal(event.State) || kind == "started") && notification.AssignmentID == event.AssignmentID && notification.NotificationType == kind) {
			return
		}
	}
	assignmentState := response.AssignmentState
	if assignmentState == AgentAssignmentState("waiting_for_user") {
		stateResponse := AIAgentTaskActionResponse{}
		applyAssignmentStateActionResponse(&stateResponse, event.State)
		assignmentState = stateResponse.AssignmentState
	}
	s.appendClientEventLocked(AgentClientEventTaskNotification, AgentTaskNotificationEvent{
		EventType: AgentClientEventTaskNotification, SchemaVersion: SchemaVersion,
		EventID: event.EventID, OccurredAt: event.At, WorkspaceID: event.WorkspaceID,
		TaskID: event.TaskID, AgentID: event.AgentID, AssignmentID: event.AssignmentID,
		TeamID: event.TeamID, RunID: response.RunID, CreatedBy: event.CreatedBy, RecipientPrincipalID: event.CreatedBy,
		NotificationType: kind, ConfirmationReason: response.ConfirmationReason,
		WorkStatus: response.WorkStatus, AssignmentState: assignmentState,
	})
}

// Recheck task access at delivery time. Workspace membership or access to the
// shared agent alone is insufficient; revoked task access must fail closed.
func (s Server) clientNotificationAuthorized(r *http.Request, principal AuthorizationResult, event ClientStreamEvent) bool {
	if event.EventType != AgentClientEventTaskNotification {
		return true
	}
	notification, ok := event.Payload.(AgentTaskNotificationEvent)
	if !ok || notification.RecipientPrincipalID != principal.PrincipalID || notification.WorkspaceID != principal.WorkspaceID || s.config.Authorizer == nil {
		return false
	}
	token, ok := requestToken(r)
	if !ok {
		return false
	}
	current, err := s.config.Authorizer.Authorize(r.Context(), token, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead,
		WorkspaceID: notification.WorkspaceID, TaskID: notification.TaskID,
	})
	return err == nil && current.PrincipalID == principal.PrincipalID && (current.WorkspaceID == "" || current.WorkspaceID == notification.WorkspaceID)
}
