package riidoaiserver

func assignmentStateCanRepairTaskThread(state AssignmentState) bool {
	return state.Code().IsKnown()
}
