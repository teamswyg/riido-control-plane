package riidoaiserver

import (
	"slices"
	"strings"
)

// splitAssignmentContextComments moves comments out of the snapshot (stored as pages), trigger first.
func splitAssignmentContextComments(c AIAgentTaskContext, trigger assignmentContextTrigger) (AIAgentTaskContext, []AIAgentTaskContextComment) {
	if c.Extended == nil {
		return c, nil
	}
	extended := *c.Extended
	comments := slices.Clone(extended.Comments)
	extended.Comments = nil
	c.Extended = &extended
	if trigger.Kind == AssignmentContextTriggerComment {
		if i := slices.IndexFunc(comments, func(item AIAgentTaskContextComment) bool { return item.ID == trigger.ID }); i >= 0 {
			item := comments[i]
			item.Priority = TaskContextCommentPriorityTrigger
			comments = append([]AIAgentTaskContextComment{item}, slices.Delete(comments, i, i+1)...)
		}
	}
	slices.SortStableFunc(comments, func(a, b AIAgentTaskContextComment) int { return a.Priority - b.Priority })
	return c, comments
}

func pageAssignmentContextComments(comments []AIAgentTaskContextComment) [][]AIAgentTaskContextComment {
	var pages [][]AIAgentTaskContextComment
	for chunk := range slices.Chunk(comments, taskContextCommentPageSize) {
		pages = append(pages, chunk)
	}
	return pages
}

func assignmentContextCommentSources(comments []AIAgentTaskContextComment) []AssignmentContextSource {
	sources := make([]AssignmentContextSource, 0, len(comments))
	for _, comment := range comments {
		sources = append(sources, AssignmentContextSource{Kind: "comment", ID: comment.ID, Part: "body", Version: assignmentContextVersion(comment.Body), UpdatedAt: comment.UpdatedAt})
	}
	return sources
}

func assignmentContextRelatedSources(extended *AIAgentTaskContextExtended) []AssignmentContextSource {
	if extended == nil {
		return nil
	}
	var sources []AssignmentContextSource
	for _, work := range extended.RelatedWork {
		sources = append(sources, AssignmentContextSource{Kind: "related_" + strings.TrimSpace(work.Kind), ID: work.ID, Part: "title"})
	}
	for _, meeting := range extended.RelatedMeetings {
		sources = append(sources, AssignmentContextSource{Kind: "related_meeting", ID: meeting.ID, Part: "title"})
	}
	return sources
}

// assignmentContextPromptSources lists what the prompt contained (CLI mode:
// parent summaries, related titles, priority-1 comments only).
func assignmentContextPromptSources(c AIAgentTaskContext, comments []AIAgentTaskContextComment) []AssignmentContextSource {
	if c.Extended == nil {
		return assignmentContextSources(c)
	}
	sources := append([]AssignmentContextSource{assignmentContextTaskSource(c, "body")}, assignmentContextResourceSources(c.Resources)...)
	for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
		source := AssignmentContextSource{Kind: ref.kind, ID: ref.ref.ID, Part: "title", UpdatedAt: normalizedTaskContextTime(ref.ref.UpdatedAt)}
		if ref.ref.Document != nil && ref.ref.Document.Content != "" {
			source.Part, source.Version = "summary", assignmentContextVersion(ref.ref.Document.Content)
		}
		sources = append(sources, source)
	}
	sources = append(sources, assignmentContextRelatedSources(c.Extended)...)
	priority := slices.DeleteFunc(slices.Clone(comments), func(item AIAgentTaskContextComment) bool { return item.Priority != TaskContextCommentPriorityTrigger })
	return append(sources, assignmentContextCommentSources(priority)...)
}
