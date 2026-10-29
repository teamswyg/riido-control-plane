package riidoaiserver

// assignmentContextClassicPromptSources mirrors the prompt written without
// extended context: task body, parent bodies and the links and attachments
// the bounded resource sections actually kept (2,500 / 1,000 runes).
func assignmentContextClassicPromptSources(c AIAgentTaskContext) []AssignmentContextSource {
	sources := append([]AssignmentContextSource{assignmentContextTaskSource(c, "body")}, promptResourceSources(c.Resources, c.ResourceStatus, 2500)...)
	for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
		source := AssignmentContextSource{Kind: ref.kind, ID: ref.ref.ID, Part: "title", UpdatedAt: normalizedTaskContextTime(ref.ref.UpdatedAt)}
		if ref.ref.Document != nil {
			source.Part, source.Version = "body", assignmentContextVersion(ref.ref.Document.Content)
		}
		sources = append(sources, source)
		sources = append(sources, promptResourceSources(ref.ref.Resources, ref.ref.ResourceStatus, 1000)...)
	}
	return append(sources, assignmentContextRelatedSources(c.Extended)...)
}
