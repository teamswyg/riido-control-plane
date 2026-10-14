package riidoaiserver

func assignmentStateCarriesResultMessage(state AssignmentState) bool {
	switch state.Code() {
	case AssignmentStateCodeCancelled, AssignmentStateCodeCompleted, AssignmentStateCodeFailed:
		return true
	default:
		return false
	}
}
