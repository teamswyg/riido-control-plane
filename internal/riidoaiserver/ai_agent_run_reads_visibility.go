package riidoaiserver

import (
	"errors"
	"net/http"
)

type runReadSourceKey struct{ kind, id, part string }

func (s Server) currentRunReadSources(w http.ResponseWriter, r *http.Request, workspaceID, taskID, assignmentID string) (map[runReadSourceKey]runReadMetadata, bool) {
	token, _ := requestBearerToken(r)
	current, err := s.loadCurrentRunReadContext(r.Context(), token, workspaceID, taskID, assignmentID)
	if errors.Is(err, ErrAuthorizationForbidden) {
		writeError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "source access unavailable")
		return nil, false
	}
	return currentRunReadMetadata(current), true
}

// Visibility is recomputed from the viewer's current response. Missing parents,
// resources, old comments or relation reads never fall back to captured content.
func visibleRunReadSources(current AIAgentTaskContext) map[runReadSourceKey]bool {
	current = runReadAvailableContext(current)
	visible := map[runReadSourceKey]bool{}
	add := func(source AssignmentContextSource) {
		if source.ID == "" {
			return
		}
		visible[runReadSourceKey{source.Kind, source.ID, source.Part}] = true
		if source.Part == "body" {
			visible[runReadSourceKey{source.Kind, source.ID, "summary"}] = true
			visible[runReadSourceKey{source.Kind, source.ID, "title"}] = true
		}
	}
	add(assignmentContextTaskSource(current, "title"))
	if current.Document.Status != "unavailable" {
		add(assignmentContextTaskSource(current, "body"))
	}
	for _, source := range assignmentContextResourceSources(current.Resources) {
		add(source)
	}
	for _, ref := range assignmentContextHierarchyRefs(current.Hierarchy) {
		for _, source := range assignmentContextReferenceSources(ref.kind, ref.ref) {
			add(source)
		}
	}
	for _, source := range visibleMeetingBodySources(current.Extended) {
		add(source)
	}
	for _, source := range assignmentContextRelatedSources(current.Extended) {
		add(source)
	}
	if current.Extended != nil {
		for _, source := range assignmentContextCommentSources(current.Extended.Comments) {
			add(source)
		}
	}
	for _, item := range assignmentReadScopeCurrentSources(current) {
		add(item.source)
	}
	return visible
}
