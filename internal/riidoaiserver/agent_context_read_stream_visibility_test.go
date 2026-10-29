package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestContextReadStreamFiltersCurrentSources(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		t.Run(mode, func(t *testing.T) {
			s, reader, _, event := contextReadStreamServer(t)
			body := contextReadStreamRequest(s, mode, event).Body.String()
			if !strings.Contains(body, `"id":"task"`) || strings.Contains(body, "hidden-") || strings.Contains(body, "omitted_sources") {
				t.Fatalf("unsafe read event: %s", body)
			}
			if len(reader.requests) != 1 || reader.requests[0].BearerToken != "viewer-token" || !reader.requests[0].Extended {
				t.Fatalf("must use the authenticated stream viewer for current context")
			}
		})
	}
}

func TestContextReadStreamRechecksTaskAccessDuringDelivery(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		t.Run(mode, func(t *testing.T) {
			s, _, auth, first := contextReadStreamServer(t)
			checks := 0
			s.config.Authorizer = authorizationFunc(func(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
				if req.TaskID != "" {
					checks++
					auth.fail = checks > 1
				}
				return auth.Authorize(ctx, token, req)
			})
			second := first
			second.Seq = 2
			body := contextReadStreamRequest(s, mode, first, second).Body.String()
			if strings.Count(body, "event: agent_context_read") != 1 || strings.Contains(body, "id: 2\n") || checks != 2 {
				t.Fatalf("revoked task leaked: checks=%d body=%s", checks, body)
			}
		})
	}
}
