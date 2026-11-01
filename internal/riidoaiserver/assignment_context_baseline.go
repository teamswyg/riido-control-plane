package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

var errAssignmentContextBaseline = errors.New("assignment context baseline unavailable")

const assignmentContextBaselineMaxDepth = 32

func validAssignmentContextSince(req AgentAssignmentContextRequest) bool {
	if req.View != AssignmentContextViewChanges {
		return req.Since == ""
	}
	since := strings.TrimSpace(req.Since)
	return since != "" && len(since) <= 256 && !strings.ContainsAny(since, "\r\n\x00") && req.Target == "" && req.Cursor == "" && req.MaxBytes == 0
}

func sameAssignmentContextScope(left, right AssignmentContextSnapshot) bool {
	return left.TaskID != "" && left.WorkspaceID != "" && left.AgentID != "" && left.RequesterPrincipalID != "" &&
		left.Context.TeamID != "" && left.TaskID == right.TaskID && left.WorkspaceID == right.WorkspaceID &&
		left.AgentID == right.AgentID && left.RequesterPrincipalID == right.RequesterPrincipalID &&
		left.Context.TeamID == right.Context.TeamID && left.RequesterAdmin == right.RequesterAdmin
}

func (s Server) assignmentContextBaseline(ctx context.Context, current AssignmentContextSnapshot, since string) (AssignmentContextSnapshot, error) {
	cursor := current
	seen := map[string]bool{}
	for depth := 0; depth <= assignmentContextBaselineMaxDepth; depth++ {
		if err := ctx.Err(); err != nil {
			return AssignmentContextSnapshot{}, err
		}
		if cursor.CapturedAt.IsZero() || cursor.ConversationID == "" || cursor.ThreadID == "" || cursor.TriggerKind == AssignmentContextTriggerAgentMention ||
			seen[cursor.AssignmentID] || !sameAssignmentContextScope(current, cursor) {
			return AssignmentContextSnapshot{}, errAssignmentContextBaseline
		}
		seen[cursor.AssignmentID] = true
		if cursor.AssignmentID == since {
			return cursor, nil
		}
		if depth == assignmentContextBaselineMaxDepth || cursor.PreviousAssignmentID == "" || cursor.PreviousThreadID == "" {
			break
		}
		parent, found, err := s.config.AssignmentContext.LoadAssignmentContextSnapshot(ctx, cursor.PreviousAssignmentID)
		if err != nil {
			return AssignmentContextSnapshot{}, err
		}
		if !found || parent.AssignmentID != cursor.PreviousAssignmentID || parent.ThreadID != cursor.PreviousThreadID ||
			parent.ConversationID != cursor.ConversationID || parent.CapturedAt.IsZero() || cursor.CapturedAt.Before(parent.CapturedAt) {
			break
		}
		cursor = parent
	}
	return AssignmentContextSnapshot{}, errAssignmentContextBaseline
}
