package riidoaiserver

func (b *taskReadScopeBuilder) captureIncompleteTypedMeeting(ref taskReadScopeMeetingRef, meeting *taskReadMeeting) AssignmentScopedResource {
	resource := AssignmentScopedResource{Kind: AssignmentContentMeeting, ID: ref.ID, TeamID: ref.TeamID, Reasons: ref.Reasons}
	if meeting.Summary != nil && meeting.Summary.ID == ref.ID && meeting.Summary.Kind == AssignmentContentMeeting {
		resource.Title = meeting.Summary.Title
		resource.UpdatedAt = meeting.Summary.UpdatedAt
	}
	if meeting.ContentPage != nil {
		resource.Parts = append(resource.Parts, b.captureTypedPrefix(resource, AssignmentContentMetadata, meeting.ContentPage))
	} else {
		raw, err := readScopeMeetingMetadata(*meeting)
		if err != nil {
			resource.Parts = append(resource.Parts, b.missingPart(resource, AssignmentContentMetadata, "unavailable"))
		} else {
			version := assignmentContextVersion(string(raw))
			resource.Parts = append(resource.Parts, b.addPart(resource, AssignmentContentMetadata, version, readScopeJSONPages(raw, 0, 0, []AssignmentContextSource{readScopeSource(resource, "metadata", version)}), readScopeMetadataComplete(meeting.PartsStatus)))
		}
	}
	for _, body := range []*taskReadBodyPage{meeting.Body, meeting.DescriptionBody, meeting.AlternateBody} {
		if body == nil {
			continue
		}
		resource.Parts = append(resource.Parts, b.captureTypedPrefix(resource, readScopeBodyPart(body.Part), body))
	}
	if meeting.Board != nil && meeting.Board.ContentPage != nil {
		resource.Parts = append(resource.Parts, b.captureTypedPrefix(resource, AssignmentContentBoard, meeting.Board.ContentPage))
	} else {
		resource.Parts = append(resource.Parts, b.missingPart(resource, AssignmentContentBoard, "unavailable"))
	}
	return resource
}
