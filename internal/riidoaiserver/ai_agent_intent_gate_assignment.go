package riidoaiserver

import "strings"

const intentGateAssignmentIDPrefix = "intent-gate-"

func intentGateAssignmentID(taskID, sequence string) string {
	return intentGateAssignmentIDPrefix + strings.TrimSpace(taskID) + "-" + strings.TrimSpace(sequence)
}

func isIntentGateAssignmentID(assignmentID string) bool {
	return strings.HasPrefix(strings.TrimSpace(assignmentID), intentGateAssignmentIDPrefix)
}
