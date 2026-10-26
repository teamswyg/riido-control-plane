package riidoaiserver

import (
	"context"
	"log"
	"strings"
	"time"
)

// RequesterAdmin is the requester's admin role at request time; roles are
// only known while the user's request is authorized, like the snapshot.
type assignmentContextTrigger struct {
	Kind           string
	ID             string
	RequesterAdmin bool
}

// captureAssignmentContext stores the prompt's task context and records its
// sources as read via the prompt; failures are logged (CLI: unavailable).
func (s Server) captureAssignmentContext(ctx context.Context, assignment Assignment, taskContext *AIAgentTaskContext, trigger assignmentContextTrigger) {
	store := s.config.AssignmentContext
	if store == nil || taskContext == nil || strings.TrimSpace(assignment.ID) == "" {
		return
	}
	// A reused assignment keeps its first snapshot: what the run received.
	if _, exists, err := store.LoadAssignmentContextSnapshot(ctx, assignment.ID); err != nil || exists {
		return
	}
	now := time.Now().UTC()
	snapshot := AssignmentContextSnapshot{
		AssignmentID:         assignment.ID,
		TaskID:               assignment.TaskID,
		WorkspaceID:          assignment.WorkspaceID,
		AgentID:              assignment.AgentID,
		RequesterPrincipalID: assignment.CreatedBy,
		TriggerKind:          trigger.Kind,
		TriggerID:            trigger.ID,
		RequesterAdmin:       trigger.RequesterAdmin,
		CapturedAt:           now,
	}
	contextOnly, comments := splitAssignmentContextComments(*taskContext, trigger)
	snapshot.Context, snapshot.Truncated = boundAssignmentContext(contextOnly)
	snapshot.CommentPages = s.saveAssignmentContextComments(ctx, &snapshot, comments)
	if err := store.SaveAssignmentContextSnapshot(ctx, snapshot); err != nil {
		log.Printf("event=assignment_context_capture_failed assignment_id=%q stage=snapshot error_type=%T", assignment.ID, err)
		return
	}
	status := AssignmentContextReadServed
	if snapshot.Truncated {
		status = AssignmentContextReadTruncated
	}
	reads := assignmentContextReadRecords(snapshot, assignmentContextPromptSources(snapshot.Context, comments), AssignmentContextViaPrompt, status, "", now)
	if err := store.AppendAssignmentContextReads(ctx, reads); err != nil {
		log.Printf("event=assignment_context_capture_failed assignment_id=%q stage=reads error_type=%T", assignment.ID, err)
		return
	}
	s.publishAssignmentContextReads(ctx, reads)
}
