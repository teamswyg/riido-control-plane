package riidoaiserver

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
