package riidoaiserver

import "strings"

// Views mirror the daemon CLI reads: base = `riido context get`,
// task = `riido work get [target]`, tree = `riido work tree`,
// resources = `riido work relations`. Comments, activity and search are not
// part of the snapshot and stay unsupported until the read scope is decided.
const (
	AssignmentContextViewBase      = "base"
	AssignmentContextViewTask      = "task"
	AssignmentContextViewTree      = "tree"
	AssignmentContextViewResources = "resources"
)

func validAssignmentContextView(view string) bool {
	switch view {
	case AssignmentContextViewBase, AssignmentContextViewTask, AssignmentContextViewTree, AssignmentContextViewResources:
		return true
	}
	return false
}

// selectAssignmentContextView returns the data and sources for one view. ok
// is false when the target is outside this assignment's snapshot.
func selectAssignmentContextView(c AIAgentTaskContext, view, target string) (any, []AssignmentContextSource, bool) {
	switch view {
	case AssignmentContextViewTask:
		return selectAssignmentContextTask(c, strings.TrimSpace(target))
	case AssignmentContextViewTree:
		tree := struct {
			Component AIAgentTaskContextComponent `json:"component"`
			Hierarchy AIAgentTaskContextHierarchy `json:"hierarchy"`
		}{c.Component, assignmentContextTitlesOnly(c.Hierarchy)}
		sources := []AssignmentContextSource{{Kind: "task", ID: c.Component.ID, Part: "title"}}
		for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
			sources = append(sources, AssignmentContextSource{Kind: ref.kind, ID: ref.ref.ID, Part: "title"})
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
		sources := append([]AssignmentContextSource{{Kind: "task", ID: c.Component.ID, Part: "body", Version: assignmentContextVersion(c.Document.Content)}}, assignmentContextResourceSources(c.Resources)...)
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
