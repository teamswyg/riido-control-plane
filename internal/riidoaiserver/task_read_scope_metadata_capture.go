package riidoaiserver

import "encoding/json"

func (b *taskReadScopeBuilder) captureWorkMetadata(resource AssignmentScopedResource, work taskReadScopeWork) AssignmentContentManifest {
	work.Body, work.DescriptionBody, work.AlternateBody, work.Comments = nil, nil, nil, nil
	data, err := json.Marshal(work)
	if err != nil {
		return b.missingPart(resource, AssignmentContentMetadata, "unavailable")
	}
	version := assignmentContextVersion(string(data))
	sources := []AssignmentContextSource{readScopeSource(resource, "metadata", version)}
	sources = append(sources, readScopeMetadataSources(work)...)
	return b.addPart(resource, AssignmentContentMetadata, version, readScopeJSONPages(data, 0, 0, sources), readScopeMetadataComplete(work.PartsStatus))
}

func readScopeMetadataSources(work taskReadScopeWork) []AssignmentContextSource {
	var sources []AssignmentContextSource
	appendSource := func(raw json.RawMessage, kind string) {
		var value struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &value) == nil && value.ID != "" {
			canonical, err := json.Marshal(raw)
			if err != nil {
				return
			}
			sources = append(sources, AssignmentContextSource{Kind: kind, ID: value.ID, Part: "metadata", Version: assignmentContextVersion(string(canonical))})
		}
	}
	for _, raw := range work.Links {
		appendSource(raw, "link")
	}
	for _, raw := range work.GitHub {
		appendSource(raw, "link")
	}
	appendSource(work.Figma, "link")
	for _, raw := range work.Attachments {
		appendSource(raw, "attachment")
	}
	return sources
}
