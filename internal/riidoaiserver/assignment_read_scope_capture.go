package riidoaiserver

import (
	"context"
	"log"
	"slices"
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
	_, comments := splitAssignmentContextComments(*taskContext, assignmentContextTrigger{Kind: snapshot.TriggerKind, ID: snapshot.TriggerID})
	reads := assignmentContextReadRecords(snapshot, readScopePromptSources(snapshot.Context, comments), AssignmentContextViaPrompt, status, "", snapshot.CapturedAt)
	if s.config.AssignmentContext.AppendAssignmentContextReads(ctx, reads) != nil {
		return
	}
	s.publishAssignmentContextReads(ctx, reads)
}

// readScopePromptSources lists only what the read-scope prompt carries: the
// task document preview, the hierarchy titles and the comment that started
// the run. Links, related items and parent bodies are read through the CLI.
func readScopePromptSources(c AIAgentTaskContext, comments []AIAgentTaskContextComment) []AssignmentContextSource {
	sources := []AssignmentContextSource{assignmentContextTaskSource(c, "body")}
	for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
		sources = append(sources, AssignmentContextSource{Kind: ref.kind, ID: ref.ref.ID, Part: "title", UpdatedAt: normalizedTaskContextTime(ref.ref.UpdatedAt)})
	}
	triggers := slices.DeleteFunc(slices.Clone(comments), func(item AIAgentTaskContextComment) bool { return item.Priority != TaskContextCommentPriorityTrigger })
	return append(sources, assignmentContextCommentSources(triggers)...)
}
