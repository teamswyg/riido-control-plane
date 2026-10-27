package riidoaiserver

import "strings"

// GET-only metadata is derived from the same current, viewer-authorized context
// as visibility. It never loads retained snapshots or triggers a second read.
type runReadMetadata struct {
	title      string
	source     AssignmentContextSource
	comparable bool
}

func currentRunReadMetadata(current AIAgentTaskContext) map[runReadSourceKey]runReadMetadata {
	current = runReadAvailableContext(current)
	metadata := map[runReadSourceKey]runReadMetadata{}
	for key := range visibleRunReadSources(current) {
		metadata[key] = runReadMetadata{}
	}
	addRunReadMetadata(metadata, assignmentContextTaskSource(current, "title"), current.Component.Title, false)
	addRunReadMetadata(metadata, assignmentContextTaskSource(current, "body"), current.Component.Title, current.Document.Status != "partial")
	addRunReadResourceMetadata(metadata, current.Resources)
	for _, ref := range assignmentContextHierarchyRefs(current.Hierarchy) {
		canCompare := ref.ref.Document != nil && ref.ref.Document.Status != "partial" && ref.ref.DocumentStatus != "partial"
		sources := assignmentContextReferenceSources(ref.kind, ref.ref)
		addRunReadMetadata(metadata, sources[0], ref.ref.Title, canCompare)
		addRunReadResourceMetadata(metadata, ref.ref.Resources)
	}
	addRunReadExtendedMetadata(metadata, current)
	return metadata
}

func addRunReadMetadata(metadata map[runReadSourceKey]runReadMetadata, source AssignmentContextSource, title string, canCompare bool) {
	key := runReadSourceKey{source.Kind, source.ID, source.Part}
	if _, visible := metadata[key]; !visible {
		return
	}
	value := runReadMetadata{title: strings.TrimSpace(title), source: source, comparable: canCompare}
	metadata[key] = value
	if source.Part == "body" {
		// Prompt summaries record the hash of the same source document, not the
		// rendered summary. Title-only reads do not record a body hash.
		key.part = "summary"
		metadata[key] = value
		key.part = "title"
		value.source.Version = ""
		value.comparable = false
		metadata[key] = value
	}
}

func addRunReadResourceMetadata(metadata map[runReadSourceKey]runReadMetadata, resources []AIAgentTaskContextResource) {
	sources := assignmentContextResourceSources(resources)
	for i, source := range sources {
		addRunReadMetadata(metadata, source, resources[i].Title, false)
	}
}
