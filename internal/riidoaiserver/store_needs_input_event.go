package riidoaiserver

// A provider's needs-input result ends the current execution lease while the
// client projection remains waiting_for_user for the same conversation.
func normalizeNeedsInputAssignmentEvent(req AgentEventRequest) AgentEventRequest {
	if req.State == AssignmentRunning && assignmentMetadataNeedsInput(req.Metadata) {
		req.State = AssignmentCompleted
	}
	return req
}
