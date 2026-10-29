package riidoaiserver

func readScopeSource(resource AssignmentScopedResource, part, version string) AssignmentContextSource {
	kind := "related_task"
	switch resource.Kind {
	case AssignmentContentComponent:
		// The default related task kind already applies.
	case AssignmentContentBacklog:
		kind = "related_backlog"
	case AssignmentContentIntake:
		kind = "related_intake"
	case AssignmentContentMeeting:
		kind = "related_meeting"
	}
	for _, reason := range resource.Reasons {
		switch reason.Kind {
		case "root":
			kind = "task"
		case "project":
			kind = "project"
		case "milestone":
			kind = "milestone"
		case "parentTask":
			kind = "parent_task"
		}
	}
	return AssignmentContextSource{Kind: kind, ID: resource.ID, Part: part, Version: version, UpdatedAt: normalizedTaskContextTime(resource.UpdatedAt)}
}
