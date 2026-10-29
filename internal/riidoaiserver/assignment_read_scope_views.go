package riidoaiserver

import "context"

func (s Server) assignmentReadScopeView(ctx context.Context, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest) (any, []AssignmentContextSource, bool, error) {
	scope := snapshot.Context.ReadScope
	if req.View == AssignmentContextViewTask || req.View == AssignmentContextViewMeeting || req.View == AssignmentContextViewComments {
		if req.Cursor != "" {
			return nil, nil, false, errAssignmentScopedInlineCursor
		}
		resource, ok := assignmentScopedTarget(*scope, req.View, req.Target)
		if !ok {
			return nil, nil, false, nil
		}
		return s.assignmentScopedInline(ctx, snapshot, resource, req.View)
	}
	if req.Target != "" && !assignmentContextTargetMatches(req.Target, scope.RootID, snapshot.Context.Component.KeyNumber) {
		return nil, nil, false, nil
	}
	resources := []AssignmentScopedResource{}
	sources := []AssignmentContextSource{}
	for _, resource := range scope.Resources {
		if req.View == AssignmentContextViewMeetings && resource.Kind != AssignmentContentMeeting {
			continue
		}
		resources = append(resources, resource)
		sources = append(sources, readScopeSource(resource, "title", ""))
	}
	index := *scope
	index.Resources = resources
	if req.View == AssignmentContextViewBase {
		value := snapshot.Context
		value.ReadScope = &index
		sources = append(sources, assignmentContextTaskSource(value, "body"))
		return value, sources, true, nil
	}
	return index, sources, true, nil
}

func assignmentScopedTarget(scope AssignmentReadScope, view, target string) (AssignmentScopedResource, bool) {
	if target == "" && view != AssignmentContextViewMeeting {
		target = scope.RootID
	}
	var found AssignmentScopedResource
	matches := 0
	for _, resource := range scope.Resources {
		if (resource.Kind == AssignmentContentMeeting) != (view == AssignmentContextViewMeeting) {
			continue
		}
		if assignmentContextTargetMatches(target, resource.ID, resource.KeyNumber) {
			found = resource
			matches++
		}
	}
	return found, matches == 1
}
