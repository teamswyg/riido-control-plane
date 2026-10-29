package riidoaiserver

type scopedCurrentSource struct {
	source   AssignmentContextSource
	title    string
	complete bool
}

func assignmentReadScopeCurrentSources(current AIAgentTaskContext) []scopedCurrentSource {
	if current.ReadScope == nil {
		return nil
	}
	var out []scopedCurrentSource
	titles := map[string]string{}
	for _, resource := range current.ReadScope.Resources {
		titles[string(resource.Kind)+"\x00"+resource.ID] = resource.Title
		out = append(out, scopedCurrentSource{source: readScopeSource(resource, "title", ""), title: resource.Title})
	}
	if current.readScopeCapture == nil {
		return out
	}
	complete := map[string]bool{}
	for _, part := range current.readScopeCapture.Manifest {
		for _, cursor := range part.Cursors {
			complete[cursor] = part.Complete
		}
	}
	for _, page := range current.readScopeCapture.Pages {
		for _, source := range page.Sources {
			title := ""
			if source.ID == page.SourceID {
				title = titles[string(page.ResourceKind)+"\x00"+page.SourceID]
			}
			out = append(out, scopedCurrentSource{source: source, title: title, complete: complete[page.Cursor]})
		}
	}
	return out
}

func addReadScopeCurrentMetadata(metadata map[runReadSourceKey]runReadMetadata, current AIAgentTaskContext) {
	for _, item := range assignmentReadScopeCurrentSources(current) {
		key := runReadSourceKey{item.source.Kind, item.source.ID, item.source.Part}
		if _, ok := metadata[key]; !ok {
			continue
		}
		previous := metadata[key]
		versions := previous.versions
		if versions == nil {
			versions = map[string]bool{}
		}
		if previous.comparable && previous.source.Version != "" {
			versions[previous.source.Version] = true
		}
		if item.complete && item.source.Version != "" {
			versions[item.source.Version] = true
		}
		addRunReadMetadata(metadata, item.source, item.title, item.complete)
		value := metadata[key]
		value.versions = versions
		metadata[key] = value
	}
}
