package riidoaiserver

import (
	"net/http"
	"strings"
	"time"
)

// Refusals are recorded without claiming the meeting body was served.
func (s Server) refuseUnavailableMeeting(w http.ResponseWriter, r *http.Request, snapshot AssignmentContextSnapshot, req AgentAssignmentContextRequest, now time.Time) bool {
	if req.View != AssignmentContextViewMeeting || snapshot.Context.Extended == nil {
		return false
	}
	for _, meeting := range snapshot.Context.Extended.RelatedMeetings {
		if meeting.ID != strings.TrimSpace(req.Target) {
			continue
		}
		if meetingHasBody(meeting) && snapshot.Context.Extended.RelatedStatus != "unavailable" && snapshot.Context.Extended.MeetingStatus != "unavailable" {
			return false
		}
		status, code, message := AssignmentContextReadUnavailable, http.StatusServiceUnavailable, "meeting body unavailable"
		if meeting.BodyStatus == "denied" {
			status, code, message = AssignmentContextReadDenied, http.StatusForbidden, "meeting body forbidden"
		}
		if meeting.BodyStatus == "deleted" {
			code = http.StatusNotFound
		}
		s.writeAssignmentContextRefusal(w, r, snapshot, req, now, status, code, message)
		return true
	}
	return false
}

func assignmentContextMeetingTruncated(c AIAgentTaskContext, req AgentAssignmentContextRequest) bool {
	if c.Extended == nil {
		return false
	}
	if req.View == AssignmentContextViewMeetings {
		return c.Extended.MeetingTruncated
	}
	if req.View == AssignmentContextViewMeeting {
		for _, m := range c.Extended.RelatedMeetings {
			if m.ID == strings.TrimSpace(req.Target) {
				return m.Truncated || m.BodyStatus == "partial"
			}
		}
	}
	return false
}
