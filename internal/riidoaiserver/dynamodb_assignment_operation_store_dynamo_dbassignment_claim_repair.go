package riidoaiserver

import (
	"fmt"
	"strings"
	"time"
)

type dynamoDBAssignmentClaimRepair struct {
	Operation            AssignmentOperationRecord
	ExpectedState        AssignmentState
	ExpectedLastEventSeq int64
}

func dynamoDBClaimAssignmentOperation(assignment Assignment, lastEventSeq int64, at time.Time, clearMessage string, clearMetadata map[string]string) AssignmentOperationRecord {
	claimed := assignment
	events := []TaskEvent{}
	nextSeq := lastEventSeq
	blockedByID := strings.TrimSpace(claimed.BlockedByAssignmentID)
	if strings.TrimSpace(clearMessage) != "" && blockedByID != "" {
		nextSeq++
		metadata := map[string]string{"blocked_by_assignment_id": blockedByID}
		for key, value := range clearMetadata {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key != "" && value != "" {
				metadata[key] = value
			}
		}
		events = append(events, TaskEvent{
			Seq:          nextSeq,
			TaskID:       claimed.TaskID,
			AssignmentID: claimed.ID,
			AgentID:      claimed.AgentID,
			Type:         EventAssignmentQueued,
			State:        AssignmentQueued,
			Message:      clearMessage,
			Metadata:     metadata,
			At:           at,
		})
		claimed.BlockedByAssignmentID = ""
		claimed.UpdatedAt = at
	}
	claimed.State = AssignmentLeased
	claimed.LeaseToken = fmt.Sprintf("%s:%d", claimed.ID, at.UnixNano())
	claimed.UpdatedAt = at
	nextSeq++
	events = append(events, TaskEvent{
		Seq:          nextSeq,
		TaskID:       claimed.TaskID,
		AssignmentID: claimed.ID,
		AgentID:      claimed.AgentID,
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Metadata:     map[string]string{"lease_token": claimed.LeaseToken},
		At:           at,
	})
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationPollStart, claimed, events),
		OperationType: AssignmentOperationPollStart,
		TaskID:        claimed.TaskID,
		AssignmentID:  claimed.ID,
		AgentID:       claimed.AgentID,
		Assignment:    claimed,
		Events:        events,
		RecordedAt:    at,
	}
}
