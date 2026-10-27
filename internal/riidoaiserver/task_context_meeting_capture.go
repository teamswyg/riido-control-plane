package riidoaiserver

import (
	"context"
	"strings"
	"sync"
)

func (c *AIAgentPrivateTaskContextClient) captureRelatedMeetings(ctx context.Context, team, bearer string, items []privateRelatedMeeting, out *AIAgentTaskContextExtended) {
	out.MeetingStatus = "available"
	seen := map[string]bool{}
	for _, item := range items {
		if item.Meeting == nil {
			continue
		}
		id := strings.TrimSpace(item.Meeting.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out.MeetingTotal++
		if len(out.RelatedMeetings) >= taskContextMaxMeetings {
			out.MeetingTruncated = true
			continue
		}
		out.RelatedMeetings = append(out.RelatedMeetings, AIAgentTaskContextRelatedMeeting{ID: id, Title: boundedTaskContextText(item.Meeting.Title, 200), StartTime: normalizedTaskContextTime(item.Meeting.StartTime), BodyStatus: "unavailable"})
	}
	var wg sync.WaitGroup
	slots := make(chan struct{}, taskContextMeetingConcurrency)
	for i := range out.RelatedMeetings {
		wg.Go(func() {
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-slots }()
			out.RelatedMeetings[i] = c.readMeetingBody(ctx, team, bearer, out.RelatedMeetings[i])
		})
	}
	wg.Wait()
}

func boundMeetingText(content, alternate string) (string, string, bool) {
	if len(content)+len(alternate) <= taskContextMeetingBodyBytes {
		return content, alternate, false
	}
	if alternate != "" {
		content = truncateUTF8(content, taskContextMeetingBodyBytes*2/3)
		alternate = truncateUTF8(alternate, taskContextMeetingBodyBytes-len(content))
	} else {
		content = truncateUTF8(content, taskContextMeetingBodyBytes)
	}
	return content, alternate, true
}
