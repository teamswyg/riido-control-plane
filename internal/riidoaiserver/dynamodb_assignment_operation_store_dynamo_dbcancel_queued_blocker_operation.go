package riidoaiserver

import (
	"time"
)

func dynamoDBCancelQueuedBlockerOperation(blocker assignmentProjectionRecord, blockedAssignmentID string, at time.Time) AssignmentOperationRecord {
	assignment := blocker.Assignment
	assignment.State = AssignmentCancelled
	assignment.UpdatedAt = at
	events := []TaskEvent{{
		Seq:          blocker.LastEventSeq + 1,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentCancelled,
		State:        AssignmentCancelled,
		Message:      "queued blocker was cancelled before queued assignment claim",
		Metadata:     map[string]string{"blocked_assignment_id": blockedAssignmentID},
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, events),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

func dynamoDBFailStaleBlockerOperation(blocker assignmentProjectionRecord, blockedAssignmentID string, at time.Time) AssignmentOperationRecord {
	assignment := blocker.Assignment
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = at
	metadata := map[string]string{"blocked_assignment_id": blockedAssignmentID}
	if assignment.LeaseToken != "" {
		metadata["lease_token"] = assignment.LeaseToken
	}
	events := []TaskEvent{{
		Seq:          blocker.LastEventSeq + 1,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "blocked queued assignment repaired after stale blocker lease expired",
		Metadata:     metadata,
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, events),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}
