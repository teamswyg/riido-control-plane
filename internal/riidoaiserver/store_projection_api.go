package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *Store) LoadAssignmentProjection(ctx context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("assignment_id is required")
	}
	if reader, ok := s.operationStore.(AssignmentProjectionReader); ok {
		return reader.LoadAssignmentProjection(ctx, assignmentID)
	}
	reply := make(chan assignmentProjectionResult, 1)
	if err := s.send(ctx, assignmentProjectionCmd{assignmentID: assignmentID, reply: reply}); err != nil {
		return AssignmentProjection{}, false, err
	}
	select {
	case res := <-reply:
		return res.projection, res.found, res.err
	case <-ctx.Done():
		return AssignmentProjection{}, false, ctx.Err()
	}
}
