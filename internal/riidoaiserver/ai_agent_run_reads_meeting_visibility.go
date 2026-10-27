package riidoaiserver

// Only an authorized, current body read can expose historical meeting body metadata.
func visibleMeetingBodySources(extended *AIAgentTaskContextExtended) []AssignmentContextSource {
	var sources []AssignmentContextSource
	if extended == nil || extended.MeetingStatus == "unavailable" {
		return sources
	}
	for _, meeting := range extended.RelatedMeetings {
		if meetingHasBody(meeting) {
			sources = append(sources, AssignmentContextSource{Kind: "related_meeting", ID: meeting.ID, Part: "body"})
		}
	}
	return sources
}
