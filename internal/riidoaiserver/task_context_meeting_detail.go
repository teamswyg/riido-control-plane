package riidoaiserver

import (
	"context"
	"net/url"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) readMeetingDetail(ctx context.Context, team, id, bearer string) (privateMeetingDetail, error) {
	if c.projectionBundles {
		var out struct {
			Meeting privateMeetingDetail `json:"meeting"`
		}
		err := c.contextGraphQL(ctx, bearer, taskContextMeetingQuery, team, id, &out)
		return out.Meeting, err
	}
	var out privateMeetingDetail
	endpoint := strings.TrimRight(c.baseURL, "/") + "/teams/" + url.PathEscape(team) + "/meetings/" + url.PathEscape(id)
	err := c.readPrivateContextComponent(ctx, endpoint, bearer, "related_meeting_body", &out)
	return out, err
}

func meetingReadFailure(ctx context.Context, err error) string {
	switch taskContextReadOutcome(ctx, err) {
	case "denied":
		return "denied"
	case "missing":
		return "deleted"
	default:
		return "unavailable"
	}
}
