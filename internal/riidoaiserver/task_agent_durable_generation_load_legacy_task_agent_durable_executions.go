package riidoaiserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func (s *Store) loadLegacyTaskAgentDurableExecutions(ctx context.Context, taskID, agentID string) ([]AssignmentProjection, error) {
	ids := map[string]bool{}
	if reader, ok := s.operationStore.(AssignmentQueueReader); ok {
		queue, err := reader.LoadAgentQueueAssignments(ctx, agentID)
		if err != nil {
			return nil, err
		}
		for _, assignment := range queue {
			if assignmentSequence(assignment.ID) > 0 {
				ids[assignment.ID] = true
			}
		}
	}
	if reader, ok := s.operationStore.(AssignmentActiveLeaseStore); ok {
		lease, found, err := reader.LoadAgentActiveAssignment(ctx, agentID)
		if err != nil {
			return nil, err
		}
		if found && assignmentSequence(lease.ActiveAssignmentID) > 0 {
			ids[lease.ActiveAssignmentID] = true
		}
	}
	var active []AssignmentProjection
	for id := range ids {
		projection, found, err := s.LoadAssignmentProjection(ctx, id)
		if err != nil {
			return nil, err
		}
		if found && projection.Assignment.TaskID == taskID && projection.Assignment.AgentID == agentID && !isTerminal(projection.Assignment.State) {
			active = append(active, projection)
		}
	}
	return active, nil
}

func taskAgentAssignmentGeneration(assignment Assignment) string {
	return assignment.ID + "@" + assignment.CreatedAt.UTC().Format(time.RFC3339Nano)
}

func (s *Store) newAssignmentID(sequence int64) (string, error) {
	if !s.taskAgentGenerationSupported() {
		return fmt.Sprintf("asn-%06d", sequence), nil
	}
	var identity [16]byte
	if _, err := rand.Read(identity[:]); err != nil {
		return "", fmt.Errorf("generate assignment identity: %w", err)
	}
	return "asn-" + hex.EncodeToString(identity[:]), nil
}

// A shared Store supplies the local gate to every HTTP Server using it. Separate
// production Stores still rely on the durable generation condition below.
func (s *Store) sharedTaskAgentExecutionGate() taskAgentExecutionGate {
	return s.taskAgentExecutionGate
}
