package riidoaiserver

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

// Assignment events commit before client projections. If projecting/saving the
// notification fails, a daemon retry must reuse the committed event identity.
func repeatedNotificationEvent(state *storeState, assignment Assignment, req AgentEventRequest) (TaskEvent, bool) {
	waiting := req.EventType == assignmentcontract.EventAssignmentWaitingForUser && req.ConfirmationReason != "" && assignment.ConfirmationReason == req.ConfirmationReason
	terminal := isTerminal(assignment.State) && req.State == assignment.State
	if !waiting && !terminal {
		return TaskEvent{}, false
	}
	events := state.events[assignment.TaskID]
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.AssignmentID != assignment.ID {
			continue
		}
		if waiting && event.Type == assignmentcontract.EventAssignmentWaitingForUser && event.ConfirmationReason == req.ConfirmationReason {
			return event, true
		}
		if terminal && event.State == assignment.State && (event.Type == EventAssignmentCompleted || event.Type == EventAssignmentFailed || event.Type == EventAssignmentCancelled) {
			return event, true
		}
	}
	return TaskEvent{}, false
}
