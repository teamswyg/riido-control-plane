package riidoaiserver

func addRunReadExtendedMetadata(metadata map[runReadSourceKey]runReadMetadata, current AIAgentTaskContext) {
	extended := current.Extended
	if extended == nil {
		return
	}
	// Related work has titles only: no version is invented from display text.
	for i, source := range assignmentContextRelatedSources(extended) {
		if i < len(extended.RelatedWork) {
			addRunReadMetadata(metadata, source, extended.RelatedWork[i].Title, false)
		}
	}
	for i, source := range assignmentContextMeetingTitleSources(extended) {
		meeting := extended.RelatedMeetings[i]
		addRunReadMetadata(metadata, source, meeting.Title, false)
		_, bodySources, _ := selectAssignmentContextMeeting(current, meeting.ID)
		for _, bodySource := range bodySources {
			canCompare := !meeting.Truncated && meeting.Document != nil && meeting.Document.Status != "partial" && meeting.BodyStatus != "partial"
			addRunReadMetadata(metadata, bodySource, meeting.Title, canCompare)
		}
	}
	for i, source := range assignmentContextCommentSources(extended.Comments) {
		addRunReadMetadata(metadata, source, "", !extended.Comments[i].Truncated)
	}
}
