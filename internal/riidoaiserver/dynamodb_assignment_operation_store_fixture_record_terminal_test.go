package riidoaiserver

import "time"

func sampleTerminalAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentCompleted,
		LeaseToken:      "lease-1",
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          3,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentCompleted,
		State:        AssignmentCompleted,
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
