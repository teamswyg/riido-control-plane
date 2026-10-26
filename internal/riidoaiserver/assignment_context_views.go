package riidoaiserver

import "strings"

// selectAssignmentContextView returns the data and sources for one view. ok
// is false when the target is outside this assignment's snapshot.
func selectAssignmentContextView(c AIAgentTaskContext, view, target string) (any, []AssignmentContextSource, bool) {
	target = strings.TrimSpace(target)
	if view == AssignmentContextViewTask {
		return selectAssignmentContextTask(c, target)
	}
	// Other views describe the assigned task only; naming any other work item
	// must not silently return this task's tree or resources.
	if target != "" && !assignmentContextTargetMatches(target, c.Component.ID, c.Component.KeyNumber) {
		return nil, nil, false
	}
	switch view {
	case AssignmentContextViewTree:
		tree := struct {
			Component AIAgentTaskContextComponent `json:"component"`
			Hierarchy AIAgentTaskContextHierarchy `json:"hierarchy"`
		}{c.Component, assignmentContextTitlesOnly(c.Hierarchy)}
		sources := []AssignmentContextSource{assignmentContextTaskSource(c, "title")}
		for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
			sources = append(sources, AssignmentContextSource{Kind: ref.kind, ID: ref.ref.ID, Part: "title", UpdatedAt: normalizedTaskContextTime(ref.ref.UpdatedAt)})
		}
		return tree, sources, true
	case AssignmentContextViewResources:
		sources := assignmentContextResourceSources(c.Resources)
		for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
			sources = append(sources, assignmentContextResourceSources(ref.ref.Resources)...)
		}
		return struct {
			Resources      []AIAgentTaskContextResource `json:"resources"`
			ResourceStatus string                       `json:"resource_status,omitempty"`
			Hierarchy      AIAgentTaskContextHierarchy  `json:"hierarchy"`
		}{c.Resources, c.ResourceStatus, assignmentContextResourcesOnly(c.Hierarchy)}, sources, true
	default:
		return c, assignmentContextSources(c), true
	}
}

func selectAssignmentContextTask(c AIAgentTaskContext, target string) (any, []AssignmentContextSource, bool) {
	if target == "" || assignmentContextTargetMatches(target, c.Component.ID, c.Component.KeyNumber) {
		sources := append([]AssignmentContextSource{assignmentContextTaskSource(c, "body")}, assignmentContextResourceSources(c.Resources)...)
		return struct {
			Component AIAgentTaskContextComponent  `json:"component"`
			Document  AIAgentTaskContextDocument   `json:"document"`
			Resources []AIAgentTaskContextResource `json:"resources,omitempty"`
		}{c.Component, c.Document, c.Resources}, sources, true
	}
	for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
		if assignmentContextTargetMatches(target, ref.ref.ID, ref.ref.KeyNumber) {
			return ref.ref, assignmentContextReferenceSources(ref.kind, ref.ref), true
		}
	}
	return nil, nil, false
}

func assignmentContextTargetMatches(target, id, keyNumber string) bool {
	return target == id || (keyNumber != "" && strings.EqualFold(target, keyNumber))
}
