package riidoaiserver

import "time"

func sampleQueuedAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentQueued,
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          1,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentQueued,
		State:        AssignmentQueued,
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAssignTask, assignment, events),
		OperationType: AssignmentOperationAssignTask,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}
