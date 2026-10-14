package riidoaiserver

import "strings"

func assignmentStateIsKnown(state AssignmentState) bool {
	return strings.TrimSpace(string(state)) != ""
}
