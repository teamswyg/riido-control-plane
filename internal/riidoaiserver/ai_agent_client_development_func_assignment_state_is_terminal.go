package riidoaiserver

func assignmentStateIsTerminal(state AssignmentState) bool {
	return state.Code().IsTerminal()
}
