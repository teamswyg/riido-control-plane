package riidoaiserver

import "strconv"

func taskThreadMessageRunID(taskID, assignmentID string, sequence int) string {
	if assignmentID != "" {
		return "run-dev-message-" + taskID + "-" + assignmentID
	}
	return "run-dev-message-" + taskID + "-" + strconv.Itoa(sequence)
}
