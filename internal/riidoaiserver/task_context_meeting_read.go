package riidoaiserver

import (
	"context"
)

const taskContextMeetingQuery = `query AgentMeetingBody($id: String!, $team: String!) {
 meeting: getMeeting(meetingId:$id,teamId:$team) { id title updatedAt deletedAt description document { id tiptapDocumentId content } }
}`

func (c *AIAgentPrivateTaskContextClient) readMeetingBody(ctx context.Context, team, bearer string, meeting AIAgentTaskContextRelatedMeeting) AIAgentTaskContextRelatedMeeting {
	meeting.Document, meeting.BodyStatus = nil, "unavailable"
	detail, err := c.readMeetingDetail(ctx, team, meeting.ID, bearer)
	if err != nil {
		meeting.BodyStatus = meetingReadFailure(ctx, err)
		return meeting
	}
	if detail.ID != meeting.ID {
		return meeting
	}
	if detail.DeletedAt != nil {
		meeting.BodyStatus = "deleted"
		return meeting
	}
	meeting.UpdatedAt = normalizedTaskContextTime(detail.UpdatedAt)
	document := AIAgentTaskContextDocument{Content: detail.Description, ContentFormat: "html", Source: "meeting_snapshot", Status: "available"}
	if detail.Document != nil {
		document.ID, document.TiptapDocumentID = detail.Document.ID, detail.Document.TiptapDocumentID
		document.Content = detail.Document.Content
		if document.Content == "" {
			document.Content = detail.Document.HTMLContent
		}
		projection, err := c.getLatestProviderDocument(ctx, team, meeting.ID, bearer)
		if err != nil {
			meeting.BodyStatus = meetingReadFailure(ctx, err)
			return meeting
		}
		document = selectTaskContextDocument(document, projection)
		if document.Source == "component_snapshot" {
			document.Source = "meeting_snapshot"
		}
	}
	document = normalizeTaskContextAlternate(document, 1<<20)
	document.ContentFormat = "markdown"
	if document.AlternateContent != "" {
		document.AlternateFormat = "markdown"
	}
	if document.Status == "available" && document.Content == "" && document.AlternateContent == "" {
		document.Status = "empty"
	}
	document.Content, document.AlternateContent, meeting.Truncated = boundMeetingText(document.Content, document.AlternateContent)
	if meeting.Truncated {
		document.Status = "partial"
	}
	meeting.Document, meeting.BodyStatus = &document, document.Status
	return meeting
}
