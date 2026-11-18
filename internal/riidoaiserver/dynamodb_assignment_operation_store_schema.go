package riidoaiserver

import (
	"fmt"
	"time"
)

const (
	dynamoDBAssignmentOperationPK  = "ASSIGNMENT_OPERATION"
	dynamoDBAssignmentProjectionSK = "STATE"
	dynamoDBAgentActiveSK          = "ACTIVE"
	dynamoDBAssignmentQueueIndex   = "agent_queue"
	dynamoDBAssignmentQueryLimit   = 50
)

type assignmentProjectionRecord AssignmentProjection

func dynamoDBAssignmentProjectionPK(assignmentID string) string {
	return "ASSIGNMENT#" + assignmentID
}

func dynamoDBAgentActivePK(agentID string) string {
	return "AGENT#" + agentID
}

func assignmentOperationSortKey(record AssignmentOperationRecord) string {
	return assignmentOperationKey(record.RecordedAt, assignmentOperationLastEventSeq(record), record.OperationID)
}

func assignmentOperationKey(recordedAt time.Time, lastEventSeq int64, operationID string) string {
	return recordedAt.UTC().Format("20060102T150405.000000000Z") +
		"#" + fmt.Sprintf("%020d", lastEventSeq) + "#" + operationID
}
