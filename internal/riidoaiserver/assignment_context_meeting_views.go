package riidoaiserver

func selectAssignmentContextMeeting(c AIAgentTaskContext, target string) (any, []AssignmentContextSource, bool) {
	if c.Extended == nil {
		return nil, nil, false
	}
	for _, meeting := range c.Extended.RelatedMeetings {
		if meeting.ID != target {
			continue
		}
		readable := meetingHasBody(meeting) && c.Extended.RelatedStatus != "unavailable" && c.Extended.MeetingStatus != "unavailable"
		document := meeting.Document
		if !readable {
			document = nil
		}
		meeting.Document = nil
		var sources []AssignmentContextSource
		if readable {
			sources = []AssignmentContextSource{{Kind: "related_meeting", ID: meeting.ID, Part: "body", Version: assignmentContextVersion(document.Content), UpdatedAt: meeting.UpdatedAt}}
		}
		return struct {
			Meeting  AIAgentTaskContextRelatedMeeting `json:"meeting"`
			Document *AIAgentTaskContextDocument      `json:"document"`
		}{meeting, document}, sources, true
	}
	return nil, nil, false
}

func assignmentContextMeetingList(c AIAgentTaskContext) any {
	meetings := []AIAgentTaskContextRelatedMeeting{}
	total, truncated := 0, false
	status := "unavailable"
	if c.Extended != nil {
		meetings = meetingTitlesOnly(c.Extended.RelatedMeetings)
		total, truncated = max(c.Extended.MeetingTotal, len(meetings)), c.Extended.MeetingTruncated
		status = c.Extended.MeetingStatus
		if status == "" {
			status = c.Extended.RelatedStatus
		}
	}
	return struct {
		Meetings  []AIAgentTaskContextRelatedMeeting `json:"meetings"`
		Status    string                             `json:"status"`
		Total     int                                `json:"total"`
		Truncated bool                               `json:"truncated"`
	}{meetings, status, total, truncated}
}

func meetingTitlesOnly(items []AIAgentTaskContextRelatedMeeting) []AIAgentTaskContextRelatedMeeting {
	out := make([]AIAgentTaskContextRelatedMeeting, len(items))
	copy(out, items)
	for i := range out {
		out[i].Document = nil
	}
	return out
}

func assignmentContextBase(c AIAgentTaskContext) AIAgentTaskContext {
	if c.Extended != nil {
		extended := *c.Extended
		extended.RelatedMeetings = meetingTitlesOnly(extended.RelatedMeetings)
		c.Extended = &extended
	}
	return c
}

func assignmentContextMeetingTitleSources(extended *AIAgentTaskContextExtended) []AssignmentContextSource {
	var out []AssignmentContextSource
	if extended != nil {
		for _, m := range extended.RelatedMeetings {
			out = append(out, AssignmentContextSource{Kind: "related_meeting", ID: m.ID, Part: "title", UpdatedAt: m.UpdatedAt})
		}
	}
	return out
}
