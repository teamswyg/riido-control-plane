package riidoaiserver

import (
	"time"
)

func failedPublicationCancellation(projection assignmentProjectionRecord, minimumEventSeq int64, at time.Time) *AssignmentOperationRecord {
	assignment := projection.Assignment
	if isTerminal(assignment.State) || assignment.State == AssignmentCancelling {
		return nil
	}
	state, eventType := AssignmentCancelling, EventAssignmentCancelling
	if assignment.State == AssignmentQueued {
		state, eventType = AssignmentCancelled, EventAssignmentCancelled
	}
	assignment.State, assignment.UpdatedAt = state, at
	events := []TaskEvent{{Seq: max(projection.LastEventSeq+1, minimumEventSeq), TaskID: assignment.TaskID, AgentID: assignment.AgentID, AssignmentID: assignment.ID, Type: eventType, State: state, Message: "client publication failed", Metadata: map[string]string{"source": "client_publication_failure"}, At: at}}
	return &AssignmentOperationRecord{SchemaVersion: AssignmentOperationSchemaVersion, OperationID: assignmentOperationID(AssignmentOperationClientStop, assignment, events), OperationType: AssignmentOperationClientStop, TaskID: assignment.TaskID, AgentID: assignment.AgentID, AssignmentID: assignment.ID, Assignment: assignment, Events: events, RecordedAt: at}
}
