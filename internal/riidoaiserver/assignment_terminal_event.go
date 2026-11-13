package riidoaiserver

func assignmentTerminalEventMatches(current Assignment, event TaskEvent) bool {
	if event.AssignmentID != current.ID || event.TaskID != current.TaskID ||
		event.AgentID != current.AgentID || event.State != current.State {
		return false
	}
	switch current.State {
	case AssignmentCompleted:
		return event.Type == EventAssignmentCompleted
	case AssignmentFailed:
		return event.Type == EventAssignmentFailed
	case AssignmentCancelled:
		return event.Type == EventAssignmentCancelled
	default:
		return false
	}
}

func assignmentProjectionTerminalResult(projection AssignmentProjection) (string, map[string]string) {
	event := projection.TerminalEvent
	if event == nil || event.Seq != projection.LastEventSeq || !assignmentTerminalEventMatches(projection.Assignment, *event) {
		return "", nil
	}
	return event.Message, event.Metadata
}
