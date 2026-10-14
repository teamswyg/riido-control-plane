package riidoaiserver

func applyCloudWatchEMFStoreState(envelope *cloudWatchEMFEnvelope, snapshot MetricsSnapshot) {
	envelope.TasksTotal = snapshot.TasksTotal
	envelope.AssignmentsTotal = snapshot.AssignmentsTotal
	envelope.AssignmentsQueued = snapshot.AssignmentsByState[AssignmentQueued]
	envelope.AssignmentsLeased = snapshot.AssignmentsByState[AssignmentLeased]
	envelope.AssignmentsReady = snapshot.AssignmentsByState[AssignmentReady]
	envelope.AssignmentsRunning = snapshot.AssignmentsByState[AssignmentRunning]
	envelope.AssignmentsCancelling = snapshot.AssignmentsByState[AssignmentCancelling]
	envelope.AssignmentsCancelled = snapshot.AssignmentsByState[AssignmentCancelled]
	envelope.AssignmentsCompleted = snapshot.AssignmentsByState[AssignmentCompleted]
	envelope.AssignmentsFailed = snapshot.AssignmentsByState[AssignmentFailed]
}
