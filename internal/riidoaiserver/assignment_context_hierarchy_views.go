package riidoaiserver

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

func assignmentContextTitlesOnly(h AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	for _, ref := range []*AIAgentTaskContextReference{&h.Project, &h.Milestone, &h.ParentTask} {
		ref.Document, ref.DocumentStatus, ref.Resources, ref.ResourceStatus = nil, "", nil, ""
	}
	return h
}

func assignmentContextResourcesOnly(h AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	for _, ref := range []*AIAgentTaskContextReference{&h.Project, &h.Milestone, &h.ParentTask} {
		ref.Document, ref.DocumentStatus = nil, ""
	}
	return h
}

// assignmentContextTaskSource names the assigned task; a body read carries the
// content hash, a title read only the modification time.
func assignmentContextTaskSource(c AIAgentTaskContext, part string) AssignmentContextSource {
	source := AssignmentContextSource{Kind: "task", ID: c.Component.ID, Part: part, UpdatedAt: normalizedTaskContextTime(c.Component.UpdatedAt)}
	if part == "body" {
		source.Version = assignmentContextVersion(c.Document.Content)
	}
	return source
}
