package riidoaiserver

import "fmt"

const (
	dynamoDBAssignmentOperationPK  = "ASSIGNMENT_OPERATION"
	dynamoDBAssignmentProjectionSK = "STATE"
	dynamoDBAgentActiveSK          = "ACTIVE"
	dynamoDBAssignmentQueueIndex   = "agent_queue"
	dynamoDBAssignmentQueryLimit   = 50
)

type assignmentProjectionRecord struct {
	Assignment   Assignment
	LastEventSeq int64
}

func dynamoDBAssignmentProjectionPK(assignmentID string) string {
	return "ASSIGNMENT#" + assignmentID
}

func dynamoDBAgentActivePK(agentID string) string {
	return "AGENT#" + agentID
}

func assignmentOperationSortKey(record AssignmentOperationRecord) string {
	return record.RecordedAt.UTC().Format("20060102T150405.000000000Z") +
		"#" + fmt.Sprintf("%020d", assignmentOperationLastEventSeq(record)) +
		"#" + record.OperationID
}
