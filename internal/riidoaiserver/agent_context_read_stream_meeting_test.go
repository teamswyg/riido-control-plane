package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestContextReadStreamRechecksMeetingBodyAccess(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		for _, status := range []string{"denied", "unavailable"} {
			t.Run(mode+"/"+status, func(t *testing.T) {
				s, reader, auth, event := contextReadStreamServer(t)
				reader.value.Extended = meetingContextFixture(t, "available").Extended
				payload := event.Payload.(AgentContextReadEvent)
				payload.Sources = []AgentContextReadSource{
					{Kind: "related_meeting", ID: "meeting-1", Part: "title", Status: "served"},
					{Kind: "related_meeting", ID: "meeting-1", Part: "body", Version: "meeting-body-hash", Status: "served"},
				}
				event.Payload = payload
				checks := 0
				s.config.Authorizer = authorizationFunc(func(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
					if req.TaskID != "" {
						checks++
						if checks > 1 {
							reader.value.Extended = meetingContextFixture(t, status).Extended
						}
					}
					return auth.Authorize(ctx, token, req)
				})
				second := event
				second.Seq = 2
				body := contextReadStreamRequest(s, mode, event, second).Body.String()
				if strings.Count(body, "event: agent_context_read") != 2 || strings.Count(body, "meeting-body-hash") != 1 || strings.Count(body, `"part":"title"`) != 2 {
					t.Fatalf("meeting access was not rechecked: %s", body)
				}
			})
		}
	}
}
