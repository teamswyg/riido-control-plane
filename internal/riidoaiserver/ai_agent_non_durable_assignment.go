package riidoaiserver

// isNonDurableAssignmentID reports whether an assignment id names a task-thread
// record that exists only in the control-plane store and was never queued for a
// daemon. Such ids must not be cancelled, leased, or treated as durable execution
// when membership is removed.
func isNonDurableAssignmentID(assignmentID string) bool {
	return isIntentGateAssignmentID(assignmentID) || isOnboardingSeedAssignmentID(assignmentID)
}
