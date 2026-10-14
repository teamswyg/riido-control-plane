package riidoaiserver

import "time"

func sampleAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentLeased,
		LeaseToken:      "lease-1",
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          2,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Metadata:     map[string]string{"lease_token": assignment.LeaseToken},
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationPollStart, assignment, events),
		OperationType: AssignmentOperationPollStart,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}
