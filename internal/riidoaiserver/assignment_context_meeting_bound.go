package riidoaiserver

func cloneAssignmentContextMeetings(extended *AIAgentTaskContextExtended) *AIAgentTaskContextExtended {
	if extended == nil {
		return nil
	}
	out := *extended
	out.RelatedMeetings = append([]AIAgentTaskContextRelatedMeeting(nil), extended.RelatedMeetings...)
	for i := range out.RelatedMeetings {
		if document := out.RelatedMeetings[i].Document; document != nil {
			documentCopy := *document
			out.RelatedMeetings[i].Document = &documentCopy
		}
	}
	return &out
}

func assignmentContextMeetingDocuments(extended *AIAgentTaskContextExtended) []*AIAgentTaskContextDocument {
	var out []*AIAgentTaskContextDocument
	if extended != nil {
		for _, meeting := range extended.RelatedMeetings {
			if meeting.Document != nil {
				out = append(out, meeting.Document)
			}
		}
	}
	return out
}

// A global snapshot cut is conservative: no surviving meeting body claims full content.
func markTruncatedMeetingDocuments(extended *AIAgentTaskContextExtended) {
	if extended == nil {
		return
	}
	for i := range extended.RelatedMeetings {
		meeting := &extended.RelatedMeetings[i]
		if meetingHasBody(*meeting) && meeting.BodyStatus != "empty" {
			meeting.Truncated = true
			meeting.BodyStatus = "partial"
			meeting.Document.Status = "partial"
		}
	}
}
