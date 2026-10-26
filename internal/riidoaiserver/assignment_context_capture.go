package riidoaiserver

import (
	"context"
	"log"
	"strings"
	"time"
)

type assignmentContextTrigger struct {
	Kind string
	ID   string
}

// captureAssignmentContext stores the prompt's task context and records its
// sources as read via the prompt. Best effort: failures are logged and the CLI
// later reports the snapshot as unavailable.
func (s Server) captureAssignmentContext(ctx context.Context, assignment Assignment, taskContext *AIAgentTaskContext, trigger assignmentContextTrigger) {
	store := s.config.AssignmentContext
	if store == nil || taskContext == nil || strings.TrimSpace(assignment.ID) == "" {
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

func assignmentContextReadRecords(snapshot AssignmentContextSnapshot, sources []AssignmentContextSource, via, status, requestID string, at time.Time) []AssignmentContextRead {
	reads := make([]AssignmentContextRead, 0, len(sources))
	for _, source := range sources {
		reads = append(reads, AssignmentContextRead{
			AssignmentID:         snapshot.AssignmentID,
			TaskID:               snapshot.TaskID,
			AgentID:              snapshot.AgentID,
			RequesterPrincipalID: snapshot.RequesterPrincipalID,
			TriggerKind:          snapshot.TriggerKind,
			TriggerID:            snapshot.TriggerID,
			SourceKind:           source.Kind,
			SourceID:             source.ID,
			SourcePart:           source.Part,
			SourceVersion:        source.Version,
			SourceUpdatedAt:      source.UpdatedAt,
			Via:                  via,
			Status:               status,
			ServedAt:             at,
			RequestID:            requestID,
		})
	}
	return reads
}
