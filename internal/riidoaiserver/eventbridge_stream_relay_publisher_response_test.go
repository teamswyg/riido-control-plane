package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEventBridgeStreamRelayPublisherRejectsBadResponses(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"malformed", `{`, "decode EventBridge"},
		{"entry-error", `{"FailedEntryCount":0,"Entries":[{"ErrorCode":"ThrottlingException","ErrorMessage":"slow"}]}`, "ThrottlingException"},
		{"failed-count", `{"FailedEntryCount":2,"Entries":[]}`, "failed_entry_count=2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			publisher := newEventBridgeBoundaryPublisher(t, server.URL)
			defer publisher.Close()
			event := streamRelayEventForEventBridgeTest(time.Date(2026, 5, 26, 5, 6, 7, 0, time.UTC))
			err := publisher.PublishStreamRelayEvent(context.Background(), event)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("PublishStreamRelayEvent error = %v, want %q", err, tc.want)
			}
		})
	}
}
