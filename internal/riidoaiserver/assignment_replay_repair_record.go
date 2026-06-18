package riidoaiserver

import "time"

func repairReplayAssignmentAsFailed(
	state *storeState,
	assignment Assignment,
	eventFn func(Assignment, int64, time.Time) TaskEvent,
	now time.Time,
) AssignmentOperationRecord {
	state.nextEventSeq++
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	event := eventFn(assignment, state.nextEventSeq, now)
	state.events[assignment.TaskID] = append(state.events[assignment.TaskID], event)
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, []TaskEvent{event}),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        []TaskEvent{event},
		RecordedAt:    now,
	}
}
