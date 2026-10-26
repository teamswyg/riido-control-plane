package riidoaiserver

import "strings"

// selectAssignmentContextView returns the data and sources for one view. ok
// is false when the target is outside this assignment's snapshot.
func selectAssignmentContextView(c AIAgentTaskContext, view, target string) (any, []AssignmentContextSource, bool) {
	target = strings.TrimSpace(target)
	if view == AssignmentContextViewMeeting {
		return selectAssignmentContextMeeting(c, target)
	}
	if view == AssignmentContextViewTask {
		return selectAssignmentContextTask(c, target)
	}
	// Other views describe the assigned task only; naming any other work item
	// must not silently return this task's tree or resources.
	if target != "" && !assignmentContextTargetMatches(target, c.Component.ID, c.Component.KeyNumber) {
		return nil, nil, false
	}
	switch view {
	case AssignmentContextViewMeetings:
		return assignmentContextMeetingList(c), assignmentContextMeetingTitleSources(c.Extended), true
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
			Related        *AIAgentTaskContextExtended  `json:"related,omitempty"`
		}{c.Resources, c.ResourceStatus, assignmentContextResourcesOnly(c.Hierarchy), assignmentContextRelatedOnly(c.Extended)}, append(sources, assignmentContextRelatedSources(c.Extended)...), true
	default:
		return assignmentContextBase(c), assignmentContextSources(c), true
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

// assignmentContextTargetMatches accepts the ID, the stored key, or the key as
// the prompt shows it ("RIID-" prefixed), so an agent can use what it was shown.
func assignmentContextTargetMatches(target, id, keyNumber string) bool {
	if target == id {
		return true
	}
	keyNumber = strings.TrimSpace(keyNumber)
	return keyNumber != "" && (strings.EqualFold(target, keyNumber) || strings.EqualFold(target, formatTaskContextKeyNumber(keyNumber)))
}
