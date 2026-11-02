package riidoaiserver

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
	if kind == "" || (kind == "started" && s.startedNotificationSentLocked(event.TaskID, event.AssignmentID)) {
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
	if kind == "started" {
		s.markStartedNotificationSentLocked(event.TaskID, event.AssignmentID)
	}
}
