package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) assignmentContextChangesView(ctx context.Context, current AssignmentContextSnapshot, req AgentAssignmentContextRequest) (any, []AssignmentContextSource, bool, error) {
	baseline, err := s.assignmentContextBaseline(ctx, current, strings.TrimSpace(req.Since))
	if err != nil {
		return nil, nil, false, err
	}
	if current.Context.ReadScope == nil || baseline.Context.ReadScope == nil {
		return nil, nil, false, errAssignmentContextBaseline
	}
	if err := validateAssignmentReadScopeManifest(baseline); err != nil {
		return nil, nil, false, err
	}
	out := compareAssignmentContextParts(baseline, current)
	return out, assignmentContextChangeSources(current, out.Changes), true, nil
}

func assignmentContextChangeSources(current AssignmentContextSnapshot, changes []AssignmentContextChange) []AssignmentContextSource {
	sources := []AssignmentContextSource{}
	seen := map[string]bool{}
	for _, change := range changes {
		if change.Current != nil {
			seen[string(change.ResourceKind)+"\x00"+change.SourceID] = true
		}
	}
	for _, resource := range current.Context.ReadScope.Resources {
		if !seen[string(resource.Kind)+"\x00"+resource.ID] {
			continue
		}
		// The caller receives titles/digests/cursors, never the referenced bodies.
		sources = append(sources, readScopeSource(resource, "title", ""))
	}
	return sources
}
