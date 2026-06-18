package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"time"
)

func dynamoDBFailStaleQueuedOperation(candidate assignmentProjectionRecord, at time.Time) AssignmentOperationRecord {
	assignment := candidate.Assignment
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = at
	event := staleQueuedAssignmentEvent(assignment, candidate.LastEventSeq+1, at)
	events := []TaskEvent{event}
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

func (s *DynamoDBAssignmentOperationStore) claimRepairOnlyTransactionPayload(repair dynamoDBAssignmentClaimRepair) ([]byte, error) {
	projectionItem, err := assignmentProjectionDynamoDBItem(repair.Operation)
	if err != nil {
		return nil, err
	}
	operationItem, err := assignmentOperationDynamoDBItem(repair.Operation)
	if err != nil {
		return nil, err
	}
	payload := struct {
		TransactItems []dynamoDBTransactWriteItem `json:"TransactItems"`
	}{
		TransactItems: []dynamoDBTransactWriteItem{
			{Put: &dynamoDBTransactWritePutAction{
				TableName:           s.tableName,
				ConditionExpression: "assignment_state = :expected_state AND last_event_seq = :expected_last_event_seq",
				ExpressionAttributeValues: map[string]map[string]string{
					":expected_state":          {"S": string(repair.ExpectedState)},
					":expected_last_event_seq": {"N": strconv.FormatInt(repair.ExpectedLastEventSeq, 10)},
				},
				Item: projectionItem,
			}},
			{Put: &dynamoDBTransactWritePutAction{
				TableName:           s.tableName,
				ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
				Item:                operationItem,
			}},
		},
	}
	return json.Marshal(payload)
}
