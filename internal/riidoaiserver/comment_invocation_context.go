package riidoaiserver

import (
	"context"
	"errors"
)

// Repair an existing receipt from an interrupted/older handoff. New handoffs
// checkpoint inherited context before making the child execution pollable.
func (s Server) ensureCommentInvocationContext(ctx context.Context, call *CommentInvocation, childID string) error {
	if call == nil || call.OriginAssignmentID == "" || s.config.AssignmentContext == nil {
		return nil
	}
	if call.OriginThreadID == "" || call.OriginMessageID == "" {
		return errors.New("calling message identity unavailable")
	}
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return errors.New("child assignment context unavailable")
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, childID)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("child assignment unavailable")
	}
	return s.InheritAssignmentContext(ctx, call.OriginAssignmentID, projection.Assignment, call.OriginMessageID)
}
