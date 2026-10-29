package riidoaiserver

import (
	"context"
	"log"
)

func (s Server) captureAssignmentReadScope(ctx context.Context, snapshot AssignmentContextSnapshot, taskContext *AIAgentTaskContext) {
	if !s.config.AssignmentReadScope || taskContext.readScopeCapture == nil {
		return
	}
	store, ok := s.config.AssignmentContext.(AssignmentContentStore)
	if !ok {
		return
	}
	snapshot.Context = *taskContext
	snapshot.Context.readScopeCapture = nil
	snapshot.Content = taskContext.readScopeCapture.Manifest
	snapshot.Truncated = taskContext.ReadScope.Status != "available"
	for _, part := range snapshot.Content {
		if !part.Complete {
			snapshot.Truncated = true
		}
	}
	if err := store.SaveAssignmentContentBundle(ctx, snapshot, taskContext.readScopeCapture.Pages); err != nil {
		log.Printf("event=assignment_context_capture_failed assignment_id=%q stage=scoped_bundle error_type=%T", snapshot.AssignmentID, err)
		return
	}
	status := AssignmentContextReadServed
	if snapshot.Truncated || snapshot.Context.Document.Status == "partial" {
		status = AssignmentContextReadTruncated
	}
	// The prompt contains only the legacy root/hierarchy projection; the other
	// captured pages become read evidence only when actually served by the CLI.
	reads := assignmentContextReadRecords(snapshot, assignmentContextPromptSources(snapshot.Context, nil), AssignmentContextViaPrompt, status, "", snapshot.CapturedAt)
	if s.config.AssignmentContext.AppendAssignmentContextReads(ctx, reads) != nil {
		return
	}
	s.publishAssignmentContextReads(ctx, reads)
}
