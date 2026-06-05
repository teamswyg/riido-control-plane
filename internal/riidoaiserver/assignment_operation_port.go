package riidoaiserver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const AssignmentOperationSchemaVersion = "riido-ai-server-assignment-operation.v1"
const AssignmentProjectionSchemaVersion = "riido-ai-server-assignment-projection.v1"
const AssignmentAgentActiveSchemaVersion = "riido-ai-server-agent-active-assignment.v1"
const DefaultAssignmentActiveLeaseSeconds = 20

type AssignmentOperationType string

const (
	AssignmentOperationAssignTask AssignmentOperationType = "assign_task"
	AssignmentOperationPollStart  AssignmentOperationType = "poll_start"
	AssignmentOperationAgentEvent AssignmentOperationType = "agent_event"
	AssignmentOperationClientStop AssignmentOperationType = "client_stop"
)

type AssignmentOperationStore interface {
	SaveAssignmentOperation(ctx context.Context, record AssignmentOperationRecord) error
	Close() error
}

type AssignmentOperationLoader interface {
	LoadAssignmentOperations(ctx context.Context) ([]AssignmentOperationRecord, error)
}

type AssignmentQueueReader interface {
	LoadAgentQueueAssignments(ctx context.Context, agentID string) ([]Assignment, error)
}

type AssignmentClaimer interface {
	ClaimNextAssignment(ctx context.Context, agentID string, at time.Time) (AssignmentClaimResult, error)
}

type AssignmentActiveLeaseStore interface {
	LoadAgentActiveAssignment(ctx context.Context, agentID string) (AssignmentActiveLease, bool, error)
	RefreshAgentActiveAssignment(ctx context.Context, assignment Assignment, at time.Time) error
}

type AssignmentProjectionReader interface {
	LoadAssignmentProjection(ctx context.Context, assignmentID string) (AssignmentProjection, bool, error)
}

type AssignmentProjection struct {
	Assignment   Assignment
	LastEventSeq int64
}

type AssignmentActiveLease struct {
	AgentID            string
	ActiveAssignmentID string
	LeaseToken         string
	HeartbeatAt        time.Time
	LeaseExpiresAt     time.Time
	LeaseExpiresUnixMS int64
}

func (lease AssignmentActiveLease) Expired(at time.Time) bool {
	if lease.LeaseExpiresUnixMS > 0 {
		return lease.LeaseExpiresUnixMS <= at.UTC().UnixMilli()
	}
	if !lease.LeaseExpiresAt.IsZero() {
		return !lease.LeaseExpiresAt.After(at)
	}
	return true
}

type AssignmentClaimResult struct {
	Claimed    bool
	Assignment Assignment
	Operation  AssignmentOperationRecord
}

type AssignmentOperationRecord struct {
	SchemaVersion string                  `json:"schema_version"`
	OperationID   string                  `json:"operation_id"`
	OperationType AssignmentOperationType `json:"operation_type"`
	TaskID        string                  `json:"task_id"`
	AssignmentID  string                  `json:"assignment_id"`
	AgentID       string                  `json:"agent_id"`
	Assignment    Assignment              `json:"assignment"`
	Events        []TaskEvent             `json:"events"`
	RecordedAt    time.Time               `json:"recorded_at"`
}

func validateAssignmentOperationRecord(record AssignmentOperationRecord) error {
	if record.SchemaVersion != AssignmentOperationSchemaVersion {
		return fmt.Errorf("unsupported assignment operation schema_version %q", record.SchemaVersion)
	}
	if strings.TrimSpace(record.OperationID) == "" {
		return fmt.Errorf("assignment operation_id is required")
	}
	if record.OperationType == "" {
		return fmt.Errorf("assignment operation_type is required")
	}
	if strings.TrimSpace(record.TaskID) == "" {
		return fmt.Errorf("assignment operation task_id is required")
	}
	if strings.TrimSpace(record.AssignmentID) == "" {
		return fmt.Errorf("assignment operation assignment_id is required")
	}
	if strings.TrimSpace(record.AgentID) == "" {
		return fmt.Errorf("assignment operation agent_id is required")
	}
	if record.Assignment.ID != record.AssignmentID {
		return fmt.Errorf("assignment operation assignment_id mismatch %q", record.Assignment.ID)
	}
	if record.Assignment.TaskID != record.TaskID {
		return fmt.Errorf("assignment operation task_id mismatch %q", record.Assignment.TaskID)
	}
	if record.Assignment.AgentID != record.AgentID {
		return fmt.Errorf("assignment operation agent_id mismatch %q", record.Assignment.AgentID)
	}
	if len(record.Events) == 0 {
		return fmt.Errorf("assignment operation events are required")
	}
	for _, event := range record.Events {
		if event.Seq <= 0 {
			return fmt.Errorf("assignment operation event seq must be positive")
		}
		if event.TaskID != record.TaskID {
			return fmt.Errorf("assignment operation event task_id mismatch %q", event.TaskID)
		}
	}
	if record.RecordedAt.IsZero() {
		return fmt.Errorf("assignment operation recorded_at is required")
	}
	return nil
}

func assignmentOperationID(operationType AssignmentOperationType, assignment Assignment, events []TaskEvent) string {
	lastSeq := int64(0)
	if len(events) > 0 {
		lastSeq = events[len(events)-1].Seq
	}
	switch operationType {
	case AssignmentOperationAssignTask:
		return fmt.Sprintf("assign:%s:%d", assignment.ID, lastSeq)
	case AssignmentOperationPollStart:
		return fmt.Sprintf("poll-start:%s:%s:%d", assignment.ID, assignment.LeaseToken, lastSeq)
	case AssignmentOperationAgentEvent:
		return fmt.Sprintf("agent-event:%s:%d", assignment.ID, lastSeq)
	case AssignmentOperationClientStop:
		return fmt.Sprintf("client-stop:%s:%d", assignment.ID, lastSeq)
	default:
		return fmt.Sprintf("%s:%s:%d", operationType, assignment.ID, lastSeq)
	}
}

func assignmentOperationLastEventSeq(record AssignmentOperationRecord) int64 {
	if len(record.Events) == 0 {
		return 0
	}
	return record.Events[len(record.Events)-1].Seq
}

func assignmentSequence(id string) int64 {
	id = strings.TrimSpace(id)
	if !strings.HasPrefix(id, "asn-") {
		return 0
	}
	value := strings.TrimPrefix(id, "asn-")
	seq, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return seq
}

func assignmentQueueSort(assignment Assignment) string {
	createdAt := assignment.CreatedAt
	if createdAt.IsZero() {
		createdAt = assignment.UpdatedAt
	}
	return createdAt.UTC().Format("20060102T150405.000000000Z") + "#" + assignment.ID
}
